package config

import (
	"context"

	"github.com/AmrSaber/redirector/src/lib/logger"
	"github.com/AmrSaber/redirector/src/lib/watchers"
)

func LoadConfig(ctx context.Context, readStdin bool, filePath, url string) *Config {
	if readStdin {
		configs := NewConfig(SOURCE_STDIN, "")

		if err := configs.Load(); err != nil {
			logger.Err.Fatal("Could not load config: ", err)
		}

		return configs
	}

	if filePath != "" {

		configs := NewConfig(SOURCE_FILE, filePath)

		if err := configs.Load(); err != nil {
			logger.Err.Fatal("Could not load config: ", err)
		}

		// Watch config file for updates
		updatesChan, err := watchers.WatchConfigFile(ctx, filePath)
		if err != nil {
			logger.Err.Println("Could not watch config file: ", err)
			return configs
		}

		// Update config on file change
		go func() {
			for range updatesChan {
				if err := configs.Load(); err != nil {
					logger.Err.Fatal("Config file changed, could not load new config: ", err)
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
			logger.Err.Fatal("Could not load config file: ", err)
		}

		return configs
	}

	return nil
}
