package config

import (
	"context"
	"log"

	"github.com/AmrSaber/redirector/src/watchers"
)

func LoadConfig(ctx context.Context, filePath, url string) *Config {
	if filePath != "" {
		configs := NewConfig(SOURCE_FILE, filePath)

		if err := configs.Load(); err != nil {
			log.Fatal(err)
		}

		// Watch config file for updates
		updatesChan, err := watchers.WatchConfigFile(ctx, filePath)
		if err != nil {
			log.Fatal("could not watch config file: ", err)
		}

		// Update config on file change
		go func() {
			for range updatesChan {
				log.Println("Config file changed, reloading...")
				configs.Load()
			}
		}()

		return configs
	}

	if url != "" {
		configs := NewConfig(SOURCE_URL, url)

		if err := configs.Load(); err != nil {
			log.Fatal(err)
		}

		return configs
	}

	return nil
}
