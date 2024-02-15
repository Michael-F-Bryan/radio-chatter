package radiochatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestFindUntranscribedTransmissions(t *testing.T) {
	ctx := testContext(t)
	db := testDatabase(ctx, t)
	stream := Stream{DisplayName: "Test", Url: "..."}
	assert.NoError(t, db.Save(&stream).Error)
	chunk := Chunk{StreamID: stream.ID}
	assert.NoError(t, db.Save(&chunk).Error)
	transmission := Transmission{ChunkID: chunk.ID}
	assert.NoError(t, db.Save(&transmission).Error)

	untranscribed, err := untranscribedTransmissions(db)

	assert.NoError(t, err)
	assert.Len(t, untranscribed, 1)
	untranscribed[0].Model = gorm.Model{ID: untranscribed[0].ID}
	transmission.Model = gorm.Model{ID: transmission.ID}
	assert.Equal(t, []Transmission{transmission}, untranscribed)
}
