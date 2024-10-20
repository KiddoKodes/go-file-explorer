package syncer

import (
	"file-system-index/indexer"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

type Syncer struct {
    Indexer *indexer.Indexer
    Watcher *fsnotify.Watcher
}

func NewSyncer(idx *indexer.Indexer) (*Syncer, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    return &Syncer{
        Indexer: idx,
        Watcher: watcher,
    }, nil
}

func (s *Syncer) Start() error {
    err := filepath.Walk(s.Indexer.Root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return s.Watcher.Add(path)
        }
        return nil
    })
    if err != nil {
        return err
    }

    go s.watch()
    return nil
}

func (s *Syncer) watch() {
    for {
        select {
        case event, ok := <-s.Watcher.Events:
            if !ok {
                return
            }
            s.handleEvent(event)
        case err, ok := <-s.Watcher.Errors:
            if !ok {
                return
            }
            log.Println("error:", err)
        }
    }
}

func (s *Syncer) handleEvent(event fsnotify.Event) {
    switch {
    case event.Op&fsnotify.Create == fsnotify.Create:
        s.Indexer.AddFile(event.Name)
    case event.Op&fsnotify.Remove == fsnotify.Remove:
        s.Indexer.RemoveFile(event.Name)
    case event.Op&fsnotify.Write == fsnotify.Write:
        s.Indexer.UpdateFile(event.Name)
    }
}