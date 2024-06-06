package handler

import (
	"crawlers/pkg/base"
	"crawlers/pkg/model/dto"
	"crawlers/pkg/model/entity"
	"crawlers/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/jeven2016/mylibs/system"
	"github.com/jeven2016/mylibs/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type SiteHandler struct{}

// NewSiteHandler Site handler for CRUD related operations
func NewSiteHandler() *SiteHandler {
	return &SiteHandler{}
}

func (h *SiteHandler) FindSites(c *gin.Context) {
	if sites, err := service.SiteService.FindSites(c); err != nil {
		zap.L().Warn("failed to find sites", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			base.FailsWithError(c, err))
		return
	} else {
		zap.L().Info("found sites", zap.Int("count", len(sites)))
		c.JSON(http.StatusOK, sites)
	}
}

func (h *SiteHandler) FindSiteCatalogs(c *gin.Context) {
	siteId := c.Param("siteId")
	siteObjectId := ensureValidId(c, siteId)
	if siteObjectId == nil {
		return
	}
	if catalogs, err := service.CatalogService.FindCatalogsBySiteId(c, *siteObjectId); err != nil {
		zap.L().Warn("failed to find catalogs", zap.String("siteId", siteId), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, base.FailsWithError(c, err))
	} else {
		zap.L().Info("found catalogs", zap.Int("count", len(catalogs)))
		c.JSON(http.StatusOK, catalogs)
	}
}

func (h *SiteHandler) FindSiteSettings(c *gin.Context) {
	siteId := c.Param("siteId")
	siteObjectId := ensureValidId(c, siteId)
	if siteObjectId == nil {
		return
	}
	site, err := service.SiteService.FindById(c, *siteObjectId)
	if err != nil {
		zap.L().Warn("failed to find site", zap.String("siteId", siteId), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, base.FailsWithError(c, err))
		return
	}
	if site == nil {
		zap.L().Warn("site not found", zap.String("siteId", siteId))
		c.Status(http.StatusNotFound)
		return
	}

	if siteSettings, err := service.SiteService.FindSettings(c, *siteObjectId); err != nil {
		zap.L().Warn("failed to find site settings", zap.String("siteId", siteId), zap.Error(err))
		c.AbortWithStatusJSON(base.HttpStatusCode(err), base.FailsWithError(c, err))
		return
	} else {
		if siteSettings == nil {
			zap.L().Warn("site settings not found", zap.String("siteId", siteId))
			c.Status(http.StatusNotFound)
			return
		}
		zap.L().Info("found site settings", zap.String("siteId", siteId))
		c.JSON(http.StatusOK, siteSettings)
	}
}

func (h *SiteHandler) SaveSiteSettings(c *gin.Context) {
	siteId := c.Param("siteId")
	siteObjectId := ensureValidId(c, siteId)
	if siteObjectId == nil {
		return
	}
	var siteSettings entity.SiteSettings
	if !bindJson(c, &siteSettings) {
		return
	}

	if siteSettings, err := service.SiteService.SaveSettings(c, &siteSettings); err != nil {
		zap.L().Warn("failed to find site settings", zap.String("siteId", siteId), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, base.FailsWithError(c, err))
	} else {
		if siteSettings == nil {
			zap.L().Warn("site settings not found even though save succeeded", zap.String("siteId", siteId))
			c.Status(http.StatusNotFound)
			return
		}
		zap.L().Info("Site settings saved successfully", zap.Any("siteSettings", siteSettings))
		c.JSON(http.StatusOK, siteSettings)
	}
}

// CreateSite create a site
// @Tags API
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
	if !bindJson(c, &site) {
		return
	}
	currentTime := time.Now()
	site.CreatedTime = &currentTime
	site.UpdatedTime = nil

	h.doCreate(c, &dto.CreateRequest{
		Key:           "site",
		Name:          site.Name,
		Entity:        site,
		Collection:    base.CollectionSite,
		RedisCacheKey: utils.GenKey(base.SiteKeyExistsPrefix, site.Name),
	})
}

func (h *SiteHandler) DeleteSite(c *gin.Context) {
	siteId := c.Param("siteId")

	objectId := h.ensureValidSiteId(c, siteId)
	if objectId == nil {
		return
	}
	if err := service.SiteService.DeleteById(c, *objectId); err != nil {
		zap.L().Warn("failed to delete site", zap.String("siteId", siteId), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, base.FailsWithError(c, err))
		return
	}
	zap.L().Info("site is deleted", zap.String("siteId", siteId))
	c.Status(http.StatusOK)

}

func (h *SiteHandler) ensureValidSiteId(c *gin.Context, siteId string) *primitive.ObjectID {
	objectId := ensureValidId(c, siteId)
	if objectId != nil {
		siteExists, err := service.SiteService.ExistsById(c, *objectId)
		if !siteExists || err != nil {
			zap.L().Warn("site does not exist", zap.String("siteId", siteId), zap.Error(err))
			c.AbortWithStatusJSON(http.StatusBadRequest, base.Fails(c, base.ErrorCode.SiteNotFound))
			return nil
		}
	}
	return objectId
}

// CreateCatalog create a catalog
// @Tags API
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
	if !bindJson(c, &catalog) {
		return
	}
	currentTime := time.Now()
	catalog.CreatedTime = &currentTime
	catalog.UpdatedTime = nil

	//check if the site exists and cache the result
	exists, err := utils.Exists(c, utils.GenKey(base.SiteKeyExistsPrefix, catalog.SiteId.Hex()), func() (any, error) {
		return service.SiteService.ExistsById(c, catalog.SiteId)
	})
	if err != nil {
		zap.L().Warn("failed to check if any sites exist with this siteId", zap.String("siteId", catalog.SiteId.Hex()),
			zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, base.FailsWithError(c, err))
		return
	}
	if !exists {
		zap.L().Warn("no site exists with this siteId", zap.String("siteId", catalog.SiteId.Hex()))
		c.AbortWithStatusJSON(http.StatusBadRequest, base.Fails(c, base.ErrorCode.SiteNotFound))
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

// perform creation
func (h *SiteHandler) doCreate(c *gin.Context, req *dto.CreateRequest) {
	col := system.GetSystem().GetCollection(req.Collection)

	exists, err := utils.Exists(c, req.RedisCacheKey, func() (any, error) {
		return service.CatalogService.ExistsByName(c, req.Name)
	})
	if err != nil {
		zap.L().Warn("failed to check if it exists", zap.Error(err), zap.Any("request", req.Entity))
		c.AbortWithStatusJSON(http.StatusInternalServerError, base.FailsWithError(c, err))
		return
	}

	if exists {
		zap.L().Warn("it's duplicated to save", zap.Any(req.Key, req.Name), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest,
			base.FailsWithParams(c, base.ErrorCode.Duplicated, map[string]string{
				req.Key: req.Name,
			}))
		return
	}

	if obj, err := col.InsertOne(c, req.Entity); err != nil {
		zap.L().Warn("failed to save", zap.Error(err), zap.Any(req.Key, req.Name))
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			base.FailsWithMessage(base.ErrorCode.Unexpected, err.Error()))
		return
	} else {
		c.JSON(http.StatusCreated, obj)
	}
}

// FindSiteById finds a site by id
// @Tags API
// @Summary 通过ID查找Site
// @Description 通过ID查找Site
// @Param siteId path string  true "Site ID"
// @Accept  application/json
// @Produce application/json
// @Success 200 array entity.Site
// @Router /sites/{siteId} [get]
func (h *SiteHandler) FindSiteById(c *gin.Context) {
	siteId := c.Param("siteId")

	objectId := h.ensureValidSiteId(c, siteId)
	if objectId == nil {
		return
	}

	if site, err := service.SiteService.FindById(c, *objectId); err != nil {
		zap.L().Warn("failed to find site", zap.String("siteId", siteId), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			base.FailsWithMessage(base.ErrorCode.Unexpected, err.Error()))
		return
	} else {
		zap.L().Info("found site", zap.Any("site", site))
		c.JSON(http.StatusOK, site)
	}
}

// FindCatalogById find catalog by id
// @Tags API
// @Summary  通过ID查找目录
// @Description 通过ID查找目录
// @Param   catalogId	path   string   true   "目录ID"
// @Accept  application/json
// @Produce application/json
// @Success 200 array entity.Catalog
// @Router /catalogs/{catalogId} [get]
func (h *SiteHandler) FindCatalogById(c *gin.Context) {
	catalogId := c.Param("catalogId")
	objectId := ensureValidId(c, catalogId)

	if objectId != nil {
		if catalog, err := service.CatalogService.FindById(c, *objectId); err != nil {
			zap.L().Warn("failed to find catalog", zap.String("catalogId", catalogId), zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				base.FailsWithMessage(base.ErrorCode.Unexpected, err.Error()))
		} else {
			zap.L().Info("found catalog", zap.Any("catalog", catalog))
			c.JSON(http.StatusOK, catalog)
		}
	}
}
