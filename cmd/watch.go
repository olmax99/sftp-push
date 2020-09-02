package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	// "github.com/fsnotify/fsnotify"
)

var Target string

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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("I AM WATCHING: %v !! \n", Target)
		// TODO Create a watcher for every target in target file
		// watcher, err := fsnotify.NewWatcher()
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// defer watcher.Close()

		// done := make(chan bool)
		// go func() {
		// 	for {
		// 		select {
		// 		case event, ok := <-watcher.Events:
		// 			if !ok {
		// 				return
		// 			}
		// 			log.Println("event:", event)
		// 			if event.Op&fsnotify.Write == fsnotify.Write {
		// 				log.Println("modified file:", event.Name)
		// 			}
		// 		case err, ok := <-watcher.Errors:
		// 			if !ok {
		// 				return
		// 			}
		// 			log.Println("error:", err)
		// 		}
		// 	}
		// }()

		// err = watcher.Watch("/tmp/foo")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// <-done
	},
}

func init() {
	rootCmd.AddCommand(cmdWatch)
	// cmdWatch.AddCommand(cmdTarget)
	cmdWatch.Flags().StringVarP(&Target, "target", "t", "", "Target directory to watch (required)")
	cmdWatch.MarkFlagRequired("target")
}
