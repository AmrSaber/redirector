package commands

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/AmrSaber/redirector/src/lib/logger"
	"github.com/AmrSaber/redirector/src/utils"
	"github.com/urfave/cli/v2"
)

var PingCommand = &cli.Command{
	Name:  "ping",
	Usage: "pings the server",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "quiet",
			Aliases: []string{"q"},
			Usage:   "Do not print any thing to stdout or stderr",
		},
	},
	Action: func(c *cli.Context) error {
		logger.ResetLoggersFlags()

		quiet := c.Bool("quiet")
		if quiet {
			logger.Std.SetOutput(io.Discard)
			logger.Err.SetOutput(io.Discard)
		}

		conn, err := net.Dial("unix", utils.SOCKET_PATH)
		if err != nil {
			return fmt.Errorf("server not running")
		}

		defer conn.Close()
		conn.SetDeadline(time.Now().Add(2 * time.Second))

		_, err = conn.Write([]byte(utils.SOCKET_MESSAGE_PING + "\n"))
		if err != nil {
			return fmt.Errorf("error writing to socket: %w", err)
		}

		response, err := io.ReadAll(conn)
		if err != nil {
			return fmt.Errorf("error reading from socket: %w", err)
		}

		if string(response) != "PONG" {
			return fmt.Errorf("unexpected response %q", response)
		}

		if !quiet {
			logger.Std.Println(string(response))
		}

		return nil
	},
}
