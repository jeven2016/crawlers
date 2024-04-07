package stream

import (
	"crawlers/pkg/base"
	"crawlers/pkg/dao"
	"crawlers/pkg/metrics"
	"crawlers/pkg/model/entity"
	"encoding/json"
	"errors"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/fatih/structs"
	"github.com/jeven2016/mylibs/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"reflect"
	"strings"
	"time"
)

type TaskProcessor interface {
	ParsePageUrls(siteName, originPageUrl string) ([]string, error)
	HandleCatalogPageTask(jsonData string) []entity.NovelTask
	HandleNovelTask(jsonData string) []entity.ChapterTask
	HandleChapterTask(jsonData string) interface{}
}

type DefaultTaskProcessor struct{}

func NewTaskProcessor() TaskProcessor {
	return &DefaultTaskProcessor{}
}

// ParsePageUrls parses all page urls based on origin page url which could be combined by multiple pages
func (d DefaultTaskProcessor) ParsePageUrls(siteName, originPageUrl string) ([]string, error) {
	cfg := base.GetSiteConfig(siteName)
	if cfg == nil {
		return nil, errors.New("Could not find site config: " + siteName)
	}
	if cfg.RegexSettings == nil || cfg.RegexSettings.ParsePageRegex == "" {
		zap.L().Info("no RegexSettings setting defined, just return origin url", zap.String("siteName", siteName),
			zap.String("url", originPageUrl))
		return []string{originPageUrl}, nil
	}
	return base.GenPageUrls(cfg.RegexSettings.ParsePageRegex, originPageUrl, cfg.RegexSettings.PagePrefix, "")
}

// HandleCatalogPageTask handles an individual catalog page to get a list of novel pages for further processing
func (d DefaultTaskProcessor) HandleCatalogPageTask(jsonData string) (novelMsgs []entity.NovelTask) {
	zap.L().Info("handle catalogPageTask", zap.String("json", jsonData))

	var catalogPageTask entity.CatalogPageTask
	var err error

	metrics.MetricsRuningCatalogPageTasksGauge.Inc()
	metrics.MetricsTotalCatalogPageTasks.Inc()
	defer func() {
		metrics.MetricsRuningCatalogPageTasksGauge.Dec()
		if err != nil {
			metrics.MetricsFailedCatalogPageTasksGauge.Inc()
		} else {
			zap.L().Info("the count of novel tasks for this catalog page", zap.Int("count", len(novelMsgs)))
			metrics.MetricsSucceedCatalogPageTasksGauge.Inc()
		}
	}()

	//convert the json string to task struct
	if !base.Convert(jsonData, &catalogPageTask) {
		return nil
	}

	cfg := base.GetSiteConfig(catalogPageTask.SiteName)

	//check if to skip specific operations
	var skipIfPresent = getSettingValue[bool](cfg, "CatalogPage", "skipIfPresent", true)
	var skipSaveIfPresent = getSettingValue[bool](cfg, "CatalogPage", "skipSaveIfPresent", true)

	//check if page url is duplicated
	exists, err := isDuplicatedTask(&entity.CatalogPageTask{},
		base.CollectionCatalogPageTask,
		catalogPageTask.Url,
		bson.M{
			base.ColumnUrl: catalogPageTask.Url, //catalogPageTask.Url
		})

	if err != nil {
		zap.L().Warn("error occurs", zap.Error(err))
		return nil
	}
	if exists && skipIfPresent {
		zap.L().Info("catalog page skipped to crawl", zap.String("url", catalogPageTask.Url),
			zap.String("siteName", catalogPageTask.SiteName))
		return nil
	}

	crawler := GetSiteCrawler(catalogPageTask.SiteName)
	if crawler == nil {
		zap.L().Error("site downloader not found", zap.String("SiteName", catalogPageTask.SiteName))
		return nil
	}

	//check if it exists in order to save or update in db
	var existingTask *entity.CatalogPageTask
	if existingTask, err = dao.CatalogPageTaskDao.FindByUrl(base.GetSystemContext(), catalogPageTask.Url); err != nil {
		zap.L().Error("failed to retrieve catalog page task", zap.String("jsonData", jsonData), zap.Error(err))
		return nil
	}

	if novelMsgs, err = crawler.CrawlCatalogPage(base.GetSystemContext(), &catalogPageTask); err != nil {
		zap.L().Warn("CrawlCatalogPage error", zap.String("catalogUrl", catalogPageTask.Url), zap.Error(err))

		//save failed, update the status
		if existingTask != nil {
			if err = convertor.CopyProperties(&catalogPageTask, existingTask); err != nil {
				zap.L().Error("failed to copy properties of catalog page task", zap.Error(err))
				return nil
			}
		}
	}
	updateTaskStatus(&catalogPageTask, existingTask != nil, err == nil)

	if c, ok := catalogPageTask.Attributes["onlyCoverImage"]; ok {
		for i := 0; i < len(novelMsgs); i++ {
			if novelMsgs[i].Attributes == nil {
				novelMsgs[i].Attributes = make(map[string]interface{})
				novelMsgs[i].Attributes["onlyCoverImage"] = c
			}
		}
	}

	if !exists || !skipSaveIfPresent {
		if _, err = dao.CatalogPageTaskDao.Save(base.GetSystemContext(), &catalogPageTask); err != nil {
			zap.L().Error("failed to save catalogPageTask", zap.Error(err))
		}
	} else {
		zap.L().Info("skip saving catalogPageTask", zap.String("url", catalogPageTask.Url),
			zap.String("siteName", catalogPageTask.SiteName))
	}

	return
}

func (d DefaultTaskProcessor) HandleNovelTask(jsonData string) (chapterMessages []entity.ChapterTask) {
	var novelTask entity.NovelTask
	var err error

	metrics.MetricsRuningNovelTasksGauge.Inc()
	metrics.MetricsTotalNovelTasks.Inc()
	defer func() {
		metrics.MetricsRuningNovelTasksGauge.Dec()
		if err != nil {
			metrics.MetricsFailedNovelTasksGauge.Inc()
		} else {
			metrics.MetricsSucceedNovelTasksGauge.Inc()
		}
	}()

	if !base.Convert(jsonData, &novelTask) {
		return nil
	}

	if slice.Contain(base.GetConfig().CrawlerSettings.ExcludedNovelUrls, novelTask.Url) {
		zap.L().Warn("excluded novel url", zap.String("url", novelTask.Url))
		return
	}

	zap.L().Info("handle novel task", zap.String("json", jsonData))

	cfg := base.GetSiteConfig(novelTask.SiteName)

	//whether to skip specific operations
	var skipIfPresent = getSettingValue[bool](cfg, "Novel", "skipIfPresent", true)
	var skipSaveIfPresent = getSettingValue[bool](cfg, "Novel", "skipIfPresent", true)
	var enableChapter = getSettingValue[bool](cfg, "Novel", "enabled", true)

	//check if page url is duplicated
	exists, err := isDuplicatedTask(&entity.NovelTask{},
		base.CollectionNovelTask,
		novelTask.Url,
		bson.M{
			base.ColumnUrl: novelTask.Url, //catalogPageTask.Url
		})
	if err != nil {
		zap.L().Warn("error occurs", zap.Error(err))
		return nil
	}
	if exists && skipIfPresent {
		zap.L().Info("novel skipped to crawl", zap.String("url", novelTask.Url),
			zap.String("name", novelTask.Name), zap.String("siteName", novelTask.SiteName))
		return nil
	}

	crawler := GetSiteCrawler(novelTask.SiteName)
	if crawler == nil {
		zap.L().Error("site crawler not found", zap.String("SiteName", novelTask.SiteName))
		return nil
	}

	//check if it exists in db
	var existingTask *entity.NovelTask
	if existingTask, err = dao.NovelTaskDao.FindByUrl(base.GetSystemContext(), novelTask.Url); err != nil {
		zap.L().Error("failed to retrieve novel page task", zap.String("jsonData", jsonData), zap.Error(err))
		return nil
	}

	currentTime := time.Now()
	if chapterMessages, err = crawler.CrawlNovelPage(base.GetSystemContext(), &novelTask, skipSaveIfPresent); err != nil {
		zap.L().Warn("CrawlNovelPage error", zap.String("novel", novelTask.Url), zap.Error(err))
		//save failed, update the status
		if existingTask != nil {
			if err = convertor.CopyProperties(&novelTask, existingTask); err != nil {
				zap.L().Error("failed to copy properties of novel task", zap.Error(err))
				return nil
			}
			//如果之前重试过，重试次数加1
			if novelTask.Status == base.TaskStatusFailed ||
				novelTask.Status == base.TaskStatusRetryFailed {
				novelTask.Retries++
				novelTask.Status = base.TaskStatusRetryFailed
			}
		} else {
			novelTask.Status = base.TaskStatusFailed
		}
		novelTask.LastUpdated = &currentTime
	} else {
		//已经处理过，记录该url
		novelTask.Status = base.TaskStatusFinished
		novelTask.CreatedDate = &currentTime
	}

	//是否不需处理chapter
	if !enableChapter {
		chapterMessages = nil
	}

	if val, ok := novelTask.Attributes["onlyCoverImage"]; ok && val.(bool) {
		chapterMessages = nil
	}

	if !exists || !skipSaveIfPresent {
		if _, err = dao.NovelTaskDao.Save(base.GetSystemContext(), &novelTask); err != nil {
			zap.L().Error("failed to save novelTask", zap.Error(err))
		}
	} else {
		zap.L().Info("skip saving novelTask", zap.String("url", novelTask.Url),
			zap.String("name", novelTask.Name), zap.String("siteName", novelTask.SiteName))
	}
	return
}

func (d DefaultTaskProcessor) HandleChapterTask(jsonData string) interface{} {
	var chapterTask entity.ChapterTask
	var err error

	metrics.MetricsRuningChapterTasksGauge.Inc()
	metrics.MetricsTotalChapterTasks.Inc()
	defer func() {
		metrics.MetricsRuningChapterTasksGauge.Dec()
		if err != nil {
			metrics.MetricsFailedChapterTasksGauge.Inc()
		} else {
			metrics.MetricsSucceedChapterTasksGauge.Inc()
		}
	}()

	if !base.Convert(jsonData, &chapterTask) {
		return nil
	}
	zap.L().Info("handle chapter task", zap.String("json", jsonData))

	cfg := base.GetSiteConfig(chapterTask.SiteName)

	//whether to skip specific operations
	var skipIfPresent = getSettingValue[bool](cfg, "Chapter", "skipIfPresent", true)
	var skipSaveIfPresent = getSettingValue[bool](cfg, "Chapter", "skipSaveIfPresent", true)
	var enableChapter = getSettingValue[bool](cfg, "Chapter", "enabled", true)

	//check if page url is duplicated
	exists, err := isDuplicatedTask(&entity.ChapterTask{},
		base.CollectionChapterTask,
		chapterTask.Url,
		bson.M{
			base.ColumnUrl: chapterTask.Url, //catalogPageTask.Url
		})
	if err != nil {
		zap.L().Warn("error occurs", zap.Error(err))
		return nil
	}
	if exists && skipIfPresent {
		zap.L().Warn("chapter skipped to crawl", zap.String("jsonData", jsonData))
		return nil
	}

	downloader := GetSiteCrawler(chapterTask.SiteName)
	if downloader == nil {
		zap.L().Error("site downloader not found", zap.String("SiteName", chapterTask.SiteName))
		return nil
	}

	//check if it exists in db
	var existingTask *entity.ChapterTask
	if existingTask, err = dao.ChapterTaskDao.FindByUrl(base.GetSystemContext(), chapterTask.Url); err != nil {
		zap.L().Error("failed to retrieve chapter page task", zap.String("jsonData", jsonData), zap.Error(err))
		return nil
	}

	var start int
	//var enabledRetry bool

	currentTime := time.Now()
	if err = downloader.CrawlChapterPage(base.GetSystemContext(), &chapterTask, skipSaveIfPresent); err != nil {
		zap.L().Error("error occurred while downloading", zap.String("url", chapterTask.Url), zap.Error(err))

		if strings.Contains(err.Error(), "Too Many Requests") {
			//enabledRetry = true
			start++
			zap.L().Error("will retry", zap.String("chapterUrl", chapterTask.Url), zap.String("chapterName", chapterTask.Name))
			time.Sleep(3 * time.Second)
		}

		//save failed, update the status
		if existingTask != nil {
			if err = convertor.CopyProperties(&chapterTask, existingTask); err != nil {
				zap.L().Error("failed to copy properties of catalog page task", zap.Error(err))
				return nil
			}
			//如果之前重试过，重试次数加1
			if chapterTask.Status == base.TaskStatusFailed ||
				chapterTask.Status == base.TaskStatusRetryFailed {
				chapterTask.Retries++
				chapterTask.Status = base.TaskStatusRetryFailed
			}
		} else {
			chapterTask.Status = base.TaskStatusFailed
		}
		chapterTask.LastUpdated = &currentTime
	} else {
		//已经处理过，记录该url
		chapterTask.Status = base.TaskStatusFinished
		chapterTask.CreatedDate = &currentTime
		//break
	}
	//}

	if (!exists || !skipSaveIfPresent) && enableChapter {
		if _, err = dao.ChapterTaskDao.Save(base.GetSystemContext(), &chapterTask); err != nil {
			zap.L().Error("failed to save chapterTask", zap.Error(err))
		}
	} else {
		zap.L().Info("skip saving chapter", zap.String("url", chapterTask.Url),
			zap.String("name", chapterTask.Name), zap.String("siteName", chapterTask.SiteName))
	}

	return nil
}

// 检查是否已经处理过的url
func isDuplicatedTask[T any](task *T, collectionName,
	url string, bsonFilter bson.M) (bool /*existence*/, error /*interrupted*/) {

	jsonString, err := utils.GetAndSet(base.GetSystemContext(), url, func() (*string, error) {
		data, err := dao.FindOneByFilter(base.GetSystemContext(),
			bsonFilter, collectionName, task, &options.FindOneOptions{})
		if err != nil || data == nil {
			return nil, err
		} else {
			taskString := convertor.ToString(data)
			if taskString == "" {
				return nil, nil
			}
			return &taskString, nil
		}
	})

	if err != nil || jsonString == nil {
		return false, err
	}
	if err = json.Unmarshal([]byte(*jsonString), task); err != nil {
		return false, err
	}

	//泛型对象使用接口调用其方法
	if res, ok := interface{}(task).(entity.Resource); ok {
		return res.GetStatus() == base.TaskStatusFinished, err
	}
	return false, err
}

func getSettingValue[T any](cfg *base.SiteConfig, mapField, mapKey string, defaultValue T) T {
	if cfg != nil && cfg.CrawlerSettings != nil {
		s := structs.New(cfg.CrawlerSettings)

		if field, ok := s.FieldOk(mapField); ok {
			kind := field.Kind()
			if kind == reflect.Struct {
				if valueField, ok := field.FieldOk(mapKey); ok {
					return valueField.Value().(T)
				}
			}
			if kind == reflect.Map {
				m := field.Value().(map[string]any)
				if v, ok := m[mapKey]; ok {
					return v.(T)
				}
			}
		}
	}
	return defaultValue
}

// update the value of fields of task
func updateTaskStatus[T any](task *T, taskExists bool, succeed bool) {
	t := reflect.ValueOf(task)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	statusField := t.FieldByName("Status")
	retriesField := t.FieldByName("Retries")
	lastUpdatedField := t.FieldByName("LastUpdated")
	createdDateField := t.FieldByName("CreatedDate")

	statusValue := base.TaskStatus(statusField.Int())
	curTime := time.Now()
	curTimeValue := reflect.ValueOf(&curTime)

	if taskExists {
		lastUpdatedField.Set(curTimeValue)
	} else {
		createdDateField.Set(curTimeValue)
	}

	if succeed {
		statusField.Set(reflect.ValueOf(base.TaskStatusFinished))
	} else {
		if taskExists {
			//increase retires
			if statusValue == base.TaskStatusFailed || statusValue == base.TaskStatusRetryFailed {
				u := retriesField.Uint() + 1
				retriesField.SetUint(u)
			}
			statusField.Set(reflect.ValueOf(base.TaskStatusRetryFailed))
		} else {
			statusField.Set(reflect.ValueOf(base.TaskStatusFailed))
		}
	}
}
