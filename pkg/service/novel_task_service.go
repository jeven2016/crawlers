package service

import (
	"crawlers/pkg/model/entity"
	"crawlers/pkg/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NovelTaskServiceInterface interface {
	FindByCatalogId(ctx *gin.Context, catalogId primitive.ObjectID) ([]entity.NovelTask, error)
	FindByUrl(ctx *gin.Context, url string) (*entity.NovelTask, error)
	Save(ctx *gin.Context, task *entity.NovelTask) (*primitive.ObjectID, error)
}

type novelTaskServiceImpl struct{}

func NewNovelTaskService() NovelTaskServiceInterface {
	return &novelTaskServiceImpl{}
}

func (impl *novelTaskServiceImpl) FindByCatalogId(ctx *gin.Context, catalogId primitive.ObjectID) ([]entity.NovelTask, error) {
	return repository.NovelTaskRepo.FindByCatalogId(ctx, catalogId)
}

func (impl *novelTaskServiceImpl) FindByUrl(ctx *gin.Context, url string) (*entity.NovelTask, error) {
	return repository.NovelTaskRepo.FindByUrl(ctx, url)
}

func (impl *novelTaskServiceImpl) Save(ctx *gin.Context, task *entity.NovelTask) (*primitive.ObjectID, error) {
	return repository.NovelTaskRepo.Save(ctx, task)
}
