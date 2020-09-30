package event

import (
	"io"
	"os"
	"sync"
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
	NewWatcher(path string, conn *s3.S3, bucket *string)
	FType(path string) (string, error)
	Listen(watcher *fsnotify.Watcher, targetevents chan<- EventInfo)
	Decompress(targetevents <-chan EventInfo, session *s3.S3, bucket *string, s3results chan<- *s3manager.UploadOutput)
	PushS3(bytes io.Reader, sess *s3.S3, s3bucket *string, s3key *string, results chan<- *s3manager.UploadOutput, wg *sync.WaitGroup)
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
	RelLoc string `json:"relloc"`
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
