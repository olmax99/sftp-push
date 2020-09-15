package event

import (
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

// encapsulates file event and the FsEventOperations interface
type FsEvent struct {
	event fsnotify.Event
	ops   FsEventOperations
}

// Captures both the source path of a new file event and
// the respective file info of the target file object.
type FsEventOperations interface {
	EventSrc(path string) (string, error)
	FsInfo(path string) (os.FileInfo, error)
	NewWatcher(path string)
}

// Implements the FsEventOperations interface
type FsEventOps struct{}

type EventInfo struct {
	Event Event `json:"event"`
	Meta  Meta  `json:"meta"`
}

type Event struct {
	Location string `json:"location"`
	Op       string `json:"op"`
}

type Meta struct {
	ModTime time.Time   `json:"modTime"`
	Mode    os.FileMode `json:"mode"`
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
}
