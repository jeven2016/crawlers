package service

import (
	"crawlers/pkg/base"
	"crawlers/pkg/model/entity"
	"crawlers/pkg/repository"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type SiteServiceInterface interface {
	FindSites(ctx *gin.Context) ([]entity.Site, *base.AppError)
	FindById(ctx *gin.Context, id primitive.ObjectID) (*entity.Site, error)
	ExistsById(ctx *gin.Context, id primitive.ObjectID) (bool, error)
	DeleteById(ctx *gin.Context, id primitive.ObjectID) error
	FindSettings(ctx *gin.Context, siteId primitive.ObjectID) (*entity.SiteSettings, *base.AppError)
	SaveSettings(ctx *gin.Context, siteSettings *entity.SiteSettings) (*entity.SiteSettings, error)
}

type siteServiceImpl struct {
}

func NewSiteService() SiteServiceInterface {
	return &siteServiceImpl{}
}

func (s siteServiceImpl) FindSites(ctx *gin.Context) ([]entity.Site, *base.AppError) {
	sites, err := repository.SiteRepo.FindSites(ctx)
	if err != nil {
		return nil, base.NewAppError(base.ErrorCode.Unexpected, err.Error())
	}
	return sites, nil
}

func (s siteServiceImpl) FindById(ctx *gin.Context, id primitive.ObjectID) (*entity.Site, error) {
	return repository.SiteRepo.FindById(ctx, id)
}

func (s siteServiceImpl) ExistsById(ctx *gin.Context, id primitive.ObjectID) (bool, error) {
	return repository.SiteRepo.ExistsById(ctx, id)
}

func (s siteServiceImpl) DeleteById(ctx *gin.Context, id primitive.ObjectID) error {
	return repository.SiteRepo.DeleteById(ctx, id)
}

func (s siteServiceImpl) FindSettings(ctx *gin.Context, siteId primitive.ObjectID) (*entity.SiteSettings, *base.AppError) {
	exists, err := SiteService.ExistsById(ctx, siteId)
	if err != nil {
		zap.L().Warn("unexpected error occurs while checking if the site exists", zap.Error(err),
			zap.String("siteId", primitive.ObjectID.Hex(siteId)))
		return nil, base.NewAppError(base.ErrorCode.Unexpected, err.Error())
	}
	if !exists {
		zap.L().Warn("site not found", zap.String("siteId", primitive.ObjectID.Hex(siteId)))
		return nil, base.NewAppError(base.ErrorCode.NotFound)
	}

	settings, err := repository.SiteRepo.FindSettings(ctx, siteId)
	if err != nil {
		siteIdString := primitive.ObjectID.Hex(siteId)
		zap.L().Warn("unexpected error occurs while finding site settings",
			zap.String("siteId", siteIdString), zap.Error(err))
		msg := fmt.Sprintf("site settings(%v): %v", siteIdString, err.Error())
		return nil, base.NewAppError(base.ErrorCode.Unexpected, msg)
	}
	return settings, nil
}

func (s siteServiceImpl) SaveSettings(ctx *gin.Context, siteSettings *entity.SiteSettings) (*entity.SiteSettings, error) {
	return repository.SiteRepo.SaveSettings(ctx, siteSettings)
}
