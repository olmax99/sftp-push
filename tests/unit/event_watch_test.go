package sftppush

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/olmax99/sftppush/pkg/event"
	"github.com/spf13/afero"
)

func RemoveF(t *testing.T, p string, a afero.Fs) {
	if err := a.Remove(p); err != nil {
		t.Errorf("RemoveF, %s\n", err)
	}
}

// Ensure that there is always an absolute path in the output of EventSrc
func Test_EventSrc(t *testing.T) {
	// Create Test files
	// TODO Will only work with ENV 'TESTING' set
	// appfs := afero.NewMemMapFs()
	appfs := afero.NewOsFs()
	if err := afero.WriteFile(appfs, "/tmp/c.txt", []byte("file c"), 0644); err != nil {
		t.Errorf("afero WriteFile, %s\n", err)
	}
	pwd, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed test setup: os.Getwd .. %s", err)
	}
	if err := afero.WriteFile(appfs, path.Join(pwd, "c.txt"), []byte("file c"), 0644); err != nil {
		t.Errorf("afero WriteFile, %s\n", err)
	}

	// cleanup only guaranteed for /tmp/*.* files
	defer RemoveF(t, path.Join(pwd, "c.txt"), appfs)

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
	if err := afero.WriteFile(appfs, "/tmp/c.txt", []byte("file c"), 0644); err != nil {
		t.Errorf("afero WriteFile, %s\n", err)
	}
	modt, _ := appfs.Stat("/tmp/c.txt")

	// create INPUT RELATIVE PATH
	var frel = fsnotify.Event{
		Name: "/tmp/c.txt",
		Op:   2,
	}

	// create EXPECTED RELATIVE PATH fsnotify.Event
	var eiabs event.EventInfo = event.EventInfo{
		Event: event.Event{
			AbsLoc: "/tmp/c.txt",
			Op:     "WRITE",
		},
		Meta: event.Meta{
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

			if eact.Event.AbsLoc != rr.out.Event.AbsLoc {
				t.Errorf("<&FsEvent>.Info() => %s, want %s", eact.Event.AbsLoc, rr.out.Event.AbsLoc)
			}

		}
	})
}
