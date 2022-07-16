package main

import (
	"os"

	"github.com/AmrSaber/redirector/src/commands"
	"github.com/AmrSaber/redirector/src/lib/logger"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "redirector",
		Usage: "Redirects requests to different servers",

		Commands: []*cli.Command{
			commands.StartCommand,
		},
	}

	// Run CLI
	if err := app.Run(os.Args); err != nil {
		logger.Err.Fatal(err)
	}
}
