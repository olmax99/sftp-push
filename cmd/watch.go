package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var Target string

type EventInfo struct {
	// event string
	Meta map[string]interface{} `json:"meta"`
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
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					eventpath := getEventPath(event.Name)
					fi, err := os.Stat(eventpath)
					if err != nil {
						fmt.Println(err)
						return
					}
					log.Printf("modified file in path: %s, %TT", eventpath, eventpath)
					emeta := map[string]interface{}{
						"name":    fi.Name(),
						"size":    fi.Size(),
						"mode":    fi.Mode(),
						"modTime": fi.ModTime(),
					}

					einfo := []EventInfo{}
					einfo = append(einfo, EventInfo{Meta: emeta})

					einfol, _ := json.Marshal(einfo)
					fmt.Println(string(einfol))
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
