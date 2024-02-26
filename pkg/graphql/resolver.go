package graphql

import (
	"time"

	"github.com/Michael-F-Bryan/radio-chatter/pkg/blob"
	"gorm.io/gorm"
)

type Resolver struct {
	DB      *gorm.DB
	Storage blob.Storage
	// How frequently to poll when updating subscriptions.
	PollInterval time.Duration
}
