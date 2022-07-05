package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/AmrSaber/redirector/src/config"
	"github.com/AmrSaber/redirector/src/server"
	"github.com/joho/godotenv"
)

// TODO:
// 2. Cache config file from URL and invalidate the cache after some time (adjustable from config)
// 3. Add tests to github actions workflow
// 4. Dockerize the app
// 5. Add github actions workflow to auto-publish docker image on github

const URL_ENV_NAME = "CONFIG_URL"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	godotenv.Load()

	var filePath string
	var url string

	flag.StringVar(&filePath, "file", "", "YAML file containing configuration")
	flag.StringVar(&url, "url", "", "URL containing configuration yaml file")
	flag.Parse()

	// given flag overwrites env variable
	urlEnvValue := os.Getenv(URL_ENV_NAME)
	if urlEnvValue != "" {
		if url == "" {
			url = urlEnvValue
		} else {
			log.Printf("Env variable %s overwritten by provided url flag", URL_ENV_NAME)
		}
	}

	configs := config.LoadConfig(ctx, filePath, url)
	if configs == nil {
		log.Fatal("No configuration provided!")
	}

	server := server.SetupServer(ctx, configs)

	wg := sync.WaitGroup{}
	wg.Add(2)

	// Start server
	go func() {
		defer wg.Done()

		fmt.Println("Starting server...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Could not start server: ", err)
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
