package main

import (
	"os"

	"github.com/AmrSaber/redirector/src/commands"
	"github.com/AmrSaber/redirector/src/lib/logger"
	"github.com/AmrSaber/redirector/src/utils"
	"github.com/urfave/cli/v2"
)

var version string

func main() {
	// Set version number if it's loaded
	if version != "" {
		utils.SetVersion(version)
	}

	app := &cli.App{
		Name:  "redirector",
		Usage: "Redirects requests to different servers",

		Commands: []*cli.Command{
			commands.StartCommand,
			commands.PingCommand,
			commands.StopCommand,
			commands.VersionCommand,
		},
	}

	// Run CLI
	if err := app.Run(os.Args); err != nil {
		logger.Err.Fatal(err)
	}
}
