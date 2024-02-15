// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"time"
)

// Properties shared by all items stored in the database.
type Node interface {
	IsNode()
	// A unique ID for this item.
	GetID() string
	// When the item was created.
	GetCreatedAt() time.Time
	// When the item was last updated.
	GetUpdatedAt() time.Time
}

type Chunk struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	// When the chunk was first broadcast.
	Timestamp time.Time `json:"timestamp"`
	// A SHA-256 checksum of the chunk's audio file.
	Sha256 string `json:"sha256"`
	// Where the chunk's audio file can be downloaded from.
	DownloadURL *string `json:"downloadUrl,omitempty"`
	// Iterate over the radio messages detected in the chunk.
	Transmissions *TransmissionsConnection `json:"transmissions"`
}

func (Chunk) IsNode() {}

// A unique ID for this item.
func (this Chunk) GetID() string { return this.ID }

// When the item was created.
func (this Chunk) GetCreatedAt() time.Time { return this.CreatedAt }

// When the item was last updated.
func (this Chunk) GetUpdatedAt() time.Time { return this.UpdatedAt }

type ChunksConnection struct {
	Edges    []Chunk   `json:"edges,omitempty"`
	PageInfo *PageInfo `json:"pageInfo"`
}

type Mutation struct {
}

// Information about a page in a paginated query.
type PageInfo struct {
	// Are there any more pages?
	HasNextPage bool `json:"hasNextPage"`
	// How many items were in this page?
	Length int `json:"length"`
	// A cursor that can be passed as an "after" parameter to read the next page.
	EndCursor *string `json:"endCursor,omitempty"`
}

type Query struct {
}

type RegisterStreamVariables struct {
	DisplayName string `json:"displayName"`
	URL         string `json:"url"`
}

// A stream to monitor and extract transmissions from.
type Stream struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	// The human-friendly name for this stream.
	DisplayName string `json:"displayName"`
	// Where the stream can be downloaded from.
	//
	// This is typically a URL, but can technically be anything ffmpeg allows as an
	// input.
	URL string `json:"url"`
	// Iterate over the raw chunks of audio downloaded for this stream.
	Chunks *ChunksConnection `json:"chunks"`
	// Iterate over the radio messages detected in the stream.
	Transmissions *TransmissionsConnection `json:"transmissions"`
}

func (Stream) IsNode() {}

// A unique ID for this item.
func (this Stream) GetID() string { return this.ID }

// When the item was created.
func (this Stream) GetCreatedAt() time.Time { return this.CreatedAt }

// When the item was last updated.
func (this Stream) GetUpdatedAt() time.Time { return this.UpdatedAt }

type StreamsConnection struct {
	Edges    []Stream  `json:"edges,omitempty"`
	PageInfo *PageInfo `json:"pageInfo"`
}

type Subscription struct {
}

type Transcription struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Content   string    `json:"content"`
}

func (Transcription) IsNode() {}

// A unique ID for this item.
func (this Transcription) GetID() string { return this.ID }

// When the item was created.
func (this Transcription) GetCreatedAt() time.Time { return this.CreatedAt }

// When the item was last updated.
func (this Transcription) GetUpdatedAt() time.Time { return this.UpdatedAt }

// A radio transmission.
type Transmission struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	// When the transmission was first broadcast.
	Timestamp time.Time `json:"timestamp"`
	// How long is the transmission, in seconds?
	Length float64 `json:"length"`
	// A SHA-256 checksum of the chunk's audio file.
	Sha256 string `json:"sha256"`
	// Where the chunk's audio file can be downloaded from.
	DownloadURL   *string        `json:"downloadUrl,omitempty"`
	Transcription *Transcription `json:"transcription,omitempty"`
}

func (Transmission) IsNode() {}

// A unique ID for this item.
func (this Transmission) GetID() string { return this.ID }

// When the item was created.
func (this Transmission) GetCreatedAt() time.Time { return this.CreatedAt }

// When the item was last updated.
func (this Transmission) GetUpdatedAt() time.Time { return this.UpdatedAt }

type TransmissionsConnection struct {
	Edges    []Transmission `json:"edges,omitempty"`
	PageInfo *PageInfo      `json:"pageInfo"`
}
