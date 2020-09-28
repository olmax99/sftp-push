package event

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

func Write(p []byte, out chan<- byte) (int, error) {
	n := 0
	for _, b := range p {
		out <- b
		n++
	}
	return n, nil
}

func Close(out chan<- byte) error {
	close(out)
	return nil
}

func fType(epath string) (string, error) {
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

//!+job-1

func (o *FsEventOps) Decompress(in <-chan EventInfo, out chan<- byte) {
	for e := range in {
		log.Printf("INFO[*] Job-1: Decompress start ..\n")
		p := e.Event.Location

		ft, err := fType(p)
		if err != nil {
			// TODO forward to main err chan
			log.Fatalf("ERROR[-] Job-1: fType %s %s", err)
		}

		f, err := os.Open(p)
		if err != nil {
			log.Printf("WARNING[-] Job-1: os.Open %s, %s\n", filepath.Base(p), err)
		}
		defer f.Close()

		switch ft {
		case "application/x-gzip":
			log.Printf("INFO[*] Job-1: fT %s, %s\n", ft, filepath.Base(p))
			gz, err := gzip.NewReader(f)
			if err != nil {
				log.Printf("WARNING[-] Job-1: gzip.NewReader, %s", err)
			}
			q := make([]byte, 512) // match bytes chan decompressed
			for {
				// Decompress and forward byte stream
				n, err := gz.Read(q)
				if err != nil {
					if err != io.EOF {
						log.Printf("WARNING[-] Job-1: gzip.Read, %s", err)
						break
					}
					if n == 0 {
						log.Print("WARNING[-] Job-1: gzip.Read, EMPTY.")
						break
					}
				}
				m, err := Write(q, out) // SEND bytes
				if err != nil {
					log.Printf("WARNING[-] Job-1: bytes.Write, %s", err)
				}
				if m == 0 {
					break
				}
			}
			// Close(out)
		case "application/zip":
			log.Printf("INFO[*] Job-1: fT %s, %s\n", ft, filepath.Base(p))
		default:
			// if strings.HasPrefix(string(buf), "\x42\x5a\x68") {
			// 	log.Printf("INFO[*] Job-1: file type %s, %s\n", ft, filepath.Base(p))
			// } else {}
			log.Printf("WARNING[-] Job-1: unexpected fT %s, %s", ft, filepath.Base(p))
		}

	}
}

//!-job-1

//+job-0

// listens to file events from fsnotify.Watcher
func (o *FsEventOps) Listen(w fsnotify.Watcher, out chan<- EventInfo) {
	for {
		select {
		case event, ok := <-w.Events: // RECEIVE event
			if !ok {
				return
			}
			// all events are caught by default
			log.Printf("event: %v, eventT: %T", event, event)
			if event.Op&fsnotify.CloseWrite == fsnotify.CloseWrite {
				fsEv := &FsEvent{
					Event: event,
					Ops:   &FsEventOps{},
				}
				ev, err := fsEv.Info()
				if err != nil {
					log.Printf("error getting event info: %s", err)
				}

				// -> Process file through decompress job-1
				//     -> Process decompressed stream to job-2 S3 upload
				out <- *ev // SEND needs close

				// only for testing
				einfo, err := json.Marshal(ev)
				if err != nil {
					log.Printf("error marshaling event: %s", err)
				}
				log.Printf("DEBUG[*] einfo: %v, eiT: %T\n", string(einfo), ev)
			}

		case err, ok := <-w.Errors: // RECEIVE eventError
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}
}

//!-job-0
