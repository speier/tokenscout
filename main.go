package main

import (
	"github.com/speier/tokenscout/internal/cli"
)

// Version information (set via ldflags during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.SetVersion(version, commit, date)
	cli.Execute()
}
