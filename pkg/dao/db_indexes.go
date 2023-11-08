package dao

import (
	"context"
	"crawlers/pkg/base"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// the collections exist with url index
var urlIndexCollection = []string{
	base.CollectionCatalogPageTask,
	base.CollectionNovelTask,
	base.CollectionChapterTask}

var nameIndexCollection = []string{
	base.CollectionCatalog,
	base.CollectionSite,
	base.CollectionNovel,
	base.CollectionChapter,
}

// EnsureMongoIndexes ensure the collections exist with indexes
func EnsureMongoIndexes(ctx context.Context) {
	zap.L().Info("ensure indexes of collections are created")
	for _, collection := range urlIndexCollection {
		// options.Index().SetUnique(true)
		//the following code fails if multiple rows conflict with url
		ensureIndex(ctx, collection, bson.M{base.ColumnUrl: -1}, nil)
	}

	for _, collection := range nameIndexCollection {
		ensureIndex(ctx, collection, bson.M{base.ColumnName: -1}, nil)
	}

	//for content
	ensureIndex(ctx, base.CollectionContent,
		bson.M{base.ColumnParentId: 0, base.ColumnPageNo: -1}, nil)
	zap.L().Info("completed checking the indexes of collections")
}

func ensureIndex(ctx context.Context, collection string, keys primitive.M, options *options.IndexOptions) {
	col := base.GetSystem().GetCollection(collection)
	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    keys,
		Options: options,
	})

	var cmdErr mongo.CommandError
	if err != nil {
		if ok := errors.As(err, &cmdErr); ok {
			//if cmdErr.Name == "DuplicateKey" {
			//	zap.S().Info(collection, " data conflicts")
			//	return
			//}
			zap.L().Warn("Failed to ensure index created",
				zap.String("collection", collection), zap.Error(err))
		}
	}
}
