package commands

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"

	"github.com/AmrSaber/redirector/src/config"
	"github.com/AmrSaber/redirector/src/lib/logger"
	"github.com/AmrSaber/redirector/src/servers"
	"github.com/AmrSaber/redirector/src/utils"
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
		logger.Std.Printf("Starting redirector %s\n", utils.GetVersion())

		// Runtime context
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

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

		configManager := config.CreateConfigManager(ctx, readStdin, filePath, url)
		if configManager == nil {
			logger.Std.Println("No configuration provided!")
			return nil
		}

		defer configManager.Close()

		logger.Std.Printf("Parsed configurations:\n\n%s\n", configManager.GetStringConfig())

		if dryRun {
			return nil
		}

		// Listen for interrupts to cancel context
		go func() {
			sigint := make(chan os.Signal, 1)
			signal.Notify(sigint, os.Interrupt)
			<-sigint

			cancel()
		}()

		httpDoneChan := servers.StartHttpServer(ctx, configManager)
		socketDoneChan := servers.StartUnixSocketListener(ctx)

		errs := make([]error, 0, 2)
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			if err := <-httpDoneChan; err != nil {
				errs = append(errs, err)
			}

			cancel()
		}()

		go func() {
			defer wg.Done()

			if err := <-socketDoneChan; err != nil {
				errs = append(errs, err)
			}

			cancel()
		}()

		wg.Wait()

		if len(errs) > 0 {
			return errors.Join(errs...)
		}

		return nil
	},
}
