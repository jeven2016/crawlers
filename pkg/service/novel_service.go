package service

import (
	"crawlers/pkg/model/entity"
	"crawlers/pkg/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NovelServiceInterface interface {
	FindById(ctx *gin.Context, id primitive.ObjectID) (*entity.Novel, error)
	FindIdByName(ctx *gin.Context, name string) (*primitive.ObjectID, error)
	ExistsByName(ctx *gin.Context, name string) (bool, error)
	Insert(ctx *gin.Context, novel *entity.Novel) (*primitive.ObjectID, error)
	Save(ctx *gin.Context, task *entity.Novel) (*primitive.ObjectID, error)

	DeleteByIds(ctx *gin.Context, ids []string) error
}

type novelServiceImpl struct {
}

func NewNovelService() NovelServiceInterface {
	return &novelServiceImpl{}
}

func (s *novelServiceImpl) FindById(ctx *gin.Context, id primitive.ObjectID) (*entity.Novel, error) {
	return repository.NovelRepo.FindById(ctx, id)
}

func (s *novelServiceImpl) FindIdByName(ctx *gin.Context, name string) (*primitive.ObjectID, error) {
	return repository.NovelRepo.FindIdByName(ctx, name)
}

func (s *novelServiceImpl) ExistsByName(ctx *gin.Context, name string) (bool, error) {
	return repository.NovelRepo.ExistsByName(ctx, name)
}

func (s *novelServiceImpl) Insert(ctx *gin.Context, novel *entity.Novel) (*primitive.ObjectID, error) {
	return repository.NovelRepo.Insert(ctx, novel)
}

func (s *novelServiceImpl) Save(ctx *gin.Context, task *entity.Novel) (*primitive.ObjectID, error) {
	return repository.NovelRepo.Save(ctx, task)
}

func (s *novelServiceImpl) DeleteByIds(ctx *gin.Context, ids []string) error {
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
