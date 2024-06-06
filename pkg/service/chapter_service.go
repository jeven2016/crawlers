package service

import (
	"crawlers/pkg/model/entity"
	"crawlers/pkg/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChapterServiceInterface interface {
	FindByName(ctx *gin.Context, name string) (*entity.Chapter, error)
	ExistsByName(ctx *gin.Context, name string) (bool, error)
	Insert(ctx *gin.Context, novel *entity.Chapter) (*primitive.ObjectID, error)
	BulkInsert(ctx *gin.Context, chapters []*entity.Chapter, novelId *primitive.ObjectID) error
	Save(ctx *gin.Context, novel *entity.Chapter) (*primitive.ObjectID, error)
}

type chapterServiceImpl struct {
}

func NewChapterService() ChapterServiceInterface {
	return &chapterServiceImpl{}
}

func (c *chapterServiceImpl) FindByName(ctx *gin.Context, name string) (*entity.Chapter, error) {
	return repository.ChapterRepo.FindByName(ctx, name)
}

func (c *chapterServiceImpl) ExistsByName(ctx *gin.Context, name string) (bool, error) {
	return repository.ChapterRepo.ExistsByName(ctx, name)
}

func (c *chapterServiceImpl) Insert(ctx *gin.Context, novel *entity.Chapter) (*primitive.ObjectID, error) {
	return repository.ChapterRepo.Insert(ctx, novel)
}

func (c *chapterServiceImpl) BulkInsert(ctx *gin.Context, chapters []*entity.Chapter, novelId *primitive.ObjectID) error {
	return repository.ChapterRepo.BulkInsert(ctx, chapters, novelId)
}

func (c *chapterServiceImpl) Save(ctx *gin.Context, novel *entity.Chapter) (*primitive.ObjectID, error) {
	return repository.ChapterRepo.Save(ctx, novel)
}
