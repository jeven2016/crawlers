package service

import (
	"crawlers/pkg/model/entity"
	"crawlers/pkg/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContentServiceInterface interface {
	FindByParentIdAndPage(ctx *gin.Context, parentId *primitive.ObjectID, pageNo int) (*entity.Content, error)
	Insert(ctx *gin.Context, content *entity.Content) (*primitive.ObjectID, error)
	Save(ctx *gin.Context, novel *entity.Content) (*primitive.ObjectID, error)
}

type contentServiceImpl struct {
}

func NewContentService() ContentServiceInterface {
	return &contentServiceImpl{}
}

func (c contentServiceImpl) FindByParentIdAndPage(ctx *gin.Context, parentId *primitive.ObjectID, pageNo int) (*entity.Content, error) {
	return repository.ContentRepo.FindByParentIdAndPage(ctx, parentId, pageNo)
}

func (c contentServiceImpl) Insert(ctx *gin.Context, content *entity.Content) (*primitive.ObjectID, error) {
	return repository.ContentRepo.Insert(ctx, content)
}

func (c contentServiceImpl) Save(ctx *gin.Context, novel *entity.Content) (*primitive.ObjectID, error) {
	return repository.ContentRepo.Save(ctx, novel)
}
