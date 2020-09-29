package event

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/fsnotify/fsnotify"
)

//!+job-2

func (o FsEventOps) PushS3(in io.Reader, c *s3.S3, s3t *string, obj *string, out chan<- *s3manager.UploadOutput) {
	uploader := s3manager.NewUploaderWithClient(c, func(u *s3manager.Uploader) {
		u.PartSize = 64 * 1024 * 1024 // 64MB per part
	})
	upl := s3manager.UploadInput{
		Body:   in,
		Bucket: s3t,
		Key:    obj,
	}
	r, err := uploader.Upload(&upl)
	if err != nil {
		log.Printf("WARNING[-] Job-2: PushS3 %s", err)
	}
	out <- r
}

//!-job-2

//!+job-1

// FType detects and returns the file type of the event location file object
func (o FsEventOps) FType(epath string) (string, error) {
	f, err := os.Open(epath)
	if err != nil {
		log.Printf("WARNING[-] Job-1: os.Open %s, %s\n", filepath.Base(epath), err)
		return "", err
	}
	defer f.Close()

	// 512 max data used for file type detection
	buf := make([]byte, 512)
	if _, err := f.Read(buf); err != nil {
		log.Printf("WARNING[-] Job-1: f.Read %s, %s\n", filepath.Base(epath), err)
		return "", err
	}

	fT := http.DetectContentType(buf)
	return fT, nil
}

// Decompress detects the file type and sends the decompressed byte stream to the PushS3 job
func (o *FsEventOps) Decompress(in <-chan EventInfo, s3client *s3.S3, s3bucket *string, out chan<- *s3manager.UploadOutput) {
	for e := range in {
		log.Printf("DEBUG[*] Job-1: Decompress start ..\n")
		p := e.Event.AbsLoc
		ft, err := o.FType(p)
		if err != nil {
			// TODO forward to main err chan
			log.Fatalf("ERROR[-] Job-1: fType %s %s", ft, err)
		}

		f, err := os.Open(p)
		if err != nil {
			log.Printf("WARNING[-] Job-1: os.Open %s, %s\n", filepath.Base(p), err)
		}
		defer f.Close()

		switch ft {
		case "application/x-gzip":
			log.Printf("DEBUG[*] Job-1: fT %s, %s\n", ft, filepath.Base(p))
			gz, err := gzip.NewReader(f)
			if err != nil {
				log.Printf("WARNING[-] Job-1: gzip.NewReader, %s\n", err)
			}
			go o.PushS3(gz, s3client, s3bucket, &e.Event.RelLoc, out)
		case "application/zip":
			log.Printf("DEBUG[*] Job-1: fT %s, %s\n", ft, filepath.Base(p))
		default:
			// if strings.HasPrefix(string(buf), "\x42\x5a\x68") {
			// 	log.Printf("INFO[*] Job-1: file type %s, %s\n", ft, filepath.Base(p))
			// } else {}
			log.Printf("WARNING[-] Job-1: unexpected fT %s, %s\n", ft, filepath.Base(p))
		}

	}
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
				out <- *ev // SEND needs close

				// only for testing
				einfo, err := json.Marshal(ev)
				if err != nil {
					log.Printf("ERROR[-] Job-0: Json, %s\n", err)
				}
				log.Printf("DEBUG[*] Job-0: %v, eiT: %T\n", string(einfo), ev)
			}

		case err, ok := <-w.Errors: // RECEIVE eventError
			if !ok {
				return
			}
			log.Printf("ERROR[-] Job-0: Listen %s\n", err)
		}
	}
}

//!-job-0
