package handler

import (
	"crawlers/pkg/base"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"net/http"
)

// convert json string to model object
func bindJson(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		//自定义error， https://juejin.cn/post/7015517416608235534
		zap.L().Warn("failed to convert json", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest,
			base.FailsWithError(c, err))
		return false
	}
	return true
}

func ensureValidId(c *gin.Context, id string) *primitive.ObjectID {
	if id == "" {
		zap.L().Warn("invalid id", zap.String("id", id))
		c.AbortWithStatusJSON(http.StatusBadRequest,
			base.FailsWithParams(c, base.ErrorCode.Required, map[string]string{"name": "id"}))
		return nil
	}
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		zap.L().Warn("invalid id", zap.String("id", id), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest,
			base.FailsWithMessage(base.ErrorCode.Unexpected, err.Error()))
		return nil
	}
	return &objectId
}
