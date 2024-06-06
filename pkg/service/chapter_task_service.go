package service

import (
	"crawlers/pkg/model/entity"
	"crawlers/pkg/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChapterTaskServiceInterface interface {
	FindByUrl(ctx *gin.Context, url string) (*entity.ChapterTask, error)
	Save(ctx *gin.Context, task *entity.ChapterTask) (*primitive.ObjectID, error)
}

type ChapterTaskServiceImpl struct{}

func NewChapterTaskService() ChapterTaskServiceInterface {
	return &ChapterTaskServiceImpl{}
}

func (c *ChapterTaskServiceImpl) FindByUrl(ctx *gin.Context, url string) (*entity.ChapterTask, error) {
	return repository.ChapterTaskRepo.FindByUrl(ctx, url)
}

func (c *ChapterTaskServiceImpl) Save(ctx *gin.Context, task *entity.ChapterTask) (*primitive.ObjectID, error) {
	return repository.ChapterTaskRepo.Save(ctx, task)
}
