package assembler

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// Watch listens for file system changes
func Watch(dir string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					rel, _ := filepath.Rel(dir, event.Name)
					log.Println("modified rel file:", rel)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			case <-done:
				watcher.Close()
			}
		}
	}()

	// starting at the root of the project, walk each file/directory searching for
	// directories
	if err := filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		// since fsnotify can watch all the files in a directory, watchers only need
		// to be added to each nested directory
		if fi.Mode().IsDir() {
			fmt.Println("path=", path)
			return watcher.Add(path)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
