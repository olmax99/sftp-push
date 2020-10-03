package cmd

import (
	"log"
	"strings"

	"github.com/olmax99/sftppush/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var wC watchConfig // global watchConfigs accessed by watch.go
var v *viper.Viper

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "generated code example",
	Short: "A brief description of your application",
	Long: strings.TrimSpace(`
A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:
 
Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`),
	// Uncomment the following line if your bare application
	// has an action associated with it
	// The action can also be to just load the initConfig
	Run: func(cmd *cobra.Command, args []string) {},
}

// Execute is the Cobra cli entrypoint and adds all child commands to the root
// command and sets flags appropriately. This is called by main.main(). It only
// needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("FATAL[-] rootCmd: %s", err)
	}
}

//!+ viper, config

func initConfig() {
	v := config.ReadConfig("SFTPPUSH", cfgFile)

	if err := v.Unmarshal(&wC); err != nil {
		log.Printf("ERROR[-] initConfig: %s", err)
	}

	log.Printf("DEBUG[*] initConfig: %s", v.GetString("defaults.userpath"))
	//log.Printf("DEBUG[+] Unmarshal: %#v", wC)
}

func init() {
	// This is only triggered when 'Run: func(..)' is active in rootCmd
	cobra.OnInitialize(initConfig)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "path to config file (default is $HOME/.sftppush.yaml)")
	// rootCmd.PersistentFlags().Bool("viper", true, "use Viper for configuration")
	// viper.BindPFlag("useViper", rootCmd.PersistentFlags().Lookup("viper"))
}

//!- viper, config
