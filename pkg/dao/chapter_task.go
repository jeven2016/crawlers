package dao

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"time"
)

type chapterTaskInterface interface {
	FindByUrl(ctx context.Context, url string) (*model.ChapterTask, error)
	Save(ctx context.Context, task *model.ChapterTask) (*primitive.ObjectID, error)
}

type chapterTaskDaoImpl struct{}

func (c *chapterTaskDaoImpl) FindByUrl(ctx context.Context, url string) (*model.ChapterTask, error) {
	task, err := FindByMongoFilter(ctx, bson.M{base.ColumnUrl: url}, base.CollectionChapterTask, &model.ChapterTask{})
	return task, err
}

func (c *chapterTaskDaoImpl) Save(ctx context.Context, task *model.ChapterTask) (*primitive.ObjectID, error) {
	collection := base.GetSystem().GetCollection(base.CollectionChapterTask)
	if collection == nil {
		zap.L().Error("collection not found: " + base.CollectionChapterTask)
		return nil, errors.New("collection not found: " + base.CollectionChapterTask)
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
