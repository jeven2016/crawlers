package service

import (
	"crawlers/pkg/model/entity"
	"crawlers/pkg/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CatalogPageTaskServiceInterface interface {
	FindTasksByCatalogId(ctx *gin.Context, catalogId primitive.ObjectID) ([]entity.CatalogPageTask, error)
	FindById(ctx *gin.Context, id primitive.ObjectID) (*entity.CatalogPageTask, error)
	FindByUrl(ctx *gin.Context, url string) (*entity.CatalogPageTask, error)
	ExistsById(ctx *gin.Context, id primitive.ObjectID) (bool, error)
	ExistsByName(ctx *gin.Context, name string) (bool, error)
	Save(ctx *gin.Context, task *entity.CatalogPageTask) (*primitive.ObjectID, error)
}

type catalogPageTaskServiceImpl struct {
}

func NewCatalogPageTaskService() CatalogPageTaskServiceInterface {
	return &catalogPageTaskServiceImpl{}
}

func (c *catalogPageTaskServiceImpl) FindTasksByCatalogId(ctx *gin.Context, catalogId primitive.ObjectID) ([]entity.CatalogPageTask, error) {
	return repository.CatalogPageTaskRepo.FindTasksByCatalogId(ctx, catalogId)
}

func (c *catalogPageTaskServiceImpl) FindById(ctx *gin.Context, id primitive.ObjectID) (*entity.CatalogPageTask, error) {
	return repository.CatalogPageTaskRepo.FindById(ctx, id)
}

func (c *catalogPageTaskServiceImpl) FindByUrl(ctx *gin.Context, url string) (*entity.CatalogPageTask, error) {
	return repository.CatalogPageTaskRepo.FindByUrl(ctx, url)
}

func (c *catalogPageTaskServiceImpl) ExistsById(ctx *gin.Context, id primitive.ObjectID) (bool, error) {
	return repository.CatalogPageTaskRepo.ExistsById(ctx, id)
}

func (c *catalogPageTaskServiceImpl) ExistsByName(ctx *gin.Context, name string) (bool, error) {
	return repository.CatalogPageTaskRepo.ExistsByName(ctx, name)
}

func (c *catalogPageTaskServiceImpl) Save(ctx *gin.Context, task *entity.CatalogPageTask) (*primitive.ObjectID, error) {
	return repository.CatalogPageTaskRepo.Save(ctx, task)
}
