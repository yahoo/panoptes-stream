package yaml

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

func (y *yaml) watcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					select {
					case y.informer <- struct{}{}:
					default:
					}

					y.logger.Info("watcher", zap.String("name", event.Name))
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(y.filename)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func (y *yaml) Informer() chan struct{} {
	return y.informer
}
