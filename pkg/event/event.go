package event

import (
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/fsnotify/fsnotify"
)

// FsEvent encapsulates file event and the FsEventOperations interface
type FsEvent struct {
	Event fsnotify.Event
	Ops   FsEventOperations
}

// FsEventOperations contains all methods necessarry for processing local file events
type FsEventOperations interface {
	EventSrc(path string) (string, error)
	FsInfo(path string) (os.FileInfo, error)
	NewWatcher(info *EventPushInfo)
	FType(path string) (string, *io.Reader)
	Listen(watcher *fsnotify.Watcher, targetevents chan<- EventInfo)
	Decompress(targetevents <-chan EventInfo, pinfo *EventPushInfo)
	PushS3(done <-chan struct{}, bytes io.Reader, pinfo EventPushInfo, einfo EventInfo) <-chan *ResultInfo
	reduceEventPath(p string, cfgp *string) (string, error)
	Remove(event EventInfo)
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
	AbsLoc string `json:"absloc"`
	Op     string `json:"op"`
}

// Implements child of parent EventInfo
type Meta struct {
	ModTime time.Time   `json:"modTime"`
	Mode    os.FileMode `json:"mode"`
	Name    string      `json:"name"`
	Size    int64       `json:"size"`
}

// EventPushInfo contains the common data for SftpPush Stages
type EventPushInfo struct {
	Session   *s3.S3
	Userpath  *string
	Watchdirs []string
	Bucket    *string
	Key       string
	Results   chan *ResultInfo
}

// ResultInfo is the data returned in the results channel
type ResultInfo struct {
	response  *s3manager.UploadOutput
	eventInfo EventInfo
}
