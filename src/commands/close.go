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

var CloseCommand = &cli.Command{
	Name:  "close",
	Usage: "closes the server",
	Action: func(c *cli.Context) error {
		logger.ResetLoggers()

		conn, err := net.Dial("unix", utils.SOCKET_PATH)
		if err != nil {
			return fmt.Errorf("server not running")
		}

		defer conn.Close()
		conn.SetDeadline(time.Now().Add(2 * time.Second))

		_, err = conn.Write([]byte(utils.SOCKET_MESSAGE_CLOSE + "\n"))
		if err != nil {
			return fmt.Errorf("error writing to socket: %w", err)
		}

		response, err := io.ReadAll(conn)
		if err != nil {
			return fmt.Errorf("error reading from socket: %w", err)
		}

		if string(response) != "OK" {
			return fmt.Errorf("unexpected response %q", response)
		}

		logger.Std.Println(string(response))

		return nil
	},
}
