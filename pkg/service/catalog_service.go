package service

import (
	"crawlers/pkg/model/entity"
	"crawlers/pkg/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CatalogServiceInterface interface {
	FindCatalogsBySiteId(ctx *gin.Context, siteId primitive.ObjectID) ([]entity.Catalog, error)
	FindById(ctx *gin.Context, id primitive.ObjectID) (*entity.Catalog, error)
	ExistsById(ctx *gin.Context, id primitive.ObjectID) (bool, error)
	ExistsByName(ctx *gin.Context, name string) (bool, error)
}

type catalogServiceImpl struct {
}

func NewCatalogService() CatalogServiceInterface {
	return &catalogServiceImpl{}
}

func (s *catalogServiceImpl) FindCatalogsBySiteId(ctx *gin.Context, siteId primitive.ObjectID) ([]entity.Catalog, error) {
	return repository.CatalogRepo.FindCatalogsBySiteId(ctx, siteId)
}

func (s *catalogServiceImpl) FindById(ctx *gin.Context, id primitive.ObjectID) (*entity.Catalog, error) {
	return repository.CatalogRepo.FindById(ctx, id)
}

func (s *catalogServiceImpl) ExistsById(ctx *gin.Context, id primitive.ObjectID) (bool, error) {
	return repository.CatalogRepo.ExistsById(ctx, id)
}

func (s *catalogServiceImpl) ExistsByName(ctx *gin.Context, name string) (bool, error) {
	return repository.CatalogRepo.ExistsByName(ctx, name)
}
