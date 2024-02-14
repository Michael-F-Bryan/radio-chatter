package graphql

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/graphql/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func modelId(value any) string {
	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)

	for _, field := range reflect.VisibleFields(t) {
		if field.Type.AssignableTo(reflect.TypeOf(gorm.Model{})) {
			model := v.FieldByIndex(field.Index).Interface().(gorm.Model)
			id := fmt.Sprintf("%s#%d", t.Name(), model.ID)
			return base64.StdEncoding.EncodeToString([]byte(id))
		}
	}

	panic("The type must embed a gorm.Model")
}

func decodeModelId[T any](encoded string) (uint, error) {
	s, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return 0, err
	}

	pieces := strings.SplitN(string(s), "#", 2)
	if len(pieces) != 2 {
		return 0, errors.New("Invalid ID format")
	}

	typeName := pieces[0]
	rawId := pieces[1]

	ty := typeOf[T]()
	if typeName != ty.Name() {
		return 0, fmt.Errorf("expected a %s, but this ID is for a %s", ty.Name(), typeName)
	}

	id, err := strconv.Atoi(rawId)
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}

func typeOf[T any]() reflect.Type {
	var dummy T
	return reflect.TypeOf(dummy)
}

func streamToGraphQL(t radiochatter.Stream) model.Stream {
	return model.Stream{
		ID:          modelId(t),
		CreatedAt:   t.CreatedAt.UTC(),
		UpdatedAt:   t.UpdatedAt.UTC(),
		DisplayName: t.DisplayName,
		URL:         t.Url,
	}
}
func transmissionToGraphQL(t radiochatter.Transmission) model.Transmission {
	return model.Transmission{
		ID:        modelId(t),
		CreatedAt: t.CreatedAt.UTC(),
		UpdatedAt: t.UpdatedAt.UTC(),
		Timestamp: t.TimeStamp,
		Length:    t.Length.Seconds(),
		Sha256:    t.Sha256,
		Content:   t.Content,
	}
}
func chunkToGraphQL(t radiochatter.Chunk) model.Chunk {
	return model.Chunk{
		ID:        modelId(t),
		CreatedAt: t.CreatedAt.UTC(),
		UpdatedAt: t.UpdatedAt.UTC(),
		Timestamp: t.TimeStamp,
		Sha256:    t.Sha256,
	}
}

func getByID[Model any, Generated any](db *gorm.DB, id string, mapFunc func(Model) Generated) (*Generated, error) {
	realID, err := decodeModelId[Model](id)
	if err != nil {
		return nil, err
	}

	var model Model
	err = db.First(&model, "id = ?", realID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	value := mapFunc(model)

	return &value, nil
}

func pollForUpdates[Model any, Generated any](
	ctx context.Context,
	db *gorm.DB,
	logger *zap.Logger,
	mapFunc func(Model) Generated,
	getCreatedAt func(Model) time.Time,
) <-chan *Generated {
	ch := make(chan *Generated)

	go func() {
		defer close(ch)
		logger.Debug("Subscription started", zap.Stringer("type", typeOf[Model]()))
		defer logger.Debug("Subscription cancelled")

		timer := time.NewTicker(radiochatter.ChunkLength)
		defer timer.Stop()

		lastCheck := time.Now()
		db := db.WithContext(ctx)

		for {
			select {
			case <-timer.C:
				var items []Model
				err := db.Where("createdAt > ?", lastCheck).Find(&items).Error
				if err != nil {
					logger.Error("Unable to fetch recently created items", zap.Error(err))
					return
				}

				for _, item := range items {
					value := mapFunc(item)
					select {
					case ch <- &value:
						lastCheck = getCreatedAt(item)
					case <-ctx.Done():
						return
					}
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

func signedURL(ctx context.Context, logger *zap.Logger, storage radiochatter.BlobStorage, sha256 string) (*string, error) {
	key, err := radiochatter.ParseBlobKey(sha256)
	if err != nil {
		return nil, err
	}

	url, err := storage.Link(ctx, key)
	if errors.Is(err, radiochatter.ErrBlobNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	logger.Debug(
		"Generated a signed URL",
		zap.Stringer("key", key),
		zap.Stringer("url", url),
	)

	u := url.String()

	return &u, nil
}
