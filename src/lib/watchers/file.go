package watchers

import (
	"context"

	"github.com/AmrSaber/redirector/src/lib/logger"
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

				logger.Err.Println("file watcher error:", err)

			case <-ctx.Done():
				logger.Std.Println("Stopping file watcher...")
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
