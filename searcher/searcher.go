package searcher

import (
	"file-system-index/hasher"
	"file-system-index/indexer"
	indexer_struct "file-system-index/indexer-structs"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Searcher struct {
    Indexer      *indexer.Indexer
    fileNameMap  map[string][]indexer_struct.FileInfo
    mapMutex     sync.RWMutex
}

type SearchResult struct {
    FileInfo indexer_struct.FileInfo
    HighlightedName string
    Score int
}

func NewSearcher(idx *indexer.Indexer) *Searcher {
    s := &Searcher{
        Indexer:     idx,
        fileNameMap: make(map[string][]indexer_struct.FileInfo),
    }
    s.buildFileNameMap()
    return s
}

func (s *Searcher) buildFileNameMap() {
    s.mapMutex.Lock()
    defer s.mapMutex.Unlock()

    s.Indexer.Tree.InOrderTraversal(func(info indexer_struct.FileInfo) bool {
        fileName := strings.ToLower(filepath.Base(info.Path))
        s.fileNameMap[fileName] = append(s.fileNameMap[fileName], info)
        return true
    })
}

func (s *Searcher) SearchByName(name string, limit int) []SearchResult {
    s.mapMutex.RLock()
    defer s.mapMutex.RUnlock()

    var results []SearchResult
    lowerName := strings.ToLower(name)

    for fileName, fileInfos := range s.fileNameMap {
        if fuzzy.Match(lowerName, fileName) {
            distance := fuzzy.LevenshteinDistance(lowerName, fileName)
            score := len(fileName) - distance
            
            // Boost score for exact substring matches
            if strings.Contains(fileName, lowerName) {
                score += len(lowerName) * 2
            }

            for _, fileInfo := range fileInfos {
                highlightedName := highlightMatch(filepath.Base(fileInfo.Path), name)
                results = append(results, SearchResult{
                    FileInfo: fileInfo,
                    HighlightedName: highlightedName,
                    Score: score,
                })
            }
        }
    }

    // Sort results by score (highest first)
    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })

    // Limit the number of results
    if len(results) > limit {
        results = results[:limit]
    }

    return results
}

func (s *Searcher) SearchByPath(path string) *indexer_struct.FileInfo {
    pathHash := hasher.HashFile(path).PathHash
    return s.Indexer.Search(pathHash)
}

func highlightMatch(fileName, searchTerm string) string {
    lowerFileName := strings.ToLower(fileName)
    lowerSearchTerm := strings.ToLower(searchTerm)
    
    var result strings.Builder
    lastIndex := 0

    for _, word := range strings.Fields(lowerSearchTerm) {
        index := strings.Index(lowerFileName[lastIndex:], word)
        if index != -1 {
            index += lastIndex
            result.WriteString(fileName[lastIndex:index])
            result.WriteString("\033[1;31m") // Start red color
            result.WriteString(fileName[index : index+len(word)])
            result.WriteString("\033[0m") // Reset color
            lastIndex = index + len(word)
        }
    }

    result.WriteString(fileName[lastIndex:])
    return result.String()
}