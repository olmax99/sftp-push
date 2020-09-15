package sftppush

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/olmax99/fsnotify"
	"github.com/olmax99/sftppush/pkg/event"
	"github.com/spf13/afero"
)

// Ensure that there is always an absolute path in the output of EventSrc
func Test_EventSrc(t *testing.T) {
	// Create Test files
	// TODO Will only work with ENV 'TESTING' set
	// appfs := afero.NewMemMapFs()
	appfs := afero.NewOsFs()
	afero.WriteFile(appfs, "/tmp/c.txt", []byte("file c"), 0644)
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed test setup: os.Getwd .. %s", err)
	}
	afero.WriteFile(appfs, path.Join(pwd, "c.txt"), []byte("file c"), 0644)

	// cleanup only guaranteed for /tmp/*.* files
	defer appfs.Remove(path.Join(pwd, "c.txt"))

	var Results = []struct {
		in  string
		out string
	}{
		{"/tmp/c.txt", "/tmp/c.txt"},
		{"c.txt", strings.Join([]string{pwd, "c.txt"}, "/")},
	}

	t.Run("Test EventSrc input abs and rel path", func(t *testing.T) {
		for _, rr := range Results {
			eact := &event.FsEventOps{}
			act, _ := eact.EventSrc(rr.in)

			if act != rr.out {
				t.Errorf("EventSrc(%s) =>  %s, want %s", rr.in, act, rr.out)
			}
		}
	})
}

func Test_info(t *testing.T) {
	appfs := afero.NewOsFs()
	afero.WriteFile(appfs, "/tmp/c.txt", []byte("file c"), 0644)
	modt, _ := appfs.Stat("/tmp/c.txt")

	// create INPUT RELATIVE PATH
	var frel = fsnotify.Event{
		Name: "/tmp/c.txt",
		Op:   2,
	}

	// create EXPECTED RELATIVE PATH fsnotify.Event
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

	var Results = []struct {
		in  fsnotify.Event
		out event.EventInfo
	}{
		{frel, eiabs},
	}

	t.Run("Test FsEvent info output", func(t *testing.T) {
		for _, rr := range Results {
			actEv := &event.FsEvent{
				Event: rr.in,
				Ops:   &event.FsEventOps{},
			}

			eact, _ := actEv.Info()

			if eact.Event.Location != rr.out.Event.Location {
				t.Errorf("<&FsEvent>.Info() => %s, want %s", eact.Event.Location, rr.out.Event.Location)
			}

		}
	})
}
