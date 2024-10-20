package indexer

import (
	"file-system-index/btree"
	indexer_struct "file-system-index/indexer-structs"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Indexer struct {
    Tree         *btree.BTree
    Root         string
    mutex        sync.Mutex
    totalSize    int64
    indexedSize  int64
    ProgressChan chan float64
    doneChan     chan bool
}

func (idx *Indexer) GetIndexSize() string {
    size := idx.Tree.EstimateSize()
    return formatSize(size)
}

func formatSize(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func NewIndexer(root string) *Indexer {
    return &Indexer{
        Tree:         btree.NewBTree(),
        Root:         root,
        ProgressChan: make(chan float64),
        doneChan:     make(chan bool),
    }
}

func (idx *Indexer) StartIndexing() {
    go idx.countFiles()
}

func (idx *Indexer) countFiles() {
    idx.totalSize = 0
    err := filepath.Walk(idx.Root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            if os.IsPermission(err) {
                log.Printf("Permission denied: %v\n", path)
                return filepath.SkipDir
            }
            log.Printf("Error accessing %s: %v\n", path, err)
            return nil
        }
        if !info.IsDir() {
            idx.totalSize += info.Size()
        }
        return nil
    })

    if err != nil {
        log.Printf("Error walking file path: %v\n", err)
    }

    if idx.totalSize == 0 {
        log.Println("No files to index")
        close(idx.ProgressChan)
        idx.doneChan <- true
    } else {
        log.Printf("Total size to index: %d bytes", idx.totalSize)
        go idx.walkAndIndexFiles()
    }
}

func (idx *Indexer) walkAndIndexFiles() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    lastReportedProgress := 0.0

    err := filepath.Walk(idx.Root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            if os.IsPermission(err) {
                log.Printf("Skipping inaccessible directory: %v\n", path)
                return filepath.SkipDir
            }
            log.Printf("Error accessing %s: %v\n", path, err)
            return nil
        }
        if !info.IsDir() {
            fileInfo := indexer_struct.NewFileInfo(path, info)
            idx.mutex.Lock()
            idx.Tree.Insert(fileInfo)
            idx.indexedSize += fileInfo.Size
            progress := float64(idx.indexedSize) / float64(idx.totalSize) * 100
            idx.mutex.Unlock()

            if progress - lastReportedProgress >= 1.0 || progress >= 99.9 {
                idx.ProgressChan <- progress
                lastReportedProgress = progress
            }
        }
        return nil
    })

    if err != nil {
        log.Printf("Error indexing files: %v\n", err)
    }

    // Ensure 100% progress is sent
    idx.ProgressChan <- 100.0
    close(idx.ProgressChan)
    idx.doneChan <- true
}

func (idx *Indexer) WaitForIndexing() {
    <-idx.doneChan
}

func (idx *Indexer) Search(hash string) *indexer_struct.FileInfo {
    idx.mutex.Lock()
    defer idx.mutex.Unlock()
    return idx.Tree.Search(hash)
}

func (idx *Indexer) AddFile(path string) error {
    info, err := os.Stat(path)
    if err != nil {
        return err
    }
    idx.mutex.Lock()
    defer idx.mutex.Unlock()
    idx.Tree.Insert(indexer_struct.NewFileInfo(path, info))
    return nil
}

func (idx *Indexer) RemoveFile(path string) {
    // Implement file removal logic
}

func (idx *Indexer) UpdateFile(path string) error {
    return idx.AddFile(path) // For simplicity, just re-add the file
}

