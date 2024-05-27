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

type catalogPageTaskRepo interface {
	FindTasksByCatalogId(ctx context.Context, catalogId primitive.ObjectID) ([]entity.CatalogPageTask, error)
	FindById(ctx context.Context, id primitive.ObjectID) (*entity.CatalogPageTask, error)
	FindByUrl(ctx context.Context, url string) (*entity.CatalogPageTask, error)
	ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	Save(ctx context.Context, task *entity.CatalogPageTask) (*primitive.ObjectID, error)
}

type catalogPageTaskRepoImpl struct{}

func (c *catalogPageTaskRepoImpl) FindTasksByCatalogId(ctx context.Context, catalogId primitive.ObjectID) ([]entity.CatalogPageTask, error) {
	findOpts := options.Find()
	//findOpts.SetProjection(bson.M{base.ColumId: 1, base.ColumnName: 1, base.ColumnDisplayName: 1})
	findOpts.SetLimit(1000)

	var tasks []entity.CatalogPageTask
	err := FindAll(ctx, &tasks, base.CollectionCatalogPageTask, bson.M{base.ColumnCatalogId: catalogId}, findOpts)

	if tasks == nil {
		tasks = []entity.CatalogPageTask{}
	}
	return tasks, err
}

func (c *catalogPageTaskRepoImpl) FindById(ctx context.Context, id primitive.ObjectID) (*entity.CatalogPageTask, error) {
	return FindById(ctx, id, base.CollectionCatalogPageTask, &entity.CatalogPageTask{})
}

func (c *catalogPageTaskRepoImpl) FindByUrl(ctx context.Context, url string) (*entity.CatalogPageTask, error) {
	task, err := FindOneByFilter(ctx, bson.M{base.ColumnUrl: url}, base.CollectionCatalogPageTask, &entity.CatalogPageTask{})
	return task, err
}

func (s *catalogPageTaskRepoImpl) ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error) {
	task, err := FindById(ctx, id, base.CollectionCatalogPageTask, &entity.CatalogPageTask{},
		&options.FindOneOptions{Projection: bson.M{base.ColumId: 1}})
	return task != nil, err
}

func (s *catalogPageTaskRepoImpl) ExistsByName(ctx context.Context, name string) (bool, error) {
	task, err := FindOneByFilter(ctx, bson.M{base.ColumnName: name}, base.CollectionCatalogPageTask, &entity.CatalogPageTask{},
		&options.FindOneOptions{Projection: bson.M{base.ColumId: 1}})
	return task != nil, err
}

func (c *catalogPageTaskRepoImpl) Save(ctx context.Context, task *entity.CatalogPageTask) (*primitive.ObjectID, error) {
	collection := system.GetSystem().GetCollection(base.CollectionCatalogPageTask)
	if collection == nil {
		zap.L().Error("collection not found: " + base.CollectionCatalogPageTask)
		return nil, errors.New("collection not found: " + base.CollectionCatalogPageTask)
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
		_, err = collection.UpdateOne(ctx, bson.M{base.ColumId: task.Id, base.ColumnSiteName: task.SiteName}, bson.M{"$set": doc})
		return &task.Id, err
	}
}
