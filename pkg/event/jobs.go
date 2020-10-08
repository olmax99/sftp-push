package event

import (
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
func (o FsEventOps) Remove(e EventInfo, wg *sync.WaitGroup) {
	// TODO This seems to conflict - need to be outside of WaitGroup()
	defer wg.Done()
	if err := os.Remove(e.Event.AbsLoc); err != nil {
		log.Printf("ERROR[-] Job-2: Remove %s", err)
	}
}

// PushS3 uploads the source event file byte stream to S3 and removes the file
func (o FsEventOps) PushS3(in io.ReadCloser, pi EventPushInfo, wg *sync.WaitGroup, ei *EventInfo) {
	// defer wg.Done()
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
		pi.results <- r
		// TODO This will only work for the first few files - refine concurrency design
		go o.Remove(*ei, wg)

	}

	if err := in.Close(); err != nil {
		log.Printf("WARNING[-] Job-2: PushS3 Close  %s", err)
	}
}

//!-job-2

//!+job-1

type ReadCloseFile struct {
	io.Reader
	io.Closer
}

// FType detects and returns the file type of the event location file object
func (o FsEventOps) FType(epath string) (string, *os.File) {
	f, err := os.Open(epath)
	if err != nil {
		log.Printf("WARNING[-] Job-1: os.Open %s, %s\n", filepath.Base(epath), err)
	}

	buf := make([]byte, 512)
	if _, err := f.Read(buf); err != nil {
		log.Printf("ERROR[-] Job-1: File Read %s, %s\n", filepath.Base(epath), err)
	}
	f.Seek(0, io.SeekStart)

	fT := http.DetectContentType(buf)
	return fT, f
}

// Decompress detects the file type and sends the decompressed byte stream to the PushS3 job
func (o *FsEventOps) Decompress(in <-chan EventInfo, pi EventPushInfo, apath *string) {
	var wg sync.WaitGroup
	for e := range in {
		newPi := &pi

		p := e.Event.AbsLoc
		cfgp := *apath

		ft, f := o.FType(p)
		switch ft {
		case "application/x-gzip":
			log.Printf("DEBUG[*] Job-1: fT %s, %s\n", ft, filepath.Base(p))
			// implement io.Closer for gzip
			gz, err := func(i *os.File) (io.ReadCloser, error) {
				g, err := gzip.NewReader(i)
				if err != nil {
					return nil, err
				}
				return &ReadCloseFile{g, i}, nil
			}(f)
			if err != nil {
				log.Printf("ERROR[-] Job-1: gzip.NewReader, %s\n", err)
			}

			newPi.key, err = o.reduceEventPath(p, cfgp)
			if err != nil {
				log.Printf("ERROR[*] Job-1: %s", err)
			}

			wg.Add(1)
			go o.PushS3(gz, *newPi, &wg, &e)
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

				// -> Process file through decompress job-1
				//     -> Process decompressed stream to job-2 S3 upload
				out <- *ev // SEND needs no close as infinite amount of Events

				// only for testing
				einfo, err := json.Marshal(ev)
				if err != nil {
					log.Printf("ERROR[-] Job-0: Json, %s\n", err)
				}
				log.Printf("DEBUG[*] Job-0: %v, eiT: %T\n", string(einfo), ev)
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
