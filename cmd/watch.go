package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var Target string

// [mocking os.Getwd] simply wraps os.Getwd() with getCurrentUserPath
type getCurrentWorkingPath interface {
	Getwd() (string, error)
}

// [mocking os.Getwd] implements Getwd() wrapper interface
type getWorkPath struct{}

// errorString is a trivial implementation of error.
// It has the Error Interface (as ponted to by th Error() method)
// type errorString struct {
// 	s string
// }

// [mocking os.Getwd]
var getWP getCurrentWorkingPath

type (
	EventInfo struct {
		Event Event `json:"event"`
		Meta  Meta  `json:"meta"`
	}

	Event struct {
		Location string `json:"location"`
		Op       string `json:"op"`
	}

	Meta struct {
		ModTime time.Time   `json:"modTime"`
		Mode    os.FileMode `json:"mode"`
		Name    string      `json:"name"`
		Size    int64       `json:"size"`
	}
)

// func (e *errorString) Error() string {
// 	return e.s
// }

// func New(text string) error {
// 	return &errorString{text}
// }

// [mocking os.Getwd] call to os.Getwd (3rd party)
func (g getWorkPath) Getwd() (string, error) {
	return os.Getwd()
}

// [mocking os.Getwd] assign instance of getWorkPath
func init() {
	getWP = getWorkPath{}
}

// func getEventPath(target string) (eventpath string) {
// 	if path.IsAbs(target) {
// 		return target
// 	}
// 	// os.Getwd called through helper mock wrapper
// 	pwd, err := getWP.Getwd()
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	eventpath = path.Join(pwd, target)
// 	return
// }

type FsEventOperations interface {
	EventPath(path string) (string, error)
	FsInfo(path string) (os.FileInfo, error)
}

type FsEventOps struct {
}

func (o *FsEventOps) EventPath(evPath string) (string, error) {
	if path.IsAbs(evPath) {
		return evPath, nil
	}
	pwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrapf(err, "getting cwd for relative path: %s", evPath)
	}
	return path.Join(pwd, evPath), nil
}

func (*FsEventOps) FsInfo(evPath string) (os.FileInfo, error) {
	osFs := afero.NewOsFs()
	return osFs.Stat(evPath)
}

type FsEvent struct {
	event fsnotify.Event
	ops   FsEventOperations
}

func (e *FsEvent) info() (*EventInfo, error) {
	path, err := e.ops.EventPath(e.event.Name)
	if err != nil {
		return nil, err
	}
	fi, err := e.ops.FsInfo(path)
	if err != nil {
		return nil, err
	}
	return &EventInfo{
		Event{
			Location: path,
			Op:       e.event.Op.String(),
		},
		Meta{
			ModTime: fi.ModTime().Truncate(time.Millisecond),
			Mode:    fi.Mode(),
			Name:    fi.Name(),
			Size:    fi.Size(),
		},
	}, nil
}

func activateDirWatcher(targetDir string) {
	// TODO Create a watcher for every target in target file
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// all events are caught by default
				log.Printf("event: %v, eventT: %T", event, event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					fsEv := &FsEvent{
						event: event,
						ops:   &FsEventOps{},
					}
					ev, err := fsEv.info()
					if err != nil {
						log.Printf("error getting event info: %s", err)
						return
					}

					einfo, err := json.Marshal(ev)
					if err != nil {
						log.Printf("error marshaling event: %s", err)
					}
					fmt.Printf("einfol: %v, eiT: %T", string(einfo), ev)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(targetDir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

// versionCmd represents the version command
var cmdWatch = &cobra.Command{
	Use:   "watch",
	Short: "Start the fsnotify file system event watcher",
	Long:  `Use the watch command with a --target flag to indicate in order to watch it.`,
	// Args: func(cmd *cobra.Command, args []string) error {
	// 	if len(args) < 1 {
	// 		return errors.New("requires a color argument")
	// 	}
	// 	if myapp.IsValidColor(args[0]) {
	// 		return nil
	// 	}
	// 	return fmt.Errorf("invalid color specified: %s", args[0])
	// },
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("I AM WATCHING: %v !! \n", Target)
		activateDirWatcher(Target)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cmdWatch)
	// cmdWatch.AddCommand(cmdTarget)
	cmdWatch.Flags().StringVarP(&Target, "target", "t", "", "Target directory to watch (required)")
	cmdWatch.MarkFlagRequired("target")
}
