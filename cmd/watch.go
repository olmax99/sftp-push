package cmd

import (
	"encoding/json"
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

func getEventPath(target string) (eventpath string) {
	if path.IsAbs(target) {
		return target
	}
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	eventpath = path.Join(pwd, target)
	return
}

func EventSrc(e fsnotify.Event) func() []EventInfo {
	ep := e.Name
	if !path.IsAbs(e.Name) {
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		ep = path.Join(pwd, e.Name)
	}
	return func() (einfo []EventInfo) {
		// TODO Create ENV 'TESTING' and set to afero.NewMemMapFs()
		var appFS = afero.NewOsFs()
		fi, err := appFS.Stat(ep)
		if err != nil {
			log.Fatal(err)
		}
		ei := EventInfo{
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
		einfo = append(einfo, ei)
		return
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

					// Just for local testing
					einfo, _ := json.Marshal(esrc())
					fmt.Printf("einfol: %v, eiT: %T", string(einfo), esrc())
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
