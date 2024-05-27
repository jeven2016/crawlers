package service

import (
	"context"
	"crawlers/pkg/model/entity"
	"crawlers/pkg/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SiteServiceInterface interface {
	FindSites(ctx context.Context) ([]entity.Site, error)
	FindById(ctx context.Context, id primitive.ObjectID) (*entity.Site, error)
	ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error)
	DeleteById(ctx context.Context, id primitive.ObjectID) error
	FindSettings(ctx context.Context, siteId primitive.ObjectID) (*entity.SiteSettings, error)
	SaveSettings(ctx context.Context, siteSettings *entity.SiteSettings) (*entity.SiteSettings, error)
}

type siteServiceImpl struct {
}

func NewSiteService() SiteServiceInterface {
	return &siteServiceImpl{}
}

func (s siteServiceImpl) FindSites(ctx context.Context) ([]entity.Site, error) {
	return repository.SiteRepo.FindSites(ctx)
}

func (s siteServiceImpl) FindById(ctx context.Context, id primitive.ObjectID) (*entity.Site, error) {
	return repository.SiteRepo.FindById(ctx, id)
}

func (s siteServiceImpl) ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error) {
	return repository.SiteRepo.ExistsById(ctx, id)
}

func (s siteServiceImpl) DeleteById(ctx context.Context, id primitive.ObjectID) error {
	return repository.SiteRepo.DeleteById(ctx, id)
}

func (s siteServiceImpl) FindSettings(ctx context.Context, siteId primitive.ObjectID) (*entity.SiteSettings, error) {
	site, err := SiteService.FindById(ctx, siteId)
	if err != nil {
		return nil, err
	}

	return repository.SiteRepo.FindSettings(ctx, siteId)
}

func (s siteServiceImpl) SaveSettings(ctx context.Context, siteSettings *entity.SiteSettings) (*entity.SiteSettings, error) {
	return repository.SiteRepo.SaveSettings(ctx, siteSettings)
}
