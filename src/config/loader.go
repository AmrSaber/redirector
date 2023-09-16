package config

import (
	"context"

	"github.com/AmrSaber/redirector/src/lib/logger"
	"github.com/AmrSaber/redirector/src/lib/watchers"
	"github.com/AmrSaber/redirector/src/models"
)

func CreateConfigManager(ctx context.Context, readStdin bool, filePath, url string) *ConfigManager {
	if readStdin {
		manager := NewConfigManager(models.SOURCE_STDIN, "")

		if err := manager.LoadConfig(); err != nil {
			logger.Err.Fatal("Could not load config: ", err)
		}

		return manager
	}

	if filePath != "" {
		manager := NewConfigManager(models.SOURCE_FILE, filePath)

		if err := manager.LoadConfig(); err != nil {
			logger.Err.Fatal("Could not load config: ", err)
		}

		// Watch config file for updates
		updatesChan, err := watchers.WatchConfigFile(ctx, filePath)
		if err != nil {
			logger.Err.Println("Could not watch config file: ", err)
			return manager
		}

		// Update config on file change
		go func() {
			for range updatesChan {
				if err := manager.LoadConfig(); err != nil {
					logger.Err.Fatal("Config file changed, could not load new config: ", err)
				} else {
					logger.Std.Printf(
						"Config file changed; config reloaded. New config:\n\n%s\n",
						manager.GetStringConfig(),
					)
				}
			}
		}()

		return manager
	}

	if url != "" {
		manager := NewConfigManager(models.SOURCE_URL, url)

		if err := manager.LoadConfig(); err != nil {
			logger.Err.Fatal("Could not load config file: ", err)
		}

		return manager
	}

	return nil
}
