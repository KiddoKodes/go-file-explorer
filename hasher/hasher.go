package hasher

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
)

type FileHash struct {
    NameHash string
    PathHash string
}

func HashFile(path string) FileHash {
    name := filepath.Base(path)
    nameHash := sha256.Sum256([]byte(name))
    pathHash := sha256.Sum256([]byte(path))

    return FileHash{
        NameHash: hex.EncodeToString(nameHash[:]),
        PathHash: hex.EncodeToString(pathHash[:]),
    }
}