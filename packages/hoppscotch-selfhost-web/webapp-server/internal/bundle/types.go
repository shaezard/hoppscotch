package bundle

import "time"

const (
	Version        = "2025.9.1"
	DefaultMaxSize = 100 * 1024 * 1024
)

type FileEntry struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Hash string `json:"hash"`
}

type Manifest struct {
	Files []FileEntry `json:"files"`
}

type Metadata struct {
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Signature string    `json:"signature"`
	Manifest  Manifest  `json:"manifest"`
}

type Bundle struct {
	Metadata Metadata
	Content  []byte
}
