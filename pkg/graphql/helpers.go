package graphql

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/graphql/model"
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
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		DisplayName: t.DisplayName,
		URL:         t.Url,
	}
}
func transmissionToGraphQL(t radiochatter.Transmission) model.Transmission {
	return model.Transmission{
		ID:        modelId(t),
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		Timestamp: t.TimeStamp,
		Length:    t.Length.Seconds(),
		Sha256:    t.Sha256,
		Content:   t.Content,
	}
}
func chunkToGraphQL(t radiochatter.Chunk) model.Chunk {
	return model.Chunk{
		ID:        modelId(t),
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
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