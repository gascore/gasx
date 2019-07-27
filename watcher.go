package gasx

import (
	"os"
	"strings"
	"log"
	"time"
	"github.com/radovskyb/watcher"
)

func StartWatcher(onUpdate func(name string), ignoringExt, watchings, recursiveWatchings []string) error {
	w := watcher.New()

	w.SetMaxEvents(1)
	w.FilterOps(watcher.Rename, watcher.Move, watcher.Write, watcher.Create, watcher.Remove)

	w.AddFilterHook(func(info os.FileInfo, fullPath string) error {
		for _, ext := range ignoringExt {
			if strings.HasSuffix(fullPath, ext) {
				return watcher.ErrSkip
			}
		}
		return nil
	})

	for _, watching := range watchings {
		w.Add(watching)
	}

	for _, watching := range recursiveWatchings {
		w.AddRecursive(watching)
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				if event.IsDir() {
					continue
				}
				onUpdate(event.Path)
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	// Start the watching process - it'll check for changes every 3s.
	return w.Start(time.Millisecond * 3000)
}