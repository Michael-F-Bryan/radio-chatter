package blob

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/url"
	"time"
)

var ErrNotFound = errors.New("blob not found")

// Storage is a content-addressable storage layer.
//
// All methods should be goroutine-safe.
type Storage interface {
	io.Closer

	// Link creates a link that can be used to download an item. The link should
	// be valid for at least the specified amount of time.
	Link(ctx context.Context, key Key, validFor time.Duration) (*url.URL, error)

	// Store will upload a blob to blob storage, returning the key it is stored
	// under.
	Store(ctx context.Context, blob []byte) (Key, error)
}

type Key [sha256.Size]byte

func (b Key) String() string {
	return hex.EncodeToString(b[:])
}

func ParseKey(key string) (Key, error) {
	var buffer [sha256.Size]byte

	bytesDecoded, err := hex.Decode(buffer[:], []byte(key))
	if err != nil {
		return Key{}, err
	}
	if bytesDecoded != len(buffer) {
		return Key{}, errors.New("incorrect blob key size")
	}

	return Key(buffer), nil
}

func KeyForBytes(data []byte) Key {
	return sha256.Sum256(data)
}
