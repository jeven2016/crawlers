package aipic

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/metrics"
	"crawlers/pkg/model/entity"
	"crawlers/pkg/repository"
	"crawlers/pkg/service"
	"errors"
	"github.com/go-creed/sat"
	"github.com/gocolly/colly/v2"
	"github.com/jeven2016/mylibs/client"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Aipic struct {
	colly       *colly.Collector
	zhConvertor sat.Dicter
}

func NewCartoonCrawler() *Aipic {
	collyClient, err := client.NewCollector("", 3)
	if err != nil {
		zap.L().Warn("Could not create collector", zap.Error(err))
	}

	return &Aipic{
		colly:       collyClient,
		zhConvertor: sat.DefaultDict(),
	}
}

func (c Aipic) CrawlHomePage(ctx context.Context, url string) error {
	//TODO implement me
	panic("implement me")
}

func (c Aipic) CrawlCatalogPage(ctx context.Context, catalogPageTask *entity.CatalogPageTask) ([]entity.NovelTask, error) {
	zap.L().Info("Got CatalogPageTask message", zap.String("url", catalogPageTask.Url))
	var novelTasks []entity.NovelTask
	cly := c.colly.Clone()
	cly.OnHTML(".t_subject a", func(element *colly.HTMLElement) {
		href := "https://www.cool18.com/bbs7/" + element.Attr("href")
		novelTasks = append(novelTasks, entity.NovelTask{
			Url:      href,
			SiteName: catalogPageTask.SiteName,
			Name:     element.Text,
		})
	})

	if err := cly.Visit(catalogPageTask.Url); err != nil {
		return nil, err
	}
	zap.L().Info("the number of novel tasks shall be processed", zap.Int("count", len(novelTasks)))
	return novelTasks, nil
}

func (c Aipic) CrawlNovelPage(ctx context.Context, novelTask *entity.NovelTask, skipSaveIfPresent bool) ([]entity.ChapterTask, error) {
	zap.L().Info("Got novel message", zap.String("url", novelTask.Url))

	siteCfg := service.ConfigService.GetSiteConfig(base.Aipic)
	if siteCfg == nil {
		return nil, errors.New("no site config found for site " + base.Aipic)
	}

	var imageUrls []string
	var createdTime = time.Now()
	var novel = entity.Novel{Attributes: make(map[string]interface{}), CreatedTime: &createdTime}
	cly := c.colly.Clone()

	//获取名称
	name := strings.ReplaceAll(novelTask.Name, "【AI生成】", "")
	name = strings.TrimSpace(name)
	novel.Name = name

	//获取图片链接
	cly.OnHTML(".show_content img[mydatasrc]", func(element *colly.HTMLElement) {
		href := element.Attr("src")
		imageUrls = append(imageUrls, href)
	})

	if err := cly.Visit(novelTask.Url); err != nil {
		return nil, err
	}

	//下载图片

	//create directory
	var dir = ""
	if novelDir, ok := siteCfg.Attributes["directory"]; ok {
		dir = filepath.Join(novelDir, novel.Name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	for _, url := range imageUrls {
		restyClient, err := client.GetRestyClient(url, true)
		if err != nil {
			return nil, err
		}

		index := strings.LastIndex(url, "/") + 1
		filename := url[index:]

		destFile := filepath.Join(dir, filename)
		if _, err = restyClient.R().SetOutput(destFile).Get(url); err != nil {
			metrics.MetricsFailedComicPicTaskGauge.Inc()
			zap.L().Error("failed to download picture", zap.String("url", url), zap.Error(err))
		} else {
			metrics.MetricsComicPicDownloaded.Inc()
			zap.L().Info("picture downloaded", zap.String("url", url), zap.String("localFile", destFile))
		}
	}

	var novelId *primitive.ObjectID
	var err error

	if novelId, err = repository.NovelRepo.FindIdByName(ctx, novel.Name); err != nil {
		return nil, err
	}

	if !skipSaveIfPresent || novelId == nil {
		//保存novel
		if novelId != nil {
			novel.Id = *novelId
		}
		if novelId, err = repository.NovelRepo.Save(ctx, &novel); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (c Aipic) CrawlChapterPage(ctx context.Context, chapterTask *entity.ChapterTask, skipSaveIfPresent bool) error {
	return nil
}
