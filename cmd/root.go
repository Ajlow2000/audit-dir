package cmd

import (
	"os"
	"strings"

	"github.com/Ajlow2000/audit-dir/app"
	"github.com/spf13/cobra"
)

var (
    debug = false
    target = ""
    logFile = ""
    ignore = ""
)


var rootCmd = &cobra.Command{
	Use:   "audit-dir",
	Short: "Generate a report of the status of all git repositories within the specified [dir]",
    Long: "Generate a report of the status of all git repositories within the specified [dir]\n\nLogging is disabled until a log file is provided via --log-file",
    Version: "", 
    Run: func(cmd *cobra.Command, args []string) {
        ignoreList := strings.Split(ignore, ":")
        app.Main(target, ignoreList)
    },
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
    rootCmd.Flags().StringVarP(&target, "target-dir", "t", "$HOME", "Target directory to search for git repos within. Defaults to $HOME")
    rootCmd.Flags().StringVarP(&ignore, "ignore-dirs", "i", "", "Provide a ':' deliminated list of paths to ignore")
}
