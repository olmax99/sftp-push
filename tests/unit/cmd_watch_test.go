package sftppush

import (
	"encoding/json"
	"testing"

	"github.com/fsnotify/fsnotify"
	"github.com/olmax99/sftppush/cmd"
	"github.com/spf13/afero"
)

func Test_EventSrc(t *testing.T) {
	// Create Test file
	appfs := afero.NewOsFs()
	afero.WriteFile(appfs, "/tmp/c.txt", []byte("file c"), 0644)

	modt, _ := appfs.Stat("/tmp/c.txt")

	// create INPUT
	var fsne = fsnotify.Event{
		Name: "/tmp/c.txt",
		Op:   2,
	}

	// create EXPECTED fsnotify.Event
	var ei cmd.EventInfo = cmd.EventInfo{
		cmd.Event{
			Location: "/tmp/c.txt",
			Op:       "WRITE",
		},
		cmd.Meta{
			ModTime: modt.ModTime(),
			Mode:    420,
			Name:    "c.txt",
			Size:    6,
		},
	}

	var Results = []struct {
		in  fsnotify.Event
		out cmd.EventInfo
	}{
		{fsne, ei},
	}
	t.Run("Test standard output", func(t *testing.T) {
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
