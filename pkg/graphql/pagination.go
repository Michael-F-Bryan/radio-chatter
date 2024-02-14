package graphql

import (
	"fmt"

	"github.com/Michael-F-Bryan/radio-chatter/pkg/graphql/model"
	"gorm.io/gorm"
)

type paginator[Model any, GeneratedModel any, Connection any] struct {
	mapModel    func(model Model) GeneratedModel
	makeConn    func(edges []GeneratedModel, page model.PageInfo) Connection
	Filter      *Model
	BeforeQuery func(db *gorm.DB) *gorm.DB
	Limit       int
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
		stmt := gorm.Statement{DB: db}
		_ = stmt.Parse(dummy)
		db = db.Where(stmt.Schema.Table+".id > ?", id)
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
