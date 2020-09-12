package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/fsnotify/fsnotify"
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

// Parses the Event FileInfo and returns an EventInfo slice or error
func EventSrc(e fsnotify.Event) func() ([]EventInfo, error) {
	ep := e.Name
	// nil by default
	var eo error
	ein := EventInfo{}
	einfo := make([]EventInfo, 0)
	if !path.IsAbs(e.Name) {
		pwd, err := os.Getwd()
		if err != nil {
			// log.Fatal(err)
			eo = errors.New("[EventSrc(event)] os.Getwd failed..")
		}
		ep = path.Join(pwd, e.Name)
	}
	return func() ([]EventInfo, error) {
		if eo != nil {
			return einfo, eo
		}
		// TODO Create ENV 'TESTING' and set to afero.NewMemMapFs()
		var appFS = afero.NewOsFs()
		fi, err := appFS.Stat(ep)
		if err != nil {
			return einfo, errors.New("[EventSrc(event)] os.Stat failed..")
		}
		ein = EventInfo{
			Event{
				Location: ep,
				Op:       e.Op.String(),
			},
			Meta{
				ModTime: fi.ModTime().Truncate(time.Millisecond),
				Mode:    fi.Mode(),
				Name:    fi.Name(),
				Size:    fi.Size(),
			},
		}
		einfo = append(einfo, ein)
		return einfo, eo
	}
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
					esrc := EventSrc(event)
					ev, err := esrc()
					if err != nil {
						log.Fatal(err)
					}

					// Just for local testing
					einfo, _ := json.Marshal(ev)
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
