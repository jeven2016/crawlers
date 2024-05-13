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

type novelTaskInterface interface {
	FindByCatalogId(ctx context.Context, catalogId primitive.ObjectID) ([]entity.NovelTask, error)
	FindByUrl(ctx context.Context, url string) (*entity.NovelTask, error)
	Save(ctx context.Context, task *entity.NovelTask) (*primitive.ObjectID, error)
}

type novelTaskDaoImpl struct{}

func (c *novelTaskDaoImpl) FindByCatalogId(ctx context.Context, catalogId primitive.ObjectID) ([]entity.NovelTask, error) {
	findOpts := options.Find()
	//findOpts.SetProjection(bson.M{base.ColumId: 1, base.ColumnName: 1, base.ColumnDisplayName: 1})
	findOpts.SetLimit(1000)

	var tasks []entity.NovelTask
	err := FindAll(ctx, &tasks, base.CollectionNovelTask, bson.M{base.ColumnCatalogId: catalogId}, findOpts)

	if tasks == nil {
		tasks = []entity.NovelTask{}
	}
	return tasks, err
}

func (c *novelTaskDaoImpl) FindByUrl(ctx context.Context, url string) (*entity.NovelTask, error) {
	task, err := FindOneByFilter(ctx, bson.M{base.ColumnUrl: url}, base.CollectionNovelTask, &entity.NovelTask{})
	return task, err
}

func (c *novelTaskDaoImpl) Save(ctx context.Context, task *entity.NovelTask) (*primitive.ObjectID, error) {
	collection := system.GetSystem().GetCollection(base.CollectionNovelTask)
	if collection == nil {
		zap.L().Error("collection not found: " + base.CollectionNovelTask)
		return nil, errors.New("collection not found: " + base.CollectionNovelTask)
	}
	if task.Id.IsZero() {
		//insert
		if result, err := collection.InsertOne(ctx, task, &options.InsertOneOptions{}); err != nil {
			return nil, err
		} else {
			insertedId := result.InsertedID.(primitive.ObjectID)
			return &insertedId, nil
		}
	} else {
		//update
		curTime := time.Now()
		task.LastUpdated = &curTime
		taskBytes, err := bson.Marshal(task)
		if err != nil {
			return nil, err
		}
		var doc bson.D
		if err = bson.Unmarshal(taskBytes, &doc); err != nil {
			return nil, err
		}
		_, err = collection.UpdateOne(ctx,
			bson.M{base.ColumId: task.Id, base.ColumnSiteName: task.SiteName}, bson.M{"$set": doc})
		return &task.Id, err
	}
}
