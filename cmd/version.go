package cmd

import (
	version "github.com/olmax99/sftppush/internal/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of generated code example",
	Long:  `All software has versions. This is generated code example`,
	Run: func(cmd *cobra.Command, args []string) {
		gL.Infoln("Build Date:", version.BuildDate)
		gL.Infoln("Git Commit:", version.GitCommit)
		gL.Infoln("Version:", version.Version)
		gL.Infoln("Go Version:", version.GoVersion)
		gL.Infoln("OS / Arch:", version.OsArch)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
