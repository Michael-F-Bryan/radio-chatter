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
	"github.com/Michael-F-Bryan/radio-chatter/pkg/blob"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/graphql/model"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/middleware"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func modelId(value any) string {
	v := reflect.ValueOf(value)
	t := v.Type()

	if t.Kind() == reflect.Pointer {
		// Automatically dereference the pointer
		if v.IsNil() {
			panic("unable to get a model ID from a nil pointer")
		}

		v = v.Elem()
		t = v.Type()
	}

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
		return 0, errors.New("invalid ID format")
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

func chunkToGraphQL(t radiochatter.Chunk) model.Chunk {
	return model.Chunk{
		ID:        modelId(t),
		CreatedAt: t.CreatedAt.UTC(),
		UpdatedAt: t.UpdatedAt.UTC(),
		Timestamp: t.TimeStamp,
		Sha256:    t.Sha256,
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
	}
}

func transcriptionToGraphQL(t radiochatter.Transcription) model.Transcription {
	return model.Transcription{
		ID:        modelId(t),
		CreatedAt: t.CreatedAt.UTC(),
		UpdatedAt: t.UpdatedAt.UTC(),
		Content:   t.Content,
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

// poller polls the database for newly created entities.
type poller[Model any, Generated any] struct {
	db           *gorm.DB
	mapFunc      func(Model) Generated
	filter       func(db *gorm.DB) *gorm.DB
	getCreatedAt func(*Generated) time.Time
	interval     time.Duration
}

// begin returns a channel that will receive new records as they are created.
//
// Under the hood, this works by periodically polling the database for anything
// where the createdAt field is later than the previous poll time.
func (p poller[Model, Generated]) begin(ctx context.Context) <-chan *Generated {
	ch := make(chan *Generated, 1)
	ty := typeOf[Model]()
	logger := middleware.GetLogger(ctx).Named("subscription").With(zap.Stringer("type", ty))

	interval := p.interval
	if interval == 0 {
		interval = radiochatter.ChunkLength
	}

	go func() {
		defer close(ch)
		logger.Debug("Subscription started")
		defer logger.Debug("Subscription cancelled")
		logger.Warn("SUBSCRIPTION")
		fmt.Printf("Started subscription\n")
		defer fmt.Printf("Ended subscription\n")

		timer := time.NewTicker(interval)
		defer timer.Stop()

		db := p.db.WithContext(ctx)
		lastCheck := time.Now()

		for {
			fmt.Printf("Loop %s %s\n", interval, lastCheck)
			select {
			case <-timer.C:
				fmt.Printf("Timer\n")
				items, err := poll(ctx, logger, db, lastCheck, p.mapFunc, p.filter)
				fmt.Printf("Return %v %e\n", items, err)
				if err != nil {
					logger.Error("Unable to fetch recently created items", zap.Error(err))
					fmt.Printf("*** Error fetching: %e\n", err)
					return
				}

				for _, item := range items {
					select {
					case ch <- item:
						lastCheck = p.getCreatedAt(item)
					case <-ctx.Done():
						return
					}
				}

			case <-ctx.Done():
				fmt.Printf("Cancelled")
				return
			}
		}
	}()

	return ch
}

func poll[Model any, Generated any](
	ctx context.Context,
	logger *zap.Logger,
	db *gorm.DB,
	lastChecked time.Time,
	mapFunc func(Model) Generated,
	filter func(*gorm.DB) *gorm.DB,
) ([]*Generated, error) {
	table := tableName[Model](db)
	fmt.Printf("Polling %s\n", table)

	if filter != nil {
		db = filter(db)
	}

	db = db.Where(table+".created_at >= ?", lastChecked)
	fmt.Printf("%#v\n", db.Statement.Clauses)

	var items []Model

	if err := db.Find(&items).Error; err != nil {
		logger.Fatal("Lookup failed", zap.Error(err))
		return nil, err
	}

	fmt.Printf("(%s %s) %v\n", table, lastChecked.Format(time.RFC3339Nano), items)

	var generated []*Generated

	for _, item := range items {
		gen := mapFunc(item)
		generated = append(generated, &gen)
	}

	return generated, nil
}

func signedURL(ctx context.Context, logger *zap.Logger, storage blob.Storage, sha256 string) (*string, error) {
	key, err := blob.ParseKey(sha256)
	if err != nil {
		return nil, err
	}

	url, err := storage.Link(ctx, key, 1*time.Hour)
	if errors.Is(err, blob.ErrNotFound) {
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

func getParentObject[Child any, Parent any, ParentGraphQL any](
	db *gorm.DB,
	childID string,
	getParentID func(Child) uint,
	mapFunc func(Parent) ParentGraphQL,
) (*ParentGraphQL, error) {
	realId, err := decodeModelId[Child](childID)
	if err != nil {
		return nil, err
	}

	var child Child

	err = db.First(&child, "id = ?", realId).Error
	if err != nil {
		return nil, fmt.Errorf("unable to find the %s with id=%d: %w", typeOf[Child]().String(), realId, err)
	}

	var parent Parent
	parentID := getParentID(child)

	if err := db.Find(&parent, "id = ?", parentID).Error; err != nil {
		return nil, fmt.Errorf("unable to find the %s with id=%d: %w", typeOf[Parent]().String(), parentID, err)
	}

	model := mapFunc(parent)
	return &model, nil
}
