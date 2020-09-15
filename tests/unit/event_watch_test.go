package sftppush

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/olmax99/sftppush/pkg/event"
	"github.com/spf13/afero"
)

// var getwdMock func() (string, error)

// type getWorkPathMock struct{}

// func (p getWorkPathMock) Getwd() (string, error) {
// 	return getwdMock()
// }

func Test_EventSrc(t *testing.T) {
	// Create Test files
	// TODO Will only work with ENV 'TESTING' set
	// appfs := afero.NewMemMapFs()
	appfs := afero.NewOsFs()
	afero.WriteFile(appfs, "/tmp/c.txt", []byte("file c"), 0644)
	pwd, _ := os.Getwd()
	afero.WriteFile(appfs, path.Join(pwd, "c.txt"), []byte("file c"), 0644)

	modt, _ := appfs.Stat("/tmp/c.txt")

	// cleanup only guaranteed for /tmp/*.* files
	defer appfs.Remove(path.Join(pwd, "c.txt"))

	// create INPUT ABS PATH
	var fabs = fsnotify.Event{
		Name: "/tmp/c.txt",
		Op:   2,
	}

	// create EXPECTED ABS PATH fsnotify.Event
	var eiabs event.EventInfo = event.EventInfo{
		event.Event{
			Location: "/tmp/c.txt",
			Op:       "WRITE",
		},
		event.Meta{
			ModTime: modt.ModTime().Truncate(time.Millisecond),
			Mode:    420,
			Name:    "c.txt",
			Size:    6,
		},
	}

	// create INPUT RELATIVE PATH
	var frel = fsnotify.Event{
		Name: path.Join(pwd, "c.txt"),
		Op:   2,
	}

	// create EXPECTED RELATIVE PATH fsnotify.Event
	var eirel event.EventInfo = event.EventInfo{
		event.Event{
			Location: path.Join(pwd, "c.txt"),
			Op:       "WRITE",
		},
		event.Meta{
			ModTime: modt.ModTime().Truncate(time.Millisecond),
			Mode:    420,
			Name:    "c.txt",
			Size:    6,
		},
	}

	var Results = []struct {
		in  fsnotify.Event
		out event.EventInfo
	}{
		{fabs, eiabs},
		{frel, eirel},
	}

	t.Run("Test input abs and rel path", func(t *testing.T) {
		for _, rr := range Results {
			eact := &event.FsEventOps{}
			act, _ := eact.EventSrc(rr.in.Name)

			if act != rr.out.Event.Location {
				t.Errorf("eventSrc(%s) =>  %s, want %s", rr.in.Name, act, rr.out.Event.Location)
			}
		}
	})
}

// func Test_EventSrcErrorPwdError(t *testing.T) {
// 	pwd, _ := os.Getwd()

// 	// create INPUT RELATIVE PATH
// 	var frel = fsnotify.Event{
// 		Name: path.Join(pwd, "c.txt"),
// 		Op:   2,
// 	}

// 	// create EXPECTED RELATIVE PATH fsnotify.Event
// 	var eirel cmd.EventInfo = cmd.EventInfo{
// 		cmd.Event{
// 			Location: path.Join(pwd, "c.txt"),
// 			Op:       "WRITE",
// 		},
// 		cmd.Meta{
// 			ModTime: modt.ModTime().Truncate(time.Millisecond),
// 			Mode:    420,
// 			Name:    "c.txt",
// 			Size:    6,
// 		},
// 	}

// 	var Results = []struct {
// 		in  fsnotify.Event
// 		out cmd.EventInfo
// 	}{
// 		{frel, eirel},
// 	}

// 	t.Run("Test failing Getwd", func(t *testing.T) {

// 		getWP := getWorkPathMock{}

// 		getwdMock = func() (string, error) {
// 			return "", errors.New("Getwd failed.")
// 		}

// 		for _, rr := range Results {

// 			esrc := cmd.EventSrc(rr.in)
// 			_, err := esrc()
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 		}
// 	})
// }
