package main

import (
	"bufio"
	"file-system-index/indexer"
	"file-system-index/searcher"
	"file-system-index/syncer"
	"fmt"
	"log"
	"os"
	"strings"
)

const searchResultLimit = 15

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go <root_directory>")
        return
    }

    rootDir := os.Args[1]
    idx := indexer.NewIndexer(rootDir)
    
    // Start indexing and show progress
    idx.StartIndexing()
    go showProgress(idx.ProgressChan)

    // Wait for indexing to complete
    idx.WaitForIndexing()

    // Display index size
    fmt.Printf("\nIndex size: %s\n", idx.GetIndexSize())

    sync, err := syncer.NewSyncer(idx)
    if err != nil {
        log.Printf("Warning: Failed to create syncer: %v", err)
        log.Println("Continuing without real-time syncing.")
    } else {
        err = sync.Start()
        if err != nil {
            log.Printf("Warning: Failed to start syncer: %v", err)
            log.Println("Continuing without real-time syncing.")
        }
    }

    search := searcher.NewSearcher(idx)

    // Visualize the directory (one level only)
    fmt.Println("\nDirectory Structure (One Level):")
    visualizeOneLevel(rootDir)

    // Interactive search
    reader := bufio.NewReader(os.Stdin)
    for {
        fmt.Print("\nEnter a file name to search (or 'quit' to exit): ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        if input == "quit" {
            break
        }

        results := search.SearchByName(input, searchResultLimit)
        if len(results) == 0 {
            fmt.Println("No files found.")
        } else {
            fmt.Printf("Found %d file(s):\n", len(results))
            for _, result := range results {
                fmt.Printf("- %s (Score: %d)\n  Full path: %s\n", result.HighlightedName, result.Score, result.FileInfo.Path)
            }
        }
    }
}


func showProgress(progressChan <-chan float64) {
    for progress := range progressChan {
        fmt.Printf("\r%s", strings.Repeat(" ", 50))  // Clear previous line
        fmt.Printf("\rIndexing progress: %.2f%%", progress)
    }
    fmt.Println("\nIndexing complete!")
}

func visualizeOneLevel(dir string) {
    files, err := os.ReadDir(dir)
    if err != nil {
        fmt.Printf("Error reading directory %s: %v\n", dir, err)
        return
    }

    fmt.Printf("%s\n", dir)
    for i, file := range files {
        icon := "├── "
        if i == len(files)-1 {
            icon = "└── "
        }
        
        name := file.Name()
        if file.IsDir() {
            name += "/"
        }
        
        fmt.Printf("%s%s\n", icon, name)
    }
}