package config

import (
	"context"

	"github.com/AmrSaber/redirector/src/logger"
	"github.com/AmrSaber/redirector/src/watchers"
)

func LoadConfig(ctx context.Context, filePath, url string) *Config {
	if filePath != "" {
		configs := NewConfig(SOURCE_FILE, filePath)

		if err := configs.Load(); err != nil {
			logger.Err.Fatal(err)
		}

		// Watch config file for updates
		updatesChan, err := watchers.WatchConfigFile(ctx, filePath)
		if err != nil {
			logger.Err.Fatal("could not watch config file: ", err)
		}

		// Update config on file change
		go func() {
			for range updatesChan {
				if err := configs.Load(); err != nil {
					logger.Err.Fatal("config file changed, could not load new config: ", err)
				} else {
					logger.Std.Printf("Config file changed; config reloaded. New config:\n\n%s\n", configs)
				}
			}
		}()

		return configs
	}

	if url != "" {
		configs := NewConfig(SOURCE_URL, url)

		if err := configs.Load(); err != nil {
			logger.Err.Fatal(err)
		}

		return configs
	}

	return nil
}
