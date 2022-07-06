package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/AmrSaber/redirector/src/config"
	"github.com/AmrSaber/redirector/src/logger"
	"github.com/AmrSaber/redirector/src/server"
	"github.com/joho/godotenv"
)

const URL_ENV_NAME = "CONFIG_URL"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	godotenv.Load()

	var filePath, url string
	var readStdin, dryRun bool

	flag.BoolVar(&dryRun, "dry-run", false, "Only read config and print results, don't start server")
	flag.BoolVar(&readStdin, "stdin", false, "Read configuration from stdin")
	flag.StringVar(&filePath, "file", "", "YAML file containing configuration")
	flag.StringVar(&url, "url", "", "URL containing configuration yaml file")
	flag.Parse()

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
		logger.Err.Fatal("No configuration provided!")
	}

	logger.Std.Printf("Parsed configurations:\n\n%s\n", configs)

	if dryRun {
		return
	}

	server := server.SetupServer(ctx, configs)

	wg := sync.WaitGroup{}
	wg.Add(2)

	// Start server
	go func() {
		defer wg.Done()

		logger.Std.Println("Starting server...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Err.Fatal("Could not start server: ", err)
		}
	}()

	// Listen for interrupts to cancel context
	go func() {
		defer wg.Done()

		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		cancel()
	}()

	wg.Wait()
}
