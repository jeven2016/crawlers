package dao

import (
	"context"
	"crawlers/pkg/base"
	"errors"
	"github.com/jeven2016/mylibs/system"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

func FindById[T any](ctx context.Context, id primitive.ObjectID, collection string,
	obj *T, opts ...*options.FindOneOptions) (*T, error) {
	return FindOneByFilter(ctx, bson.M{base.ColumId: id}, collection, obj, opts...)
}

func DeleteById(ctx context.Context, id primitive.ObjectID, collection string,
	opts ...*options.DeleteOptions) error {
	col := system.GetSystem().GetCollection(collection)
	if col == nil {
		return errors.New("collection not found: " + collection)
	}
	_, err := col.DeleteOne(ctx, bson.M{base.ColumId: id}, opts...)
	return err
}

func FindAll(ctx context.Context, list any, collection string, filter any, opts *options.FindOptions) error {
	col := system.GetSystem().GetCollection(collection)

	//findOpts := options.Find()
	//findOpts.SetProjection(bson.M{base.ColumId: 1, base.ColumnName: 1, base.ColumnDisplayName: 1})
	//findOpts.SetProjection(projection)
	//findOpts.SetLimit(1000)

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return err
	}
	defer func() {
		err = cursor.Close(context.TODO())
		if err != nil {
			zap.L().Warn("an error occurs while closing a cursor", zap.Error(err))
		}
	}()

	err = cursor.All(context.Background(), list)
	if err != nil {
		return err
	}
	return nil
}

func FindOneByFilter[T any](ctx context.Context, mongoFilter interface{}, collection string,
	decodedObj *T, opts ...*options.FindOneOptions) (*T, error) {
	col := system.GetSystem().GetCollection(collection)
	if col == nil {
		return nil, errors.New("collection not found: " + collection)
	}
	if result := col.FindOne(ctx, mongoFilter, opts...); result.Err() != nil {
		err := result.Err()
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	} else {
		if decodedObj != nil {
			err := result.Decode(decodedObj)
			return decodedObj, err
		}
		return nil, base.ErrDecodingDocument
	}
}

func ExistsByMongoFilter(ctx context.Context, mongoFilter interface{},
	collection string, opts ...*options.FindOneOptions) (bool, error) {
	var obj *interface{}
	_, err := FindOneByFilter(ctx, mongoFilter, collection, obj, opts...)
	if err != nil && errors.Is(err, base.ErrDecodingDocument) {
		return true, nil
	}
	return false, err
}
