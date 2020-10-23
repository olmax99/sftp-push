package cmd

import (
	"errors"
	log1 "log" // Use built-in log prior to logrus
	"strings"

	config "github.com/olmax99/sftppush/internal/config"
	log "github.com/olmax99/sftppush/internal/log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	gCfg    watchConfig // global watchConfig accessed by watch.go and event/watcher.go
	gL      *logrus.Logger
	msg     string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sftppush watch",
	Short: "Use the watch command to start the fsnotify file watcher.",
	Long: strings.TrimSpace(`
..:: WELCOME ::..
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
____

The sftppush project provides a mini pipeline for 

..:: upload --> file event --> decompress --> s3 archive ::..

Most likely you want to run this project inside an Sftp server, that 
receives a constant stream of data files.

The sftppush project is intended to run in a Linux (Ubuntu/Debian) VM.
It captures WRITE_CLOSE events for files on the file system based on a 
single or multiple source directories.

The watch --source flag can read single and multiple directories. 
However, it is recommended to use a configuration file. In case of 
multiple directory targets there will be a separate go watch process 
spawned for each target directory, respectively.
____

Config:

defaults:
  userpath: # Set by default to /home/
  s3target: test-bucket
  awsprofile: my-profile
  awsregion: my-region
watch:
  source:
    - name: user1
      paths:
        - /path/to/source/directory1
        - /path/to/source/directory2

Examples:

SFTPPUSH_DEFAULTS_USERPATH=/my/user/dir/ sftppush --config config.yaml watch

SFTPPUSH_DEFAULTS_AWSPROFILE=my-profile sftppush watch \
  --source="name=user1,paths=/device1/data /device2/data" \
  --source="name=user2,paths=/device1/data /device2/data"
`),
	// Uncomment the following line if your bare application
	// has an action associated with it
	// The action can also be to just load the initConfig
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
		}
	},
}

// Execute is the Cobra cli entrypoint and adds all child commands to the root
// command and sets flags appropriately. This is called by main.main(). It only
// needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		gL.Fatalf("%s", err)
	}
}

//!+ viper, config

func initConfig() {
	v := config.ReadConfig("SFTPPUSH", cfgFile)
	if err := v.Unmarshal(&gCfg); err != nil {
		log1.Fatalf("%s", errors.New("unmarshal"))
	}

	// initialize custom logger
	gL, msg = log.NewLogger(v)
	gL.Infof("Initialize log: %s", msg)
}

func init() {
	// This is only triggered when 'Run: func(..)' is active in rootCmd
	cobra.OnInitialize(initConfig)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to config file (default is $HOME/.sftppush/config.yaml)")
	// rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	// viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
}

//!- viper, config
