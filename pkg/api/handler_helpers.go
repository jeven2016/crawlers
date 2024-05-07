package api

import (
	"crawlers/pkg/base"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"net/http"
)

func ensureValidId(c *gin.Context, id string) *primitive.ObjectID {
	if id == "" {
		zap.L().Warn("invalid id", zap.String("id", id))
		c.AbortWithStatusJSON(http.StatusBadRequest,
			base.FailsWithParams(base.ErrSiteNotFound, id))
		return nil
	}
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		zap.L().Warn("invalid id", zap.String("id", id), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest,
			base.FailsWithMessage(base.ErrCodeUnknown, err.Error()))
		return nil
	}
	return &objectId
}
