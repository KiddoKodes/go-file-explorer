package indexer_struct

import (
	"os"
	"time"

	"file-system-index/hasher"
)

type FileInfo struct {
    Path         string
    Size         int64
    ModTime      time.Time
    Hash         hasher.FileHash
    IsDir        bool
}

func NewFileInfo(path string, info os.FileInfo) FileInfo {
    return FileInfo{
        Path:    path,
        Size:    info.Size(),
        ModTime: info.ModTime(),
        Hash:    hasher.HashFile(path),
        IsDir:   info.IsDir(),
    }
}