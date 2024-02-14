package graphql

import (
	"fmt"

	"github.com/Michael-F-Bryan/radio-chatter/pkg/graphql/model"
	"gorm.io/gorm"
)

type paginator[Model any, GeneratedModel any, Connection any] struct {
	// Convert the model from the database type to its GraphQL counterpart.
	mapModel func(model Model) GeneratedModel
	// Use the GraphQL model and page info to create a Connection object.
	makeConn func(edges []GeneratedModel, page model.PageInfo) Connection
	// A simple filter which looks for items where the model's field are set to
	// a particular value.
	Filter *Model
	// A callback fired just before executing the query, typically used for more
	// complex filtering.
	BeforeQuery func(db *gorm.DB) *gorm.DB
	// The maximum number of results this query is allowed to return.
	Limit int
}

func (p paginator[Model, GeneratedModel, Connection]) Page(db *gorm.DB, after *string, count int) (*Connection, error) {
	var items []Model

	var dummy Model
	db = db.Model(&dummy)

	if p.Limit > 0 && p.Limit < count {
		// Make sure users can't ask for too many items at once
		count = p.Limit
	}
	db = db.Limit(count)

	if after != nil {
		id, err := decodeModelId[Model](*after)
		if err != nil {
			return nil, fmt.Errorf("invalid ID: %w", err)
		}
		// Note: We explicitly use the table name (e.g. "transmissions.id") so
		// later steps (e.g. a join) can't introduce ambiguities.
		db = db.Where(tableName[Model](db)+".id > ?", id)
	}

	if p.Filter != nil {
		db = db.Where(p.Filter)
	}

	if p.BeforeQuery != nil {
		db = p.BeforeQuery(db)
	}

	if err := db.Find(&items).Error; err != nil {
		return nil, err
	}

	info := model.PageInfo{
		HasNextPage: len(items) >= count,
		Length:      len(items),
	}
	if len(items) > 0 {
		endCursor := modelId(items[len(items)-1])
		info.EndCursor = &endCursor
	}

	var edges []GeneratedModel
	for _, item := range items {
		edges = append(edges, p.mapModel(item))
	}

	conn := p.makeConn(edges, info)
	return &conn, nil
}

func tableName[T any](db *gorm.DB) string {
	var dummy T
	stmt := gorm.Statement{DB: db}
	_ = stmt.Parse(dummy)
	return stmt.Schema.Table
}
