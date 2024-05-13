package repository

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model/entity"
	"errors"
	"github.com/jeven2016/mylibs/system"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"time"
)

type novelInterface interface {
	FindById(ctx context.Context, id primitive.ObjectID) (*entity.Novel, error)
	FindIdByName(ctx context.Context, name string) (*primitive.ObjectID, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	Insert(ctx context.Context, novel *entity.Novel) (*primitive.ObjectID, error)
	Save(ctx context.Context, task *entity.Novel) (*primitive.ObjectID, error)

	DeleteByIds(ctx context.Context, ids []*primitive.ObjectID) error
}

type novelDaoImpl struct{}

func (n *novelDaoImpl) FindById(ctx context.Context, id primitive.ObjectID) (*entity.Novel, error) {
	novel, err := FindOneByFilter(ctx, bson.M{base.ColumId: id}, base.CollectionNovel, &entity.Novel{},
		&options.FindOneOptions{})
	if err != nil || novel == nil {
		return nil, err
	}
	return novel, err
}

func (n *novelDaoImpl) FindIdByName(ctx context.Context, name string) (*primitive.ObjectID, error) {
	novel, err := FindOneByFilter(ctx, bson.M{base.ColumnName: name}, base.CollectionNovel, &entity.Novel{},
		&options.FindOneOptions{Projection: bson.M{base.ColumId: 1}})
	if err != nil || novel == nil {
		return nil, err
	}
	return &novel.Id, err
}

func (n *novelDaoImpl) ExistsByName(ctx context.Context, name string) (bool, error) {
	novel, err := FindOneByFilter(ctx, bson.M{base.ColumnName: name}, base.CollectionNovel, &entity.Novel{},
		&options.FindOneOptions{Projection: bson.M{base.ColumId: 1}})
	return novel != nil, err
}

func (n *novelDaoImpl) Insert(ctx context.Context, novel *entity.Novel) (*primitive.ObjectID, error) {
	collection := system.GetSystem().GetCollection(base.CollectionNovel)
	//for creating
	if !novel.Id.IsZero() {
		return nil, base.ErrDocumentIdExists
	}
	//check if name conflicts
	exists, err := n.ExistsByName(ctx, novel.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, base.ErrDuplicatedDocument
	}
	//insert
	if result, err := collection.InsertOne(ctx, novel, &options.InsertOneOptions{}); err != nil {
		return nil, err
	} else {
		insertedId := result.InsertedID.(primitive.ObjectID)
		return &insertedId, nil
	}
}

func (c *novelDaoImpl) Save(ctx context.Context, novel *entity.Novel) (*primitive.ObjectID, error) {
	if novel.Id.IsZero() {
		//insert
		return c.Insert(ctx, novel)
	} else {
		collection := system.GetSystem().GetCollection(base.CollectionNovel)
		if collection == nil {
			zap.L().Error("collection not found: " + base.CollectionNovel)
			return nil, errors.New("collection not found: " + base.CollectionNovel)
		}
		//update
		curTime := time.Now()
		novel.UpdatedTime = &curTime

		taskBytes, err := bson.Marshal(novel)
		if err != nil {
			return nil, err
		}
		var doc bson.D
		if err = bson.Unmarshal(taskBytes, &doc); err != nil {
			return nil, err
		}
		_, err = collection.UpdateOne(ctx, bson.M{base.ColumId: novel.Id}, bson.M{"$set": doc})
		return &novel.Id, err
	}
}

func (c *novelDaoImpl) DeleteByIds(ctx context.Context, ids []*primitive.ObjectID) error {
	if len(ids) == 0 {
		return nil
	}
	collection := system.GetSystem().GetCollection(base.CollectionNovelTask)
	if collection == nil {
		zap.L().Error("collection not found: " + base.CollectionNovelTask)
		return errors.New("collection not found: " + base.CollectionNovelTask)
	}

	_, err := collection.DeleteMany(ctx, bson.M{base.ColumId: bson.M{"$in": ids}})
	return err
}
