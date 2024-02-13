// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"time"
)

type Node interface {
	IsNode()
	GetID() string
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

type Chunk struct {
	ID            string                   `json:"id"`
	CreatedAt     time.Time                `json:"createdAt"`
	UpdatedAt     time.Time                `json:"updatedAt"`
	Timestamp     time.Time                `json:"timestamp"`
	Sha256        string                   `json:"sha256"`
	DownloadURL   *string                  `json:"downloadUrl,omitempty"`
	Transmissions *TransmissionsConnection `json:"transmissions"`
}

func (Chunk) IsNode()                      {}
func (this Chunk) GetID() string           { return this.ID }
func (this Chunk) GetCreatedAt() time.Time { return this.CreatedAt }
func (this Chunk) GetUpdatedAt() time.Time { return this.UpdatedAt }

type ChunksConnection struct {
	Edges    []Chunk   `json:"edges,omitempty"`
	PageInfo *PageInfo `json:"pageInfo"`
}

type PageInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	Length      int     `json:"length"`
	EndCursor   *string `json:"endCursor,omitempty"`
}

type Query struct {
}

type Stream struct {
	ID          string            `json:"id"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	DisplayName string            `json:"displayName"`
	URL         string            `json:"url"`
	Chunks      *ChunksConnection `json:"chunks"`
}

func (Stream) IsNode()                      {}
func (this Stream) GetID() string           { return this.ID }
func (this Stream) GetCreatedAt() time.Time { return this.CreatedAt }
func (this Stream) GetUpdatedAt() time.Time { return this.UpdatedAt }

type StreamsConnection struct {
	Edges    []Stream  `json:"edges,omitempty"`
	PageInfo *PageInfo `json:"pageInfo"`
}

type Transmission struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Timestamp   time.Time `json:"timestamp"`
	Length      float64   `json:"length"`
	Sha256      string    `json:"sha256"`
	DownloadURL *string   `json:"downloadUrl,omitempty"`
	Content     *string   `json:"content,omitempty"`
}

func (Transmission) IsNode()                      {}
func (this Transmission) GetID() string           { return this.ID }
func (this Transmission) GetCreatedAt() time.Time { return this.CreatedAt }
func (this Transmission) GetUpdatedAt() time.Time { return this.UpdatedAt }

type TransmissionsConnection struct {
	Edges    []Transmission `json:"edges,omitempty"`
	PageInfo *PageInfo      `json:"pageInfo"`
}
