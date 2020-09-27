package event

import (
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

// encapsulates file event and the FsEventOperations interface
type FsEvent struct {
	Event fsnotify.Event
	Ops   FsEventOperations
}

// Captures both the source path of a new file event and
// the respective file info of the target file object.
type FsEventOperations interface {
	EventSrc(path string) (string, error)
	FsInfo(path string) (os.FileInfo, error)
	NewWatcher(path string)
	Listen(watcher fsnotify.Watcher, targetevents chan<- EventInfo)
	Decompress(targetevents <-chan EventInfo)
}

// Implements the FsEventOperations interface
type FsEventOps struct{}

// Parent of the Event output with two children: Event and Meta
type EventInfo struct {
	Event Event `json:"event"`
	Meta  Meta  `json:"meta"`
}

// Implements child of parent EventInfo
type Event struct {
	Location string `json:"location"`
	Op       string `json:"op"`
}

// Implements child of parent EventInfo
type Meta struct {
	ModTime time.Time   `json:"modTime"`
	Mode    os.FileMode `json:"mode"`
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
}
