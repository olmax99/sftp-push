package event

import (
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// reduceEventPath returns the absolute Event path reduced by the gCfg userpath
func (o *FsEventOps) reduceEventPath(evp string, cfgp string) (string, error) {
	eventPath := strings.Split(evp, "/")
	eventPathDepth := len(eventPath)
	reducedPath := make([]string, 0)

	for i := 0; i < eventPathDepth; i++ {
		dEventPath := eventPath[:len(eventPath)-i]
		evPath := strings.Join(dEventPath, "/")
		b, err := path.Match(cfgp, evPath+"/")
		if err != nil {
			log.Printf("ERROR[-] Derive relative event path, %s", err)
			return "", errors.Wrapf(err, "Derive relative event path: %s != %s", evp, cfgp)
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
func (o *FsEventOps) NewWatcher(targetDirs []string, conn *s3.S3, targetBucket *string, apath *string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close() // close SEND Channel

	// Should this channel not be passed on from fsnotify.NewWatcher?
	// watcher.done seems to be private and not accessible from here..
	// How is a new done channel even working here?
	// fmt.Printf("%#v", watcher)
	done := make(chan bool)
	targetevent := make(chan EventInfo)
	results := make(chan *ResultInfo)

	epi := EventPushInfo{
		session: conn,
		bucket:  *targetBucket,
		key:     "",
		results: results,
	}

	go o.Listen(watcher, targetevent) // fsnotify event implementation
	go o.Decompress(targetevent, epi, apath)

	// Wait for all results in the background
	go func() {
		for f := range results {
			//log.Printf("INFO[+] Results: %#v\n", f)
			go o.Remove(f.eventInfo)
		}
	}()

	// Add directories to *Watcher
	for _, d := range targetDirs {
		err = watcher.Add(d)
		if err != nil {
			log.Printf("ERROR[-] NewWatcher.Add %s, %s", d, err)
		}
	}

	// Which token is released here? as there was none being send to new done channel..
	// Is this just blocking? Why?
	<-done // Release the token
}
