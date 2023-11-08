package api

import (
	"crawlers/pkg/base"
	"crawlers/pkg/dao"
	"crawlers/pkg/model/dto"
	"crawlers/pkg/model/entity"
	"github.com/gin-gonic/gin"
	"github.com/jeven2016/mylibs/system"
	"github.com/jeven2016/mylibs/utils"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type SiteHandler struct {
	sys *system.System
}

func NewSiteHandler() *SiteHandler {
	return &SiteHandler{
		sys: base.GetSystem(),
	}
}

// CreateSite create a site
// @Tags 测试
// @Summary  创建新的可解析的网站
// @Description 创建新的可解析的网站，管理目录、Novel、章节等数据
// @Param   name	       body   entity.Site   true   "网站名称"
// @Param   displayName    body   entity.Site   true   "显示名称"
// @Param   crawlerType    body   entity.Site   true   "网站提供的资源类型"
// @Accept  application/json
// @Produce application/json
// @Success 201
// @Router /sites [post]
func (h *SiteHandler) CreateSite(c *gin.Context) {
	var site entity.Site
	if err := c.ShouldBindJSON(&site); err != nil {
		//自定义error， https://juejin.cn/post/7015517416608235534
		zap.L().Warn("failed to convert json", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest,
			base.FailsWithMessage(base.ErrCodeUnknown, err.Error()))
		return
	}
	currentTime := time.Now()
	site.CreatedTime = &currentTime

	h.doCreate(c, &dto.CreateRequest{
		Key:           "site",
		Name:          site.Name,
		Entity:        site,
		Collection:    base.CollectionSite,
		RedisCacheKey: utils.GenKey(base.SiteKeyExistsPrefix, site.Name),
	})
}

// CreateCatalog create a catalog
// @Tags 测试
// @Summary  创建网站下的目录
// @Description 创建新的创建网站目录，管理Novel、章节等数据
// @Param   siteId	body   entity.Catalog   true   "网站ID"
// @Param   name    body   entity.Catalog   true   "目录名称"
// @Param   url     body   entity.Catalog   true   "目录URL"
// @Accept  application/json
// @Produce application/json
// @Success 201
// @Router /sites [post]
func (h *SiteHandler) CreateCatalog(c *gin.Context) {
	var catalog entity.Catalog
	if err := c.ShouldBindJSON(&catalog); err != nil {
		zap.L().Warn("failed to convert json", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest,
			base.FailsWithMessage(base.ErrCodeUnknown, err.Error()))
		return
	}

	//check if the site exists and cache the result
	exists, err := utils.Exists(c, utils.GenKey(base.SiteKeyExistsPrefix, catalog.SiteId.Hex()), func() (any, error) {
		return dao.SiteDao.ExistsById(c, catalog.SiteId)
	})
	if err != nil {
		zap.L().Warn("failed to check if any sites exist with this siteId", zap.String("siteId", catalog.SiteId.Hex()),
			zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			base.FailsWithMessage(base.ErrCodeUnknown, err.Error()))
		return
	}
	if !exists {
		zap.L().Warn("no site exists with this siteId", zap.String("siteId", catalog.SiteId.Hex()))
		c.AbortWithStatusJSON(http.StatusBadRequest,
			base.FailsWithParams(base.ErrSiteNotFound, catalog.SiteId.Hex()))
		return
	}

	h.doCreate(c, &dto.CreateRequest{
		Key:           "catalog",
		Name:          catalog.Name,
		Entity:        catalog,
		Collection:    base.CollectionCatalog,
		RedisCacheKey: utils.GenKey(base.CatalogKeyExistsPrefix, catalog.Name),
	})
}

func (h *SiteHandler) doCreate(c *gin.Context, req *dto.CreateRequest) {
	col := base.GetSystem().GetCollection(req.Collection)

	exists, err := utils.Exists(c, req.RedisCacheKey, func() (any, error) {
		return dao.CatalogDao.ExistsByName(c, req.Name)
	})
	if err != nil {
		zap.L().Warn("failed to check if it exists", zap.Error(err), zap.Any("request", req.Entity))
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			base.FailsWithMessage(base.ErrCodeUnknown, err.Error()))
		return
	}

	if exists {
		zap.L().Warn("it's duplicated to save", zap.Any(req.Key, req.Name), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest,
			base.FailsWithParams(base.ErrDuplicated, req.Key, req.Name))
		return
	}

	if obj, err := col.InsertOne(c, req.Entity); err != nil {
		zap.L().Warn("failed to save", zap.Error(err), zap.Any(req.Key, req.Name))
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			base.FailsWithMessage(base.ErrCodeUnknown, err.Error()))
		return
	} else {
		c.JSON(http.StatusCreated, obj)
	}
}
