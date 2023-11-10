package api

import (
	"crawlers/pkg/base"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

// convert json string to model object
func bindJson(c *gin.Context, obj any) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		//自定义error， https://juejin.cn/post/7015517416608235534
		zap.L().Warn("failed to convert json", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest,
			base.FailsWithMessage(base.ErrCodeUnknown, err.Error()))
		return false
	}
	return true
}
