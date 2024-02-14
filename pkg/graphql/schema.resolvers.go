package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.43

import (
	"context"
	"time"

	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/graphql/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DownloadURL is the resolver for the downloadUrl field.
func (r *chunkResolver) DownloadURL(ctx context.Context, obj *model.Chunk) (*string, error) {
	return signedURL(ctx, r.Logger, r.Storage, obj.Sha256)
}

// Transmissions is the resolver for the transmissions field.
func (r *chunkResolver) Transmissions(ctx context.Context, obj *model.Chunk, after *string, count int) (*model.TransmissionsConnection, error) {
	chunkId, err := decodeModelId[radiochatter.Chunk](obj.ID)
	if err != nil {
		return nil, err
	}

	p := paginator[radiochatter.Transmission, model.Transmission, model.TransmissionsConnection]{
		mapModel: transmissionToGraphQL,
		makeConn: func(edges []model.Transmission, page model.PageInfo) model.TransmissionsConnection {
			return model.TransmissionsConnection{Edges: edges, PageInfo: &page}
		},
		Filter: &radiochatter.Transmission{ChunkID: chunkId},
		Limit:  30,
	}

	return p.Page(r.DB, after, count)
}

// RegisterStream is the resolver for the registerStream field.
func (r *mutationResolver) RegisterStream(ctx context.Context, input model.RegisterStreamVariables) (*model.Stream, error) {
	stream := radiochatter.Stream{
		DisplayName: input.DisplayName,
		Url:         input.URL,
	}

	if err := r.DB.Save(&stream).Error; err != nil {
		return nil, err
	}

	r.Logger.Info("Stream created", zap.Any("stream", stream))

	model := streamToGraphQL(stream)
	return &model, nil
}

// RemoveStream is the resolver for the removeStream field.
func (r *mutationResolver) RemoveStream(ctx context.Context, id string) (*model.Stream, error) {
	realID, err := decodeModelId[radiochatter.Stream](id)
	if err != nil {
		return nil, err
	}

	var stream radiochatter.Stream
	if err := r.DB.Delete(&stream, "id = ?", realID).Error; err != nil {
		return nil, err
	}

	r.Logger.Info("Stream deleted", zap.Any("stream", stream))

	value := streamToGraphQL(stream)
	return &value, nil
}

// GetStreams is the resolver for the getStreams field.
func (r *queryResolver) GetStreams(ctx context.Context, after *string, count int) (*model.StreamsConnection, error) {
	p := paginator[radiochatter.Stream, model.Stream, model.StreamsConnection]{
		mapModel: streamToGraphQL,
		makeConn: func(edges []model.Stream, page model.PageInfo) model.StreamsConnection {
			return model.StreamsConnection{Edges: edges, PageInfo: &page}
		},
		Limit: 5,
	}

	return p.Page(r.DB, after, count)
}

// GetStreamByID is the resolver for the getStreamById field.
func (r *queryResolver) GetStreamByID(ctx context.Context, id string) (*model.Stream, error) {
	return getByID[radiochatter.Stream, model.Stream](r.DB, id, streamToGraphQL)
}

// GetChunkByID is the resolver for the getChunkById field.
func (r *queryResolver) GetChunkByID(ctx context.Context, id string) (*model.Chunk, error) {
	return getByID[radiochatter.Chunk, model.Chunk](r.DB, id, chunkToGraphQL)
}

// GetTransmissionByID is the resolver for the getTransmissionById field.
func (r *queryResolver) GetTransmissionByID(ctx context.Context, id string) (*model.Transmission, error) {
	return getByID[radiochatter.Transmission, model.Transmission](r.DB, id, transmissionToGraphQL)
}

// Chunks is the resolver for the chunks field.
func (r *streamResolver) Chunks(ctx context.Context, obj *model.Stream, after *string, count int) (*model.ChunksConnection, error) {
	streamId, err := decodeModelId[radiochatter.Stream](obj.ID)
	if err != nil {
		return nil, err
	}

	p := paginator[radiochatter.Chunk, model.Chunk, model.ChunksConnection]{
		mapModel: chunkToGraphQL,
		makeConn: func(edges []model.Chunk, page model.PageInfo) model.ChunksConnection {
			return model.ChunksConnection{Edges: edges, PageInfo: &page}
		},
		Filter: &radiochatter.Chunk{StreamID: streamId},
		Limit:  30,
	}

	return p.Page(r.DB, after, count)
}

// Transmissions is the resolver for the transmissions field.
func (r *streamResolver) Transmissions(ctx context.Context, obj *model.Stream, after *string, count int) (*model.TransmissionsConnection, error) {
	streamId, err := decodeModelId[radiochatter.Stream](obj.ID)
	if err != nil {
		return nil, err
	}

	p := paginator[radiochatter.Transmission, model.Transmission, model.TransmissionsConnection]{
		mapModel: transmissionToGraphQL,
		makeConn: func(edges []model.Transmission, page model.PageInfo) model.TransmissionsConnection {
			return model.TransmissionsConnection{Edges: edges, PageInfo: &page}
		},
		BeforeQuery: func(db *gorm.DB) *gorm.DB {
			return db.Joins("JOIN chunks ON chunks.id = transmissions.chunk_id AND chunks.stream_id = ?", streamId)
		},
		Limit: 30,
	}

	return p.Page(r.DB, after, count)
}

// Chunk is the resolver for the chunk field.
func (r *subscriptionResolver) Chunk(ctx context.Context) (<-chan *model.Chunk, error) {
	ch := pollForUpdates[radiochatter.Chunk, model.Chunk](
		ctx,
		r.DB,
		r.Logger,
		chunkToGraphQL,
		func(c radiochatter.Chunk) time.Time { return c.CreatedAt },
	)
	return ch, nil
}

// Transmission is the resolver for the transmission field.
func (r *subscriptionResolver) Transmission(ctx context.Context) (<-chan *model.Transmission, error) {
	ch := pollForUpdates[radiochatter.Transmission, model.Transmission](
		ctx,
		r.DB,
		r.Logger,
		transmissionToGraphQL,
		func(c radiochatter.Transmission) time.Time { return c.CreatedAt },
	)
	return ch, nil
}

// DownloadURL is the resolver for the downloadUrl field.
func (r *transmissionResolver) DownloadURL(ctx context.Context, obj *model.Transmission) (*string, error) {
	return signedURL(ctx, r.Logger, r.Storage, obj.Sha256)
}

// Chunk returns ChunkResolver implementation.
func (r *Resolver) Chunk() ChunkResolver { return &chunkResolver{r} }

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Stream returns StreamResolver implementation.
func (r *Resolver) Stream() StreamResolver { return &streamResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

// Transmission returns TransmissionResolver implementation.
func (r *Resolver) Transmission() TransmissionResolver { return &transmissionResolver{r} }

type chunkResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type streamResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
type transmissionResolver struct{ *Resolver }
