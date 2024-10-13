package servers

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/AmrSaber/redirector/src/lib/logger"
	"github.com/AmrSaber/redirector/src/utils"
)

func StartUnixSocketListener(ctx context.Context) <-chan error {
	ctx, cancel := context.WithCancel(ctx)
	doneChan := make(chan error)

	go func() {
		defer close(doneChan)

		err := os.RemoveAll(utils.SOCKET_PATH)
		if err != nil {
			doneChan <- fmt.Errorf("error clearing socket file: %v", err)
		}

		listener, err := net.Listen("unix", utils.SOCKET_PATH)
		if err != nil {
			doneChan <- fmt.Errorf("error creating socket: %v", err)
		}

		// Close listener on end of context
		go func() {
			<-ctx.Done()

			logger.Std.Println("Closing socket listener...")
			_ = listener.Close()
			logger.Std.Println("Socket listener closed")

			_ = os.Remove(utils.SOCKET_PATH)
		}()

		// Accept and handle connections
		logger.Std.Println("Listening on socket", utils.SOCKET_PATH)
		for {
			conn, err := listener.Accept()

			if errors.Is(err, net.ErrClosed) {
				return
			}

			if err != nil {
				logger.Err.Println("error accepting connection:", err)
				continue
			}

			go func(conn net.Conn) {
				defer conn.Close()

				// Connection will timeout in 2 seconds
				conn.SetDeadline(time.Now().Add(2 * time.Second))

				reader := bufio.NewReader(conn)

				command, err := reader.ReadString('\n')
				if err != nil {
					logger.Err.Println("error reading from socket connection:", err)
					return
				}

				command = strings.TrimSpace(command)
				switch command {
				case utils.SOCKET_MESSAGE_PING:
					conn.Write([]byte("PONG"))

				case utils.SOCKET_MESSAGE_CLOSE:
					conn.Write([]byte("OK"))
					cancel()

				default:
					logger.Err.Printf("unknown socket message: %q\n", command)
				}
			}(conn)
		}
	}()

	return doneChan
}
