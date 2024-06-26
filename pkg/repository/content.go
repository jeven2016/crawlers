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

type contentRepo interface {
	FindByParentIdAndPage(ctx context.Context, parentId *primitive.ObjectID, pageNo int) (*entity.Content, error)
	Insert(ctx context.Context, content *entity.Content) (*primitive.ObjectID, error)
	Save(ctx context.Context, novel *entity.Content) (*primitive.ObjectID, error)
}

type contentRepoImpl struct{}

func (c *contentRepoImpl) Insert(ctx context.Context, content *entity.Content) (*primitive.ObjectID, error) {
	collection := system.GetSystem().GetCollection(base.CollectionContent)
	//for creating
	if !content.Id.IsZero() {
		return nil, base.ErrDocumentIdExists
	}
	//check if name conflicts
	existingContent, err := c.FindByParentIdAndPage(ctx, &content.ParentId, content.Page)
	if err != nil {
		return nil, err
	}
	if existingContent != nil {
		return nil, base.ErrDuplicatedDocument
	}
	//insert
	if result, err := collection.InsertOne(ctx, content, &options.InsertOneOptions{}); err != nil {
		return nil, err
	} else {
		insertedId := result.InsertedID.(primitive.ObjectID)
		return &insertedId, nil
	}
}

func (c *contentRepoImpl) FindByParentIdAndPage(ctx context.Context, parentId *primitive.ObjectID, pageNo int) (*entity.Content, error) {
	task, err := FindOneByFilter(ctx, bson.M{base.ColumnParentId: parentId}, //TODO: common.ColumnPageNo: pageNo
		base.CollectionContent, &entity.Content{},
		&options.FindOneOptions{})
	return task, err
}

func (c *contentRepoImpl) Save(ctx context.Context, content *entity.Content) (*primitive.ObjectID, error) {
	if content.Id.IsZero() {
		//insert
		return c.Insert(ctx, content)
	} else {
		collection := system.GetSystem().GetCollection(base.CollectionContent)
		if collection == nil {
			zap.L().Error("collection not found: " + base.CollectionContent)
			return nil, errors.New("collection not found: " + base.CollectionContent)
		}
		//update
		curTime := time.Now()
		content.UpdatedTime = &curTime

		taskBytes, err := bson.Marshal(content)
		if err != nil {
			return nil, err
		}
		var doc bson.D
		if err = bson.Unmarshal(taskBytes, &doc); err != nil {
			return nil, err
		}
		_, err = collection.UpdateOne(ctx, bson.M{base.ColumId: content.Id}, bson.M{"$set": doc})
		return &content.Id, err
	}
}
