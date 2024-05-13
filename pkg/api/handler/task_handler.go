package handler

import (
	"crawlers/pkg/base"
	"crawlers/pkg/model/entity"
	"crawlers/pkg/repository"
	"crawlers/pkg/service"
	"crawlers/pkg/stream"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/jeven2016/mylibs/system"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

// TaskHandler task handler for submitting tasks to message stream
type TaskHandler struct{}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{}
}

func (h *TaskHandler) FindTasksOfCatalogPage(c *gin.Context) {
	catalogId := c.Query("catalogId")
	objectId := ensureValidId(c, catalogId)

	if objectId == nil {
		return
	}
	if catalogs, err := repository.CatalogPageTaskDao.FindTasksByCatalogId(c, *objectId); err != nil {
		zap.L().Warn("failed to find catalogPage tasks", zap.String("catalogId", catalogId), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			base.FailsWithMessage(base.ErrCodeUnknown, err.Error()))
		return
	} else {
		c.JSON(http.StatusOK, catalogs)
	}
}

func (h *TaskHandler) FindTasksOfNovel(c *gin.Context) {
	catalogId := c.Query("catalogId")
	objectId := ensureValidId(c, catalogId)

	if objectId == nil {
		return
	}
	if novelTasks, err := repository.NovelTaskDao.FindByCatalogId(c, *objectId); err != nil {
		zap.L().Warn("failed to find novel tasks", zap.String("catalogId", catalogId), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			base.FailsWithMessage(base.ErrCodeUnknown, err.Error()))
		return
	} else {
		c.JSON(http.StatusOK, novelTasks)
	}
}

// CreateCatalogPageTask handler for catalog page request and to parse the novel links for further processing
// @Tags API
// @Summary  处理目录页面请求
// @Description 处理目录页面请求,解析出Novel的地址并发送到消息对列中去
// @Param   request 	body    entity.CatalogPageTask   true   "目录ID"
// @Accept  application/json
// @Produce application/json
// @Success 200
// @Router /tasks/catalog-pages [post]
func (h *TaskHandler) CreateCatalogPageTask(c *gin.Context) {
	var pageTask entity.CatalogPageTask
	if !bindJson(c, &pageTask) {
		return
	}

	var sp stream.TaskProcessor
	var site *entity.Site
	var hasError bool
	var urls []string
	var err error

	if site, hasError = h.getTaskEntity(c, pageTask.CatalogId); hasError {
		return
	}

	//if multiple pages need to handle
	if sp = stream.GetSiteTaskProcessor(site.Name); sp == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, base.Fails(base.ErrCodeUnSupportedCatalog))
		zap.L().Warn("no processor found for this siteKey", zap.String("siteKey", site.Name))
		return
	}
	//parse all page urls if page parameter is specified in such format: page=1-5
	urls, err = sp.ParsePageUrls(site.Name, pageTask.Url)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, base.FailsWithParams(base.ErrParsePageUrl, err.Error()))
		zap.L().Warn("failed to process pageUrl", zap.String("pageUrl", pageTask.Url), zap.Error(err))
		return
	}

	// publish corresponding messages for these urls
	for _, url := range urls {
		if url == "" {
			zap.L().Warn("invalid page url", zap.String("pageUrl", url))
			continue
		}

		//construct  a catalog page message
		pageMsg := &entity.CatalogPageTask{
			SiteName:   site.Name,
			CatalogId:  pageTask.CatalogId,
			Url:        url,
			Attributes: pageTask.Attributes,
			Status:     base.TaskStatusNotStared,
		}

		//publish it
		if err = system.GetSystem().RedisClient.PublishMessage(c, pageMsg, stream.CatalogPageUrlStream); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				base.FailsWithParams(base.ErrPublishMessage, err.Error()))
			zap.L().Warn("failed to publish a message",
				zap.String("pageUrl", pageTask.Url), zap.Error(err))
			return
		}
	}
	zap.S().Info("published", strconv.Itoa(len(urls)), "task messages for catalog page:", pageTask.Url)
	c.JSON(http.StatusAccepted, base.SuccessCode(base.ErrCodeTaskSubmitted))
}

// check if both site and catalog exist
func (h *TaskHandler) getTaskEntity(c *gin.Context, catalogId primitive.ObjectID) (site *entity.Site, hasError bool) {
	var err error
	var catalog *entity.Catalog
	catalogStringId := catalogId.Hex()
	siteStringId := catalogId.Hex()
	if catalog, err = repository.CatalogDao.FindById(c, catalogId); err != nil {
		zap.L().Warn("catalog does not exist", zap.String("catalogId", catalogStringId), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, base.FailsWithParams(base.ErrCatalogNotFound, catalogStringId))
		hasError = true
		return
	}
	if site, err = repository.SiteDao.FindById(c, catalog.SiteId); err != nil {
		zap.L().Warn("site does not exist", zap.String("siteId", siteStringId), zap.Error(err))
		c.AbortWithStatusJSON(http.StatusBadRequest, base.FailsWithParams(base.ErrSiteNotFound, siteStringId))
		hasError = true
	}
	return
}

// CreateNovelPageTask handle for novel page request and parse the chapter links for further processing
// @Tags API
// @Summary  处理Novel页面请求
// @Description 处理Novel页面请求,解析出章节的地址并发送到消息对列中去
// @Param   request	body   entity.NovelTask   true   "Novel Task"
// @Accept  application/json
// @Produce application/json
// @Success 200
// @Router /tasks/novels [post]
func (h *TaskHandler) CreateNovelPageTask(c *gin.Context) {
	var novelTask entity.NovelTask
	if !bindJson(c, &novelTask) {
		return
	}

	var site *entity.Site
	var hasError bool

	//check if the url is excluded
	if slice.Contain(service.ConfigService.GetConfig().CrawlerSettings.ExcludedNovelUrls, novelTask.Url) {
		zap.L().Warn("excluded novel url", zap.String("url", novelTask.Url))
		c.AbortWithStatusJSON(http.StatusBadRequest, base.Fails(base.ErrExcludedNovel))
		return
	}

	if site, hasError = h.getTaskEntity(c, novelTask.CatalogId); hasError {
		return
	}
	novelTask.Status = base.TaskStatusNotStared
	novelTask.SiteName = site.Name

	if err := system.GetSystem().RedisClient.PublishMessage(c, novelTask, stream.NovelUrlStream); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError,
			base.FailsWithParams(base.ErrPublishMessage, err.Error()))
		zap.L().Warn("failed to publish a message", zap.String("pageUrl", novelTask.Url), zap.Error(err))
		return
	}
}

func (h *TaskHandler) DeleteNovelPageTasks(c *gin.Context) {
	ids := c.Query("idArray")

	if ids == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, base.FailsWithParams(base.ErrRequired, "idArray"))
		return
	}
	idArray := strings.Split(ids, ",")

	if err := service.NovelService.DeleteByIds(c, idArray); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, base.FailsWithParams(base.ErrCodeUnknown, err.Error()))
		zap.L().Warn("failed to delete novel tasks", zap.Any("request", ids), zap.Error(err))
		return
	}
	c.Status(http.StatusNoContent)
}
