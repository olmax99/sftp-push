package event

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// returns the absolute source path of the triggered file event
func (o *FsEventOps) EventSrc(evPath string) (string, error) {
	if path.IsAbs(evPath) {
		return evPath, nil
	}
	pwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrapf(err, "getting cwd for relative path: %s", evPath)
	}
	return path.Join(pwd, evPath), nil
}

// wrapper around os Stat call for retrieving file info
func (o *FsEventOps) FsInfo(evPath string) (os.FileInfo, error) {
	osFs := afero.NewOsFs()
	return osFs.Stat(evPath)
}

// returns the EventInfo object based on FsEvent interface operations
func (e *FsEvent) info() (*EventInfo, error) {
	path, err := e.ops.EventSrc(e.event.Name)
	if err != nil {
		return nil, err
	}
	fi, err := e.ops.FsInfo(path)
	if err != nil {
		return nil, err
	}
	return &EventInfo{
		Event{
			Location: path,
			Op:       e.event.Op.String(),
		},
		Meta{
			ModTime: fi.ModTime().Truncate(time.Millisecond),
			Mode:    fi.Mode(),
			Name:    fi.Name(),
			Size:    fi.Size(),
		},
	}, nil
}

// Implements fsnotify file event watcher on a target directory
func (o *FsEventOps) NewWatcher(targetDir string) {
	// TODO Create a watcher for every target in target file
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// all events are caught by default
				log.Printf("event: %v, eventT: %T", event, event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					fsEv := &FsEvent{
						event: event,
						ops:   &FsEventOps{},
					}
					ev, err := fsEv.info()
					if err != nil {
						log.Printf("error getting event info: %s", err)
						return
					}

					einfo, err := json.Marshal(ev)
					if err != nil {
						log.Printf("error marshaling event: %s", err)
					}
					fmt.Printf("einfo: %v, eiT: %T", string(einfo), ev)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(targetDir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
