package main

import (
	"github.com/cosmo-workspace/cosmo/internal/cmd"
	"github.com/cosmo-workspace/cosmo/pkg/cli"
)

var (
	// goreleaser default https://goreleaser.com/customization/builds/
	version = "snapshot"
	commit  = "snapshot"
	date    = "snapshot"
)

func main() {
	cmd.Execute(cli.VersionInfo{Version: version, Commit: commit, Date: date})
}
