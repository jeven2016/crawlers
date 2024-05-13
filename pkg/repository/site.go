package repository

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type siteInterface interface {
	FindSites(ctx context.Context) ([]entity.Site, error)
	FindById(ctx context.Context, id primitive.ObjectID) (*entity.Site, error)
	ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error)
	DeleteById(ctx context.Context, id primitive.ObjectID) error
}

type siteDaoImpl struct{}

func (s *siteDaoImpl) FindSites(ctx context.Context) ([]entity.Site, error) {
	findOpts := options.Find()
	//findOpts.SetProjection(bson.M{base.ColumId: 1, base.ColumnName: 1, base.ColumnDisplayName: 1})
	findOpts.SetLimit(1000)

	var sites []entity.Site
	err := FindAll(ctx, &sites, base.CollectionSite, bson.D{}, findOpts)
	return sites, err
}

func (s *siteDaoImpl) FindById(ctx context.Context, id primitive.ObjectID) (*entity.Site, error) {
	return FindById(ctx, id, base.CollectionSite, &entity.Site{})
}

func (s *siteDaoImpl) ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error) {
	site, err := FindById(ctx, id, base.CollectionSite, &entity.Site{},
		&options.FindOneOptions{Projection: bson.M{base.ColumId: 1}})
	return site != nil, err
}

func (s *siteDaoImpl) DeleteById(ctx context.Context, id primitive.ObjectID) error {
	return DeleteById(ctx, id, base.CollectionSite)
}
