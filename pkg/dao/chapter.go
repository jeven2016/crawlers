package dao

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model/entity"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"time"
)

type chapterInterface interface {
	FindByName(ctx context.Context, name string) (*entity.Chapter, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	Insert(ctx context.Context, novel *entity.Chapter) (*primitive.ObjectID, error)
	BulkInsert(ctx context.Context, chapters []*entity.Chapter, novelId *primitive.ObjectID) error
	Save(ctx context.Context, novel *entity.Chapter) (*primitive.ObjectID, error)
}

type chapterDaoImpl struct{}

func (n *chapterDaoImpl) FindByName(ctx context.Context, name string) (*entity.Chapter, error) {
	chapter, err := FindByMongoFilter(ctx, bson.M{base.ColumnName: name}, base.CollectionChapter, &entity.Chapter{},
		&options.FindOneOptions{})
	if err != nil || chapter == nil {
		return nil, err
	}
	return chapter, err
}

func (n *chapterDaoImpl) ExistsByName(ctx context.Context, name string) (bool, error) {
	task, err := FindByMongoFilter(ctx, bson.M{base.ColumnName: name}, base.CollectionChapter, &entity.Chapter{},
		&options.FindOneOptions{Projection: bson.M{base.ColumId: 1}})
	return task != nil, err
}

func (n *chapterDaoImpl) Insert(ctx context.Context, novel *entity.Chapter) (*primitive.ObjectID, error) {
	collection := base.GetSystem().GetCollection(base.CollectionChapter)
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

func (n *chapterDaoImpl) Save(ctx context.Context, chapter *entity.Chapter) (*primitive.ObjectID, error) {
	if chapter.Id.IsZero() {
		//insert
		return n.Insert(ctx, chapter)
	} else {
		collection := base.GetSystem().GetCollection(base.CollectionChapter)
		if collection == nil {
			zap.L().Error("collection not found: " + base.CollectionChapter)
			return nil, errors.New("collection not found: " + base.CollectionChapter)
		}
		//update
		curTime := time.Now()
		chapter.UpdatedTime = &curTime

		taskBytes, err := bson.Marshal(chapter)
		if err != nil {
			return nil, err
		}
		var doc bson.D
		if err = bson.Unmarshal(taskBytes, &doc); err != nil {
			return nil, err
		}
		_, err = collection.UpdateOne(ctx, bson.M{base.ColumId: chapter.Id}, bson.M{"$set": doc})
		return &chapter.Id, err
	}
}

func (n *chapterDaoImpl) BulkInsert(ctx context.Context, chapters []*entity.Chapter, novelId *primitive.ObjectID) error {
	collection := base.GetSystem().GetCollection(base.CollectionChapter)

	documents := make([]interface{}, len(chapters))

	//保存chapters
	for i := 0; i < len(chapters); i++ {
		chapters[i].NovelId = *novelId
		documents[i] = chapters[i]
	}

	// 指定每个批次的文档数量
	BulkBatchSize := 10

	// 计算批次数量
	numBatches := (len(documents) + BulkBatchSize - 1) / BulkBatchSize
	// 分批插入文档
	for i := 0; i < numBatches; i++ {
		// 计算当前批次的起始和结束索引
		startIndex := i * BulkBatchSize
		endIndex := (i + 1) * BulkBatchSize
		if endIndex > len(chapters) {
			endIndex = len(chapters)
		}

		// 获取当前批次的文档
		batch := documents[startIndex:endIndex]

		// 执行批量插入操作
		_, err := collection.InsertMany(ctx, batch)
		if err != nil {
			zap.L().Error("failed to insert chapters", zap.String("novelId", novelId.Hex()))
			return err
		}
		zap.S().Info("the number of inserted chapters: ", numBatches*(i+1))
	}
	return nil
}
