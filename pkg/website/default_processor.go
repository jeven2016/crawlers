package website

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/dao"
	"crawlers/pkg/metrics"
	"crawlers/pkg/model"
	"encoding/json"
	"errors"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/jeven2016/mylibs/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"strings"
	"time"
)

type TaskProcessor interface {
	ParsePageUrls(siteName, originPageUrl string) ([]string, error)
	HandleCatalogPageTask(jsonData string) []model.NovelTask
	HandleNovelTask(jsonData string) []model.ChapterTask
	HandleChapterTask(jsonData string) interface{}
}

type DefaultTaskProcessor struct{}

func NewTaskProcessor() TaskProcessor {
	return &DefaultTaskProcessor{}
}

func (d DefaultTaskProcessor) ParsePageUrls(siteName, originPageUrl string) ([]string, error) {
	cfg := base.GetSiteConfig(siteName)
	if cfg == nil {
		zap.L().Sugar().Error("Could not find site config", zap.String("siteName", siteName))
		return nil, errors.New("Could not find site config: " + siteName)
	}
	if cfg.RegexSettings == nil || cfg.RegexSettings.ParsePageRegex == "" {
		zap.L().Info("no RegexSettings setting defined, just return origin url", zap.String("siteName", siteName),
			zap.String("url", originPageUrl))
		return []string{originPageUrl}, nil
	}
	return base.GenPageUrls(cfg.RegexSettings.ParsePageRegex, originPageUrl, cfg.RegexSettings.PagePrefix, "")
}

func (d DefaultTaskProcessor) HandleCatalogPageTask(jsonData string) (novelMsgs []model.NovelTask) {
	var catalogPageTask model.CatalogPageTask
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

	if err = json.Unmarshal([]byte(jsonData), &catalogPageTask); err != nil {
		zap.L().Error("json decode", zap.Error(err), zap.String("data", jsonData))
		return nil
	}
	zap.L().Info("handle catalogPageTask task", zap.String("json", jsonData))

	cfg := base.GetSiteConfig(catalogPageTask.SiteName)

	//whether to skip specific operations
	var skipNovelIfPresent = true
	var skipSaveIfPresent = true
	if cfg != nil && cfg.CrawlerSettings != nil && cfg.CrawlerSettings.CatalogPage != nil {
		if val, ok := cfg.CrawlerSettings.CatalogPage["skipIfPresent"]; ok {
			skipNovelIfPresent = val.(bool)
		}
		if val, ok := cfg.CrawlerSettings.CatalogPage["skipSaveIfPresent"]; ok {
			skipSaveIfPresent = val.(bool)
		}
	}

	//check if page url is duplicated
	exists, err := d.isDuplicatedCatalogPageTask(&model.CatalogPageTask{},
		base.CollectionCatalogPageTask,
		catalogPageTask.Url,
		bson.M{
			base.ColumnUrl: catalogPageTask.Url, //catalogPageTask.Url
		})

	if err != nil {
		zap.L().Warn("error occurs", zap.Error(err))
		return nil
	}
	if exists && skipNovelIfPresent {
		zap.L().Info("catalog page skipped to crawl", zap.String("url", catalogPageTask.Url),
			zap.String("siteName", catalogPageTask.SiteName))
		return nil
	}

	crawler := GetSiteCrawler(catalogPageTask.SiteName)
	if crawler == nil {
		zap.L().Error("site downloader not found", zap.String("SiteName", catalogPageTask.SiteName))
		return nil
	}

	//check if it exists in db
	var existingTask *model.CatalogPageTask
	if existingTask, err = dao.CatalogPageTaskDao.FindByUrl(context.Background(), catalogPageTask.Url); err != nil {
		zap.L().Error("failed to retrieve catalog page task", zap.String("jsonData", jsonData), zap.Error(err))
		return nil
	}

	currentTime := time.Now()
	if novelMsgs, err = crawler.CrawlCatalogPage(context.Background(), &catalogPageTask); err != nil {
		zap.L().Warn("CrawlCatalogPage error", zap.String("catalogUrl", catalogPageTask.Url), zap.Error(err))

		//save failed, update the status
		if existingTask != nil {
			if err = convertor.CopyProperties(&catalogPageTask, existingTask); err != nil {
				zap.L().Error("failed to copy properties of catalog page task", zap.Error(err))
				return nil
			}
			//如果之前重试过，重试次数加1
			if catalogPageTask.Status == base.TaskStatusFailed ||
				catalogPageTask.Status == base.TaskStatusRetryFailed {
				catalogPageTask.Retries++
				catalogPageTask.Status = base.TaskStatusRetryFailed
			}
		} else {
			catalogPageTask.Status = base.TaskStatusFailed
		}
		catalogPageTask.LastUpdated = &currentTime
	} else {
		//已经处理过，记录该url
		catalogPageTask.Status = base.TaskStatusFinished
		catalogPageTask.CreatedDate = &currentTime
	}

	if c, ok := catalogPageTask.Attributes["onlyCoverImage"]; ok {
		for i := 0; i < len(novelMsgs); i++ {
			if novelMsgs[i].Attributes == nil {
				novelMsgs[i].Attributes = make(map[string]interface{})
				novelMsgs[i].Attributes["onlyCoverImage"] = c
			}
		}
	}

	if !exists || !skipSaveIfPresent {
		if _, err = dao.CatalogPageTaskDao.Save(context.Background(), &catalogPageTask); err != nil {
			zap.L().Error("failed to save catalogPageTask", zap.Error(err))
		}
	} else {
		zap.L().Info("skip saving catalogPageTask", zap.String("url", catalogPageTask.Url),
			zap.String("siteName", catalogPageTask.SiteName))
	}

	return
}

func (d DefaultTaskProcessor) HandleNovelTask(jsonData string) (chapterMessages []model.ChapterTask) {
	var novelTask model.NovelTask
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

	if err = json.Unmarshal([]byte(jsonData), &novelTask); err != nil {
		zap.L().Error("json decode", zap.Error(err), zap.String("data", jsonData))
		return
	}

	if slice.Contain(base.GetConfig().CrawlerSettings.ExcludedNovelUrls, novelTask.Url) {
		zap.L().Warn("excluded novel url", zap.String("url", novelTask.Url))
		return
	}

	zap.L().Info("handle novel task", zap.String("json", jsonData))

	cfg := base.GetSiteConfig(novelTask.SiteName)

	//whether to skip specific operations
	var skipNovelIfPresent = true
	var skipSaveIfPresent = true
	var enableChapter = true
	if cfg != nil && cfg.CrawlerSettings != nil {
		if cfg.CrawlerSettings.Novel != nil {
			if val, ok := cfg.CrawlerSettings.Novel["skipIfPresent"]; ok {
				skipNovelIfPresent = val.(bool)
			}
			if val, ok := cfg.CrawlerSettings.Novel["skipSaveIfPresent"]; ok {
				skipSaveIfPresent = val.(bool)
			}
		}
		if cfg.CrawlerSettings.Chapter != nil {
			if val, ok := cfg.CrawlerSettings.Chapter["enabled"]; ok {
				enableChapter = val.(bool)
			}
		}
	}

	//check if page url is duplicated
	exists, err := d.isDuplicatedNovelTask(&model.NovelTask{},
		base.CollectionNovelTask,
		novelTask.Url,
		bson.M{
			base.ColumnUrl: novelTask.Url, //catalogPageTask.Url
		})
	if err != nil {
		zap.L().Warn("error occurs", zap.Error(err))
		return nil
	}
	if exists && skipNovelIfPresent {
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
	var existingTask *model.NovelTask
	if existingTask, err = dao.NovelTaskDao.FindByUrl(context.Background(), novelTask.Url); err != nil {
		zap.L().Error("failed to retrieve novel page task", zap.String("jsonData", jsonData), zap.Error(err))
		return nil
	}

	currentTime := time.Now()
	if chapterMessages, err = crawler.CrawlNovelPage(context.Background(), &novelTask, skipSaveIfPresent); err != nil {
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
		if _, err = dao.NovelTaskDao.Save(context.Background(), &novelTask); err != nil {
			zap.L().Error("failed to save novelTask", zap.Error(err))
		}
	} else {
		zap.L().Info("skip saving novelTask", zap.String("url", novelTask.Url),
			zap.String("name", novelTask.Name), zap.String("siteName", novelTask.SiteName))
	}
	return
}

func (d DefaultTaskProcessor) HandleChapterTask(jsonData string) interface{} {
	var chapterTask model.ChapterTask
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

	if err = json.Unmarshal([]byte(jsonData), &chapterTask); err != nil {
		zap.L().Error("json decode", zap.Error(err), zap.String("data", jsonData))
		return nil
	}
	zap.L().Info("handle chapter task", zap.String("json", jsonData))

	cfg := base.GetSiteConfig(chapterTask.SiteName)

	//whether to skip specific operations
	var skipIfPresent = true
	var skipSaveIfPresent = true
	var enableChapter = true
	if cfg != nil && cfg.CrawlerSettings != nil && cfg.CrawlerSettings.Chapter != nil {
		if val, ok := cfg.CrawlerSettings.Chapter["skipIfPresent"]; ok {
			skipIfPresent = val.(bool)
		}
		if val, ok := cfg.CrawlerSettings.Chapter["skipSaveIfPresent"]; ok {
			skipSaveIfPresent = val.(bool)
		}
		if val, ok := cfg.CrawlerSettings.Chapter["enabled"]; ok {
			enableChapter = val.(bool)
		}
	}

	//check if page url is duplicated
	exists, err := d.isDuplicatedChapterTask(&model.ChapterTask{},
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
	var existingTask *model.ChapterTask
	if existingTask, err = dao.ChapterTaskDao.FindByUrl(context.Background(), chapterTask.Url); err != nil {
		zap.L().Error("failed to retrieve chapter page task", zap.String("jsonData", jsonData), zap.Error(err))
		return nil
	}

	var start int
	var enabledRetry bool
	for {
		if enabledRetry && start >= base.DefaultRetries {
			zap.L().Error("failed to retry for multiple times", zap.String("chapterUrl", chapterTask.Url),
				zap.String("chapterName", chapterTask.Name))
			break
		}
		currentTime := time.Now()
		if err = downloader.CrawlChapterPage(context.Background(), &chapterTask, skipSaveIfPresent); err != nil {
			zap.L().Error("error occurred while downloading", zap.String("url", chapterTask.Url), zap.Error(err))

			if strings.Contains(err.Error(), "Too Many Requests") {
				enabledRetry = true
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
			break
		}
	}

	if (!exists || !skipSaveIfPresent) && enableChapter {
		if _, err = dao.ChapterTaskDao.Save(context.Background(), &chapterTask); err != nil {
			zap.L().Error("failed to save chapterTask", zap.Error(err))
		}
	} else {
		zap.L().Info("skip saving chapter", zap.String("url", chapterTask.Url),
			zap.String("name", chapterTask.Name), zap.String("siteName", chapterTask.SiteName))
	}

	return nil
}

// 检查是否已经处理过的url, 状态是finished状态
func (t DefaultTaskProcessor) isDuplicatedNovelTask(cpTask *model.NovelTask, collectionName,
	url string, bsonFilter bson.M) (bool /*existence*/, error /*interrupted*/) {
	jsonString, err := utils.GetAndSet(context.Background(), url, func() (*string, error) {
		if data, err := dao.FindByMongoFilter(context.Background(), bsonFilter, collectionName, cpTask, &options.FindOneOptions{}); err != nil {
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
	if err = json.Unmarshal([]byte(*jsonString), cpTask); err != nil {
		return false, err
	}

	return cpTask.GetStatus() == base.TaskStatusFinished, err
}

// 检查是否已经处理过的url
func (t DefaultTaskProcessor) isDuplicatedChapterTask(cpTask *model.ChapterTask, collectionName,
	url string, bsonFilter bson.M) (bool /*existence*/, error /*interrupted*/) {
	jsonString, err := utils.GetAndSet(context.Background(), url, func() (*string, error) {
		if data, err := dao.FindByMongoFilter(context.Background(), bsonFilter, collectionName,
			cpTask, &options.FindOneOptions{}); err != nil {
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
	if err = json.Unmarshal([]byte(*jsonString), cpTask); err != nil {
		return false, err
	}

	return cpTask.GetStatus() == base.TaskStatusFinished, err
}

// 检查是否已经处理过的url
func (t DefaultTaskProcessor) isDuplicatedCatalogPageTask(cpTask *model.CatalogPageTask, collectionName,
	url string, bsonFilter bson.M) (bool /*existence*/, error /*interrupted*/) {

	jsonString, err := utils.GetAndSet(context.Background(), url, func() (*string, error) {
		if data, err := dao.FindByMongoFilter(context.Background(), bsonFilter, collectionName, cpTask, &options.FindOneOptions{}); err != nil {
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
	if err = json.Unmarshal([]byte(*jsonString), cpTask); err != nil {
		return false, err
	}
	return (model.Resource(cpTask)).GetStatus() == base.TaskStatusFinished, err
}
