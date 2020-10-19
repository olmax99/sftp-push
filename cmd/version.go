package cmd

import (
	"github.com/olmax99/sftppush/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of generated code example",
	Long:  `All software has versions. This is generated code example`,
	Run: func(cmd *cobra.Command, args []string) {
		gL.Println("Build Date:", version.BuildDate)
		gL.Println("Git Commit:", version.GitCommit)
		gL.Println("Version:", version.Version)
		gL.Println("Go Version:", version.GoVersion)
		gL.Println("OS / Arch:", version.OsArch)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
