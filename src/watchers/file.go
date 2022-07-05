package watchers

import (
	"context"
	"log"

	"github.com/fsnotify/fsnotify"
)

func WatchConfigFile(ctx context.Context, filePath string) (<-chan any, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	updateChan := make(chan any)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op == fsnotify.Write {
					updateChan <- nil
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				log.Println("file watcher error:", err)

			case <-ctx.Done():
				log.Println("Stopping file watcher...")
				close(updateChan)
				watcher.Close()
				return
			}
		}
	}()

	err = watcher.Add(filePath)
	if err != nil {
		close(updateChan)
		watcher.Close()
		return nil, err
	}

	return updateChan, nil
}
