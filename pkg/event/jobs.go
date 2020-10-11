package event

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/fsnotify/fsnotify"
)

//!+stage-4

// Remove waits for S3 upload to finish and removes event file
func (o FsEventOps) Remove(e EventInfo) {
	// TODO This seems to conflict - need to be outside of WaitGroup()
	// defer wg.Done()
	if err := os.Remove(e.Event.AbsLoc); err != nil {
		log.Printf("ERROR[-] Stage-4: Remove %s", err)
	}
}

//!-stage-4

//!+stage-3

// PushS3 uploads the source event file byte stream to S3 and removes the file
func (o FsEventOps) PushS3(in io.Reader, pi EventPushInfo, wg *sync.WaitGroup, ei EventInfo) {
	defer wg.Done()
	uploader := s3manager.NewUploaderWithClient(pi.Session, func(u *s3manager.Uploader) {
		u.PartSize = 64 * 1024 * 1024 // 64MB per part
	})
	upl := s3manager.UploadInput{
		Body:   in,
		Bucket: pi.Bucket,
		Key:    &pi.Key,
	}

	r, err := uploader.Upload(&upl)
	if err != nil {
		// TODO send upload error to RETRY
		log.Printf("WARNING[-] Stage-3: PushS3 %s", err)
	} else {
		pi.Results <- &ResultInfo{response: r, eventInfo: ei}
	}

}

//!-stage-3

//!+stage-2

// FType detects and returns the file type along with the initial file io.Reader
func (o *FsEventOps) FType(epath string) (string, *io.Reader) {
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

// Decompress detects the file type and sends the decompressed byte stream to the PushS3 stage
func (o *FsEventOps) Decompress(in <-chan EventInfo, pi *EventPushInfo) {
	var wg sync.WaitGroup
	for e := range in {

		p := e.Event.AbsLoc
		//cfgp := *apath

		ft, b := o.FType(p)
		switch ft {
		case "application/x-gzip":
			log.Printf("DEBUG[*] Stage-2: fT %s, %s\n", ft, filepath.Base(p))
			gz, err := gzip.NewReader(*b)
			if err != nil {
				log.Printf("ERROR[-] Stage-2: gzip.NewReader, %s\n", err)
			}

			pi.Key, err = o.reduceEventPath(p, pi.Userpath)
			if err != nil {
				log.Printf("ERROR[*] Stage-2: %s", err)
			}

			wg.Add(1)
			go o.PushS3(gz, *pi, &wg, e)
		case "application/zip":
			log.Printf("DEBUG[*] Stage-2: fT %s, %s\n", ft, filepath.Base(p))
		default:
			// if strings.HasPrefix(string(buf), "\x42\x5a\x68") {
			// 	log.Printf("INFO[*] Stage-1: file type %s, %s\n", ft, filepath.Base(p))
			// } else {}
			log.Printf("WARNING[-] Stage-2: unexpected fT %s, %s\n", ft, filepath.Base(p))
		}

	}
	go func() {
		// Blocking until stage-2 and stage-3 are finished
		wg.Wait()
	}()
}

//!-stage-2

//!+stage-1

// Listen listens to file events from fsnotify.Watcher and sends them to the stage-1 channel
func (o *FsEventOps) Listen(w *fsnotify.Watcher, out chan<- EventInfo) {
	for {
		select {
		case event, ok := <-w.Events: // RECEIVE event
			// check if channel is closed (!ok == closed)
			if !ok {
				return
			}
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

		case err, ok := <-w.Errors: // RECEIVE eventError
			log.Printf("ERROR[-] Stage-1: Listen %s\n", err)
			// check if channel is closed (!ok == closed)
			if !ok {
				// TODO Will this exit the Listen() func?? when channel
				// w.Errors gets closed??
				// Where gets the w.Errors channel closed??
				return
			}
		}
	}
}

//!-stage-1
