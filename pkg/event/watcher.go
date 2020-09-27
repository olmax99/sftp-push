package event

import (
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
			Location: path,
			Op:       e.Event.Op.String(),
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
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close() // close SEND Channel

	// Should this channel not be passed on from fsnotify.NewWatcher?
	// watcher.done seems to be private and not acessible from here..
	// How is a new done channel even working here?
	// fmt.Printf("%#v", watcher)
	done := make(chan bool)
	targetevent := make(chan EventInfo)
	// decompressed := make(chan string)

	go o.Listen(*watcher, targetevent)
	go o.Decompress(targetevent)
	// go o.Push(decompressed)

	// Add directory to *Watcher
	err = watcher.Add(targetDir)
	if err != nil {
		log.Printf("watcher.Add %s", err)
	}

	// Which token is released here? as there was none being send to new done channel..
	// Is this just blocking? Why?
	<-done // Release the token
}
