package repository

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model/entity"
	"errors"
	"github.com/jeven2016/mylibs/system"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"time"
)

type siteRepo interface {
	FindSites(ctx context.Context) ([]entity.Site, error)
	FindById(ctx context.Context, id primitive.ObjectID) (*entity.Site, error)
	ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error)
	DeleteById(ctx context.Context, id primitive.ObjectID) error
	FindSettings(ctx context.Context, siteId primitive.ObjectID) (*entity.SiteSettings, error)
	SaveSettings(ctx context.Context, siteSettings *entity.SiteSettings) (*entity.SiteSettings, error)
}

type siteRepoImpl struct{}

func (s *siteRepoImpl) FindSites(ctx context.Context) ([]entity.Site, error) {
	findOpts := options.Find()
	//findOpts.SetProjection(bson.M{base.ColumId: 1, base.ColumnName: 1, base.ColumnDisplayName: 1})
	findOpts.SetLimit(1000)

	var sites []entity.Site
	err := FindAll(ctx, &sites, base.CollectionSite, bson.D{}, findOpts)
	return sites, err
}

func (s *siteRepoImpl) FindById(ctx context.Context, id primitive.ObjectID) (*entity.Site, error) {
	return FindById(ctx, id, base.CollectionSite, &entity.Site{})
}

func (s *siteRepoImpl) ExistsById(ctx context.Context, id primitive.ObjectID) (bool, error) {
	site, err := FindById(ctx, id, base.CollectionSite, &entity.Site{},
		&options.FindOneOptions{Projection: bson.M{base.ColumId: 1}})
	return site != nil, err
}

func (s *siteRepoImpl) DeleteById(ctx context.Context, id primitive.ObjectID) error {
	return DeleteById(ctx, id, base.CollectionSite)
}

func (s *siteRepoImpl) FindSettings(ctx context.Context, siteId primitive.ObjectID) (*entity.SiteSettings, error) {
	return FindByColumn(ctx, base.ColumnSiteId, siteId, base.CollectionSiteSettings, &entity.SiteSettings{})
}

// SaveSettings saves or updates the site settings in the database.
// If the site settings already exist, it updates the existing record.
// If the site settings do not exist, it creates a new record.
//
// Parameters:
// ctx (context.Context): The context for the database operation.
// siteSettings (*entity.SiteSettings): The site settings to be saved or updated.
//
// Returns:
// (*entity.SiteSettings, error): The saved or updated site settings and any error that occurred during the operation.
// If the operation is successful, the error will be nil.
func (s *siteRepoImpl) SaveSettings(ctx context.Context, siteSettings *entity.SiteSettings) (*entity.SiteSettings, error) {
	// Get the collection for site settings from the system
	collection := system.GetSystem().GetCollection(base.CollectionSiteSettings)
	if collection == nil {
		zap.L().Error("collection not found: " + base.CollectionSiteSettings)
		return nil, errors.New("collection not found: " + base.CollectionSiteSettings)
	}

	// Update the updated time of the site settings
	curTime := time.Now()
	siteSettings.UpdatedTime = &curTime

	// Marshal the site settings to BSON
	taskBytes, err := bson.Marshal(siteSettings)
	if err != nil {
		return nil, err
	}

	// Unmarshal the BSON to a BSON document
	var doc bson.D
	if err = bson.Unmarshal(taskBytes, &doc); err != nil {
		return nil, err
	}

	// Set the upsert option to true to upsert the record
	opts := options.Update().SetUpsert(true)

	// Upsert the record in the collection
	result, err := collection.UpdateOne(ctx, bson.M{base.ColumId: siteSettings.SiteId}, doc, opts)
	if err != nil {
		return nil, err
	}

	// Find and return the saved or updated site settings
	return FindById(ctx, result.UpsertedID.(primitive.ObjectID), base.CollectionSiteSettings, &entity.SiteSettings{})
}
