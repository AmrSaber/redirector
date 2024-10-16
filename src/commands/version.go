package commands

import (
	"github.com/AmrSaber/redirector/src/lib/logger"
	"github.com/AmrSaber/redirector/src/utils"
	"github.com/urfave/cli/v2"
)

var VersionCommand = &cli.Command{
	Name:  "version",
	Usage: "print redirector version",
	Action: func(c *cli.Context) error {
		logger.ResetLoggersFlags()

		version := utils.GetVersion()
		if version == "" {
			version = "??"
		}

		logger.Std.Println(version)

		return nil
	},
}
