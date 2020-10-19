package event

import (
	"os"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// reduceEventPath returns the absolute Event path reduced by the gCfg userpath
func (o *FsEventOps) reduceEventPath(evp string, cfgp *string) (string, error) {
	cfgUserPath := *cfgp
	eventPath := strings.Split(evp, "/")
	eventPathDepth := len(eventPath)
	reducedPath := make([]string, 0)

	for i := 0; i < eventPathDepth; i++ {
		dEventPath := eventPath[:len(eventPath)-i]
		evPath := strings.Join(dEventPath, "/")
		b, err := path.Match(cfgUserPath, evPath+"/")
		if err != nil {
			return "", errors.Wrapf(err, "Derive relative event path: %s != %s", evp, *cfgp)
		}
		if b {
			break
		}
		reducedPath = append([]string{eventPath[len(dEventPath)-1]}, reducedPath...)
	}
	res := strings.Join(reducedPath, "/")
	res = strings.TrimSuffix(res, ".gzip")
	res = strings.TrimSuffix(res, ".gz")
	return res, nil
}

//  EventSrc returns the absolute source path of the triggered file event
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
func (e *FsEvent) Info() (*EventInfo, error) {
	path, err := e.Ops.EventSrc(e.Event.Name)
	if err != nil {
		return nil, err
	}

	fi, err := e.Ops.FsInfo(path)
	if err != nil {
		return nil, err
	}

	return &EventInfo{
		Event{
			AbsLoc: path,
			Op:     e.Event.Op.String(),
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
func (o *FsEventOps) NewWatcher(epIn *EventPushInfo, lg *logrus.Logger) {
	//!+stage-0
	// 1. Sets up the Pipeline
	// 2. runs the final stage <- receiving from all open channels
	watcher, err := fsnotify.NewWatcher() // watcher: implements producer stage-0
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close() // close SEND Channel

	// Add directories to *Watcher
	for _, d := range epIn.Watchdirs {
		err = watcher.Add(d)
		if err != nil {
			lg.Fatalf("NewWatcher.Add %s, %s", d, err)
		}
	}
	//!-stage-0

	//!+stage-1
	//!+stage-2
	done := make(chan bool)

	targetEvent := make(chan EventInfo)
	// eventErr := make(chan errors)

	go o.listen(watcher, targetEvent) // fsnotify event implementation
	go o.controlWorkers(targetEvent, epIn)

	// Wait for all results in the background
	go func() {
		for f := range epIn.Results {
			log.Printf("INFO[+] Results: %#v\n", f)
		}
	}()
	<-done // Block for listen, controlWorkers to run
	//!-stage-2
	//!-stage-1
}
