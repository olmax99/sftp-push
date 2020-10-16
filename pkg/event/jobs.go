package event

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/fsnotify/fsnotify"
)

//!+stage-3

// Remove waits for S3 upload to finish and removes event file
func (o FsEventOps) removeF(e EventInfo) error {
	if err := os.Remove(e.Event.AbsLoc); err != nil {
		return err
	}
	return nil
}

// PushS3 uploads the source event file byte stream to S3 and removes the file
func (o FsEventOps) pushS3(done <-chan struct{}, in io.Reader, pi EventPushInfo, ei EventInfo) <-chan *ResultInfo {
	out := make(chan *ResultInfo)
	go func() {
		defer close(out)
		uploader := s3manager.NewUploaderWithClient(pi.Session, func(u *s3manager.Uploader) {
			u.PartSize = 64 * 1024 * 1024 // 64MB per part
		})

		// Create a context with a timeout that will abort the upload if it takes
		// more than the passed in timeout.
		ctx := context.Background()
		var cancelFn func()
		ctx, cancelFn = context.WithTimeout(ctx, time.Duration(3*time.Second))
		// Ensure the context is canceled to prevent leaking.
		// See context package for more information, https://golang.org/pkg/context/
		defer cancelFn()

		// Uploads the object to S3. The Context will interrupt the request if the
		// timeout expires.
		r, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
			Body:   in,
			Bucket: pi.Bucket,
			Key:    &pi.Key,
		})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
				// If the SDK can determine the request or retry delay was canceled
				// by a context the CanceledErrorCode error code will be returned.
				log.Printf("request's context canceled, %s", err)
			} else {
				log.Printf("WARNING[-] Stage-3: PushS3 %s", err)
			}
		} else {
			select {
			case out <- &ResultInfo{response: r, eventInfo: ei}:
			case <-done:
				return
			}
			if err := o.removeF(ei); err != nil {
				log.Printf("Error[-] Stage-3: removeF %s", err)
			}
		}
	}()
	return out
}

//!-stage-3

//!+stage-2

// FType detects and returns the file type along with the initial file io.Reader
func (o *FsEventOps) fType(epath string) (string, *io.Reader) {
	f, err := os.Open(epath)
	if err != nil {
		log.Printf("WARNING[-] Stage-2: Open %s, %s\n", filepath.Base(epath), err)
	}

	buf := make([]byte, 32)
	if _, err := f.Read(buf); err != nil {
		log.Printf("ERROR[-] Stage-2: File Read %s, %s\n", filepath.Base(epath), err)
	}
	fT := http.DetectContentType(buf)

	// glue those bytes back onto the reader
	r := io.MultiReader(bytes.NewReader(buf), f)

	return fT, &r
}

// controlWorkers detects the file type and sends the decompressed byte stream to the PushS3 stage
func (o *FsEventOps) controlWorkers(in <-chan EventInfo, pi *EventPushInfo) {
	var (
		wg  sync.WaitGroup
		gz  io.Reader
		err error
	)
	done := make(chan struct{})
	defer close(done)
	for {
		select {
		case e := <-in:
			p := e.Event.AbsLoc
			ft, b := o.fType(p)
			switch ft {
			case "application/x-gzip":
				log.Printf("DEBUG[*] Stage-2: fT %s, %s\n", ft, filepath.Base(p))
				gz, err = gzip.NewReader(*b)
				if err != nil {
					log.Printf("ERROR[-] Stage-2: gzip.NewReader, %s\n", err)
					done <- struct{}{}
				}
			case "application/zip":
				log.Printf("DEBUG[*] Stage-2: fT %s, %s\n", ft, filepath.Base(p))
			default:
				// if strings.HasPrefix(string(buf), "\x42\x5a\x68") {
				// 	log.Printf("INFO[*] Stage-1: file type %s, %s\n", ft, filepath.Base(p))
				// } else {}
				log.Printf("WARNING[-] Stage-2: unexpected fT %s, %s\n", ft, filepath.Base(p))
			}
			pi.Key, err = o.reduceEventPath(p, pi.Userpath)
			if err != nil {
				log.Printf("ERROR[*] Stage-2: %s", err)
				done <- struct{}{}
			}
			wg.Add(1) // only single result in PushS3 chan
			for n := range o.pushS3(done, gz, *pi, e) {
				pi.Results <- n
			}
			go func() {
				wg.Wait()
			}()
		}
	}
}

//!-stage-2

//!+stage-1

// Listen listens to file events from fsnotify.Watcher and sends them to the stage-1 channel
func (o *FsEventOps) listen(w *fsnotify.Watcher, out chan<- EventInfo) {
	// out := make(chan EventInfo)
	// go func() { for { select { .. }} close(out) }()
	// return out

	// TODO There can only be one EventInfo per event <- use buffered
	for {
		select {
		case event := <-w.Events: // RECEIVE event
			// all events are logged by default
			log.Printf("DEBUG[+] Stage-1: %v, eventT: %T\n", event, event)

			if event.Op&fsnotify.CloseWrite == fsnotify.CloseWrite {
				fsEv := &FsEvent{
					Event: event,
					Ops:   &FsEventOps{},
				}
				ev, err := fsEv.Info()
				if err != nil {
					log.Printf("WARNING[-] Stage-1: Listen %s\n", err)
				}

				// TODO work with cases:
				// see github.com/olmax99/fsnotify@v1.5.0/inotify.go (l 178)
				// M-. NewWatcher + M-. ReadEvents

				// 32 bytes needed for determining file type
				if ev.Meta.Size >= int64(32) {
					out <- *ev // SEND needs no close as infinite amount of Events
				} else {
					// only for testing
					einfo, err := json.Marshal(ev)
					if err != nil {
						log.Printf("ERROR[-] Stage-1: Json, %s\n", err)
					}
					log.Printf("DEBUG[*] Stage-1: Unknown File Type, %v, eiT: %T\n", string(einfo), ev)
				}

			}
		case err := <-w.Errors: // RECEIVE eventError
			// check if channel is closed (!ok == closed)
			log.Printf("ERROR[-] Stage-1: Listen %s\n", err)
		}
	}
}

//!-stage-1
