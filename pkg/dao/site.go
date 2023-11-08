package dao

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type siteInterface interface {
	FindById(ctx context.Context, id primitive.ObjectID) (*entity.Site, error)
	ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error)
}

type siteDaoImpl struct{}

func (s *siteDaoImpl) FindById(ctx context.Context, id primitive.ObjectID) (*entity.Site, error) {
	return FindById(ctx, id, base.CollectionSite, &entity.Site{})
}

func (s *siteDaoImpl) ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error) {
	site, err := FindById(ctx, id, base.CollectionSite, &entity.Site{},
		&options.FindOneOptions{Projection: bson.M{base.ColumId: 1}})
	return site != nil, err
}
