package main

import (
	_ "embed"

	"github.com/speier/tokenscout/internal/cli"
	"github.com/speier/tokenscout/internal/config"
)

//go:embed config.yaml
var defaultConfigYAML string

//go:embed .env.example
var defaultEnvTemplate string

// Version information (set via ldflags during build)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Set embedded templates
	config.SetEmbeddedTemplates(defaultConfigYAML, defaultEnvTemplate)

	cli.SetVersion(version, commit, date)
	cli.Execute()
}
