package cmd

import (
	internal "github.com/olmax99/sftppush/internal/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of generated code example",
	Long:  `All software has versions. This is generated code example`,
	Run: func(cmd *cobra.Command, args []string) {
		gL.Infoln("Build Date:", internal.BuildDate)
		gL.Infoln("Git Commit:", internal.GitCommit)
		gL.Infoln("Version:", internal.Version)
		gL.Infoln("Go Version:", internal.GoVersion)
		gL.Infoln("OS / Arch:", internal.OsArch)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
