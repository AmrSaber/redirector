package watchers

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/AmrSaber/redirector/src/config"
	"github.com/fsnotify/fsnotify"
)

func WatchConfigFile(ctx context.Context, filePath string, configs *config.Config) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op == fsnotify.Write {
					log.Println("Config file changed, reloading...")

					yamlFile, err := ioutil.ReadFile(filePath)
					if err != nil {
						log.Println("Could not read config file: ", err)
						continue
					}

					newConfigs, err := config.ConfigFromYaml(yamlFile)
					if err != nil {
						log.Println("Could not parse config file: ", err)
						continue
					}

					configs.CopyFrom(newConfigs)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.Println("file watcher error:", err)

			case <-ctx.Done():
				watcher.Close()
				return
			}
		}
	}()

	err = watcher.Add(filePath)
	if err != nil {
		watcher.Close()
		return err
	}

	return nil
}
