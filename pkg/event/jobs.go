package event

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
)

//!+job-1

func (o *FsEventOps) Decompress(in <-chan EventInfo) {
	for e := range in {
		fmt.Printf("INFO[*] Job-1: Decompress start ..\n")
		p := e.Event.Location
		// Do work
		var delay time.Duration = 2 * time.Second
		time.Sleep(delay)
		fmt.Printf("INFO[*] Job-1: Done %s\n", p)
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
				fmt.Printf("DEBUG[*] einfo: %v, eiT: %T\n", string(einfo), ev)
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
