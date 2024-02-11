package radiochatter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// BlobStorage is a content-addressable storage layer.
//
// All methods should be goroutine-safe.
type BlobStorage interface {
	// Link creates a link that can be used to download an item.
	Link(ctx context.Context, key BlobKey) (*url.URL, error)

	// Store will upload a blob to blob storage, returning the key it is stored
	// under.
	Store(ctx context.Context, blob []byte) (BlobKey, error)
}

type BlobKey [sha256.Size]byte

func (b BlobKey) String() string {
	return hex.EncodeToString(b[:])
}

func blobKeyForBytes(data []byte) BlobKey {
	return sha256.Sum256(data)
}

func NewOnDiskStorage(logger *zap.Logger, rootDir string) BlobStorage {
	return &onDiskStorage{RootDir: rootDir, Logger: logger}
}

type onDiskStorage struct {
	RootDir string
	Logger  *zap.Logger
}

func (s *onDiskStorage) path(item BlobKey) string {
	return path.Join(s.RootDir, item.String())
}

func (s *onDiskStorage) Link(ctx context.Context, key BlobKey) (*url.URL, error) {
	filename := s.path(key)

	if _, err := os.Stat(filename); err != nil {
		return nil, fmt.Errorf("unable to find %q: %w", filename, err)
	}

	return filePathToURL(filename)
}

func (s *onDiskStorage) Store(ctx context.Context, blob []byte) (BlobKey, error) {
	key := blobKeyForBytes(blob)
	filename := s.path(key)

	parent := path.Dir(filename)
	if err := os.MkdirAll(parent, 0766); err != nil {
		return BlobKey{}, fmt.Errorf("unable to create %s/: %w", s.RootDir, err)
	}

	if _, err := os.Stat(filename); err == nil {
		s.Logger.Debug(
			"Already exists",
			zap.String("filename", filename),
			zap.Stringer("key", key),
		)

		return key, nil
	}

	s.Logger.Debug(
		"Saving blob",
		zap.String("filename", filename),
		zap.Stringer("key", key),
	)
	if err := os.WriteFile(filename, blob, 0766); err != nil {
		return BlobKey{}, fmt.Errorf("unable to save to %s: %w", filename, err)
	}

	return key, nil
}

func filePathToURL(filePath string) (*url.URL, error) {
	// Clean the file path to resolve any ".." or "." elements and convert to absolute path
	absPath, err := filepath.Abs(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}

	// Convert backslashes to slashes for URL compatibility
	// This is crucial for Windows paths
	absPath = strings.Replace(absPath, "\\", "/", -1)

	// Ensure Windows drive letters are correctly formatted for URLs
	if len(absPath) > 0 && absPath[1] == ':' {
		// Upper-case drive letter and prepend a slash
		absPath = "/" + strings.ToUpper(absPath[0:1]) + ":" + absPath[2:]
	}

	// Create a file URL
	u := &url.URL{
		Scheme: "file",
		Path:   absPath,
	}

	return u, nil
}
