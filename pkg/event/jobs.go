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

//!+job-2

// Remove waits for S3 upload to finish and removes event file
func (o FsEventOps) Remove(e EventInfo) {
	// TODO This seems to conflict - need to be outside of WaitGroup()
	// defer wg.Done()
	if err := os.Remove(e.Event.AbsLoc); err != nil {
		log.Printf("ERROR[-] Job-2: Remove %s", err)
	}
}

// PushS3 uploads the source event file byte stream to S3 and removes the file
func (o FsEventOps) PushS3(in io.Reader, pi EventPushInfo, wg *sync.WaitGroup, ei EventInfo) {
	defer wg.Done()
	uploader := s3manager.NewUploaderWithClient(pi.session, func(u *s3manager.Uploader) {
		u.PartSize = 64 * 1024 * 1024 // 64MB per part
	})
	upl := s3manager.UploadInput{
		Body:   in,
		Bucket: &pi.bucket,
		Key:    &pi.key,
	}

	r, err := uploader.Upload(&upl)
	if err != nil {
		// TODO send upload error to RETRY
		log.Printf("WARNING[-] Job-2: PushS3 %s", err)
	} else {
		pi.results <- &ResultInfo{response: r, eventInfo: ei}
		// TODO This will only work for the first few files - refine concurrency design
		// go o.Remove(*ei, wg)

	}

}

//!-job-2

//!+job-1

// type ReadCloseFile struct {
// 	r io.Reader
// 	c io.Closer
// }

// FType detects and returns the file type along with the initial file io.Reader
func (o *FsEventOps) FType(epath string) (string, *io.Reader) {
	// f is an io.Reader
	f, err := os.Open(epath)
	if err != nil {
		log.Printf("WARNING[-] Job-1: Open %s, %s\n", filepath.Base(epath), err)
	}

	buf := make([]byte, 32)
	if _, err := f.Read(buf); err != nil {
		log.Printf("ERROR[-] Job-1: File Read %s, %s\n", filepath.Base(epath), err)
	}
	fT := http.DetectContentType(buf)

	// glue those bytes back onto the reader
	r := io.MultiReader(bytes.NewReader(buf), f)

	return fT, &r
}

// Decompress detects the file type and sends the decompressed byte stream to the PushS3 job
func (o *FsEventOps) Decompress(in <-chan EventInfo, pi EventPushInfo, apath *string) {
	var wg sync.WaitGroup
	for e := range in {
		newPi := &pi

		p := e.Event.AbsLoc
		cfgp := *apath

		ft, b := o.FType(p)
		switch ft {
		case "application/x-gzip":
			log.Printf("DEBUG[*] Job-1: fT %s, %s\n", ft, filepath.Base(p))
			// implement io.Closer for gzip
			gz, err := gzip.NewReader(*b)
			if err != nil {
				log.Printf("ERROR[-] Job-1: gzip.NewReader, %s\n", err)
			}

			newPi.key, err = o.reduceEventPath(p, cfgp)
			if err != nil {
				log.Printf("ERROR[*] Job-1: %s", err)
			}

			wg.Add(1)
			go o.PushS3(gz, *newPi, &wg, e)
		case "application/zip":
			log.Printf("DEBUG[*] Job-1: fT %s, %s\n", ft, filepath.Base(p))
		default:
			// if strings.HasPrefix(string(buf), "\x42\x5a\x68") {
			// 	log.Printf("INFO[*] Job-1: file type %s, %s\n", ft, filepath.Base(p))
			// } else {}
			log.Printf("WARNING[-] Job-1: unexpected fT %s, %s\n", ft, filepath.Base(p))
		}

	}
	go func() {
		// Blocking until job-1 and job-2 are finished
		wg.Wait()
	}()
}

//!-job-1

//!+job-0

// Listen listens to file events from fsnotify.Watcher and sends them to the job-1 channel
func (o *FsEventOps) Listen(w *fsnotify.Watcher, out chan<- EventInfo) {
	for {
		select {
		case event, ok := <-w.Events: // RECEIVE event
			if !ok {
				return
			}
			// all events are caught by default
			log.Printf("DEBUG[+] Job-0: %v, eventT: %T\n", event, event)
			if event.Op&fsnotify.CloseWrite == fsnotify.CloseWrite {
				fsEv := &FsEvent{
					Event: event,
					Ops:   &FsEventOps{},
				}
				ev, err := fsEv.Info()
				if err != nil {
					log.Printf("WARNING[-] Job-0: Listen %s\n", err)
				}

				// 32 bytes needed for determining file type
				if ev.Meta.Size >= int64(32) {
					out <- *ev // SEND needs no close as infinite amount of Events
				} else {
					// only for testing
					einfo, err := json.Marshal(ev)
					if err != nil {
						log.Printf("ERROR[-] Job-0: Json, %s\n", err)
					}
					log.Printf("DEBUG[*] Job-0: Unknown File Type, %v, eiT: %T\n", string(einfo), ev)
				}

			}

		case err, ok := <-w.Errors: // RECEIVE eventError
			if !ok {
				// TODO Will this exit the Listen() func?? when channel
				// w.Errors gets closed??
				// Where gets the w.Errors channel closed??
				return
			}
			log.Printf("ERROR[-] Job-0: Listen %s\n", err)
		}
	}
}

//!-job-0
