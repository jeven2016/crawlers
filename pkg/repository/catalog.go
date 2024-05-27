package repository

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type catalogRepo interface {
	FindCatalogsBySiteId(ctx context.Context, siteId primitive.ObjectID) ([]entity.Catalog, error)
	FindById(ctx context.Context, id primitive.ObjectID) (*entity.Catalog, error)
	ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
}

type catalogRepoImpl struct{}

func (c *catalogRepoImpl) FindCatalogsBySiteId(ctx context.Context, siteId primitive.ObjectID) ([]entity.Catalog, error) {
	findOpts := options.Find()
	//findOpts.SetProjection(bson.M{base.ColumId: 1, base.ColumnName: 1, base.ColumnDisplayName: 1})
	findOpts.SetLimit(1000)

	var catalogs []entity.Catalog
	err := FindAll(ctx, &catalogs, base.CollectionCatalog, bson.M{base.ColumnsiteId: siteId}, findOpts)

	if catalogs == nil {
		catalogs = []entity.Catalog{}
	}
	return catalogs, err
}

func (c *catalogRepoImpl) FindById(ctx context.Context, id primitive.ObjectID) (*entity.Catalog, error) {
	return FindById(ctx, id, base.CollectionCatalog, &entity.Catalog{})
}

func (s *catalogRepoImpl) ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error) {
	site, err := FindById(ctx, id, base.CollectionCatalog, &entity.Site{},
		&options.FindOneOptions{Projection: bson.M{base.ColumId: 1}})
	return site != nil, err
}

func (s *catalogRepoImpl) ExistsByName(ctx context.Context, name string) (bool, error) {
	site, err := FindOneByFilter(ctx, bson.M{base.ColumnName: name}, base.CollectionCatalog, &entity.Site{},
		&options.FindOneOptions{Projection: bson.M{base.ColumId: 1}})
	return site != nil, err
}
