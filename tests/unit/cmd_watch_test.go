package sftppush

import (
	"encoding/json"
	"os"
	"path"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/olmax99/sftppush/cmd"
	"github.com/spf13/afero"
)

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
	var eiabs cmd.EventInfo = cmd.EventInfo{
		cmd.Event{
			Location: "/tmp/c.txt",
			Op:       "WRITE",
		},
		cmd.Meta{
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
	var eirel cmd.EventInfo = cmd.EventInfo{
		cmd.Event{
			Location: path.Join(pwd, "c.txt"),
			Op:       "WRITE",
		},
		cmd.Meta{
			ModTime: modt.ModTime().Truncate(time.Millisecond),
			Mode:    420,
			Name:    "c.txt",
			Size:    6,
		},
	}

	var Results = []struct {
		in  fsnotify.Event
		out cmd.EventInfo
	}{
		{fabs, eiabs},
		{frel, eirel},
	}

	t.Run("Test input abs and rel path", func(t *testing.T) {
		for _, rr := range Results {
			esrc := cmd.EventSrc(rr.in)
			act := esrc()

			a, _ := json.Marshal(act)
			// t.Logf("ACTUAL: %v", string(a))

			// ei needs to go into a slice
			s := make([]cmd.EventInfo, 0)
			s = append(s, rr.out)

			e, _ := json.Marshal(s)
			// t.Logf("EXPECTED: %v", string(e))

			if act[0] != s[0] {
				t.Errorf("eventSrc(%v) =>  %v, want %v", rr.in, string(e), string(a))
			}
		}
	})
}
