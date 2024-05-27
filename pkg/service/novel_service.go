package service

import (
	"context"
	"crawlers/pkg/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NovelServiceInterface interface {
	DeleteByIds(ctx context.Context, ids []string) error
}

type novelServiceImpl struct {
}

func NewNovelService() NovelServiceInterface {
	return &novelServiceImpl{}
}

func (s *novelServiceImpl) DeleteByIds(ctx context.Context, ids []string) error {
	var objectIdArray []*primitive.ObjectID

	for _, id := range ids {
		objectId, err := primitive.ObjectIDFromHex(id)

		if err != nil {
			return err
		}
		objectIdArray = append(objectIdArray, &objectId)
	}
	err := repository.NovelRepo.DeleteByIds(ctx, objectIdArray)
	return err
}
