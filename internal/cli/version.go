package cli

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// SetVersion sets version information from main
func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display version, build information, and VCS details.`,
	Run: func(cmd *cobra.Command, args []string) {
		printVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func printVersion() {
	fmt.Printf("TokenScout %s\n", version)
	fmt.Printf("  Commit:      %s\n", commit)
	fmt.Printf("  Built:       %s\n", date)
	fmt.Printf("  Go version:  %s\n", runtime.Version())
	fmt.Printf("  OS/Arch:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
	
	// Try to get VCS info from build metadata
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if commit == "none" {
					fmt.Printf("  VCS commit:  %s\n", setting.Value)
				}
			case "vcs.time":
				if date == "unknown" {
					fmt.Printf("  VCS time:    %s\n", setting.Value)
				}
			case "vcs.modified":
				if setting.Value == "true" {
					fmt.Printf("  VCS dirty:   yes\n")
				}
			}
		}
	}
}
