package watchers

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/fsnotify/fsnotify"
)

func WatchConfigFile(ctx context.Context, filePath string) (chan []byte, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fileChan := make(chan []byte)

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

					fileChan <- yamlFile
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.Println("file watcher error:", err)

			case <-ctx.Done():
				log.Println("Stopping file watcher...")
				close(fileChan)
				watcher.Close()
				return
			}
		}
	}()

	err = watcher.Add(filePath)
	if err != nil {
		watcher.Close()
		return nil, err
	}

	return fileChan, nil
}
