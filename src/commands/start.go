package commands

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/AmrSaber/redirector/src/config"
	"github.com/AmrSaber/redirector/src/lib/logger"
	"github.com/AmrSaber/redirector/src/server"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

const URL_ENV_NAME = "CONFIG_URL"

var StartCommand = &cli.Command{
	Name:  "start",
	Usage: "Start redirector server",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "file",
			Usage: "YAML file containing configuration",
		},
		&cli.StringFlag{
			Name:  "url",
			Usage: "URL containing configuration yaml file",
		},
		&cli.BoolFlag{
			Name:  "stdin",
			Usage: "Read configuration from stdin",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Only read config and print results, don't start server",
		},
	},
	Action: func(c *cli.Context) error {
		// Runtime context
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Load env variables if any
		godotenv.Load()

		// Get command flags
		url := c.String("url")
		filePath := c.String("file")
		readStdin := c.Bool("stdin")
		dryRun := c.Bool("dry-run")

		// given flag overwrites env variable
		if urlEnvValue := os.Getenv(URL_ENV_NAME); urlEnvValue != "" {
			if url == "" {
				url = urlEnvValue
			} else {
				logger.Std.Printf("Effect of env variable %s overwritten by provided url flag", URL_ENV_NAME)
			}
		}

		configs := config.LoadConfig(ctx, readStdin, filePath, url)
		if configs == nil {
			logger.Std.Println("No configuration provided!")
			return nil
		}

		logger.Std.Printf("Parsed configurations:\n\n%s\n", configs)

		if dryRun {
			return nil
		}

		server := server.SetupServer(configs)
		done := make(chan error)

		// Start server
		go func() {
			logger.Std.Println("Starting server...")
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				done <- fmt.Errorf("could not start server: %w", err)
			}
		}()

		// Stop server on SIGINT
		go func() {
			select {
			case <-ctx.Done():
				logger.Std.Println("Stopping server...")

				if err := server.Shutdown(context.Background()); err != nil {
					done <- fmt.Errorf("error stopping server: %w", err)
				} else {
					logger.Std.Println("Server stopped")
					done <- nil
				}

			// Incase server could not start, exit directly
			case <-done:
			}
		}()

		// Listen for interrupts to cancel context
		go func() {
			sigint := make(chan os.Signal, 1)
			signal.Notify(sigint, os.Interrupt)
			<-sigint

			cancel()
		}()

		return <-done
	},
}
