package bundle

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/zstd"
	"github.com/zeebo/blake3"
)

type Builder struct{}

func NewBuilder() (*Builder, error) {
	return &Builder{}, nil
}

func (b *Builder) Close() error {
	return nil
}

type zstdCompressor struct{}

func (z zstdCompressor) Compress(w io.Writer, r io.Reader) error {
	encoder, err := zstd.NewWriter(w)
	if err != nil {
		return err
	}
	defer encoder.Close()

	_, err = io.Copy(encoder, r)
	return err
}

func init() {
	zip.RegisterCompressor(93, func(w io.Writer) (io.WriteCloser, error) {
		return zstd.NewWriter(w)
	})
}

// Build creates a bundle by packaging frontend files into a ZIP archive with zstd compression.
//
// Compression Strategy:
// Files are added to the ZIP archive with zstd compression applied at the ZIP level,
// matching the Rust implementation's approach using CompressionMethod::Zstd.
func (b *Builder) Build(frontendPath string) ([]byte, []FileEntry, error) {
	if _, err := os.Stat(frontendPath); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("frontend path does not exist: %s", frontendPath)
	}

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	var files []FileEntry

	err := filepath.Walk(frontendPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		relPath, err := filepath.Rel(frontendPath, path)
		if err != nil {
			return err
		}

		header := &zip.FileHeader{
			Name:   filepath.ToSlash(relPath),
			Method: 93,
		}
		header.SetMode(0644)

		log.Printf("ZIP entry name: %s", header.Name)

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if _, err := writer.Write(content); err != nil {
			log.Printf("ERROR writing file %s: %v", relPath, err)
			return fmt.Errorf("failed to write file %s: %w", relPath, err)
		}

		log.Printf("Successfully wrote file %s (%d bytes) to ZIP", relPath, len(content))

		hasher := blake3.New()
		hasher.Write(content)
		hash := hasher.Sum(nil)

		files = append(files, FileEntry{
			Path: filepath.ToSlash(relPath),
			Size: info.Size(),
			Hash: base64.StdEncoding.EncodeToString(hash),
		})

		log.Printf("Added to manifest: %s", filepath.ToSlash(relPath))

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, nil, err
	}

	log.Printf("Built bundle with %d files", len(files))
	return buf.Bytes(), files, nil
}
