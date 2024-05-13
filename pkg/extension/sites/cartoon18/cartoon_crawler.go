package cartoon18

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/metrics"
	"crawlers/pkg/model/entity"
	"crawlers/pkg/repository"
	"crawlers/pkg/service"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/go-creed/sat"
	"github.com/go-resty/resty/v2"
	"github.com/gocolly/colly/v2"
	"github.com/jeven2016/mylibs/client"
	"github.com/jeven2016/mylibs/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CartoonCrawler struct {
	colly       *colly.Collector
	zhConvertor sat.Dicter
}

func NewCartoonCrawler() *CartoonCrawler {
	collyClient, err := client.NewCollector("", 3)
	if err != nil {
		zap.L().Warn("Could not create collector", zap.Error(err))
	}

	return &CartoonCrawler{
		colly:       collyClient,
		zhConvertor: sat.DefaultDict(),
	}
}

func (c CartoonCrawler) CrawlHomePage(ctx context.Context, url string) error {
	//TODO implement me
	panic("implement me")
}

func (c CartoonCrawler) CrawlCatalogPage(ctx context.Context, catalogPageTask *entity.CatalogPageTask) ([]entity.NovelTask, error) {
	zap.L().Info("Got CatalogPageTask message", zap.String("url", catalogPageTask.Url))
	var novelTasks []entity.NovelTask
	cly := c.colly.Clone()
	cly.OnHTML(".card .lines.lines-2 a.visited", func(element *colly.HTMLElement) {
		href := element.Attr("href")
		novelUrl := utils.BuildUrl(catalogPageTask.Url, href)
		novelTasks = append(novelTasks, entity.NovelTask{
			Url:      novelUrl,
			SiteName: catalogPageTask.SiteName,
		})
	})

	if err := cly.Visit(catalogPageTask.Url); err != nil {
		return nil, err
	}
	zap.L().Info("the number of novel tasks shall be processed", zap.Int("count", len(novelTasks)))
	return novelTasks, nil
}

func (c CartoonCrawler) CrawlNovelPage(ctx context.Context, novelTask *entity.NovelTask, skipSaveIfPresent bool) ([]entity.ChapterTask, error) {
	zap.L().Info("Got novel message", zap.String("url", novelTask.Url))

	siteCfg := service.ConfigService.GetSiteConfig(base.Cartoon18)
	if siteCfg == nil {
		return nil, errors.New("no site config found for site " + base.Cartoon18)
	}

	var createdTime = time.Now()
	var novel = entity.Novel{Attributes: make(map[string]interface{}), CreatedTime: &createdTime}
	var chpTasks []entity.ChapterTask
	cly := c.colly.Clone()
	//获取名称
	cly.OnHTML(".title.py-1", func(element *colly.HTMLElement) {
		name := c.zhConvertor.Read(element.Text)
		name = strings.Split(name, "\t\n\t\t")[0]
		name = strings.ReplaceAll(name, "\n\t", "")
		name = strings.TrimSpace(name)

		if strings.Contains(name, "/") {
			name = strings.Split(name, "/")[1]
		}
		novel.Name = name
	})

	//只有单章的情况
	cly.OnHTML(".btn.btn-primary.mr-2.mb-2", func(a *colly.HTMLElement) {
		chapterName := c.zhConvertor.Read(a.Text)
		chpTask := entity.ChapterTask{
			Name:     chapterName,
			SiteName: novelTask.SiteName,
			Url:      utils.BuildUrl(novelTask.Url, a.Attr("href")),
		}
		chpTasks = append(chpTasks, chpTask)
	})

	//多章节情况：获取每一页上面的chapter内容
	cly.OnHTML(".btn.btn-info.mr-2.mb-2", func(a *colly.HTMLElement) {
		chapterName := c.zhConvertor.Read(a.Text)
		chpTask := entity.ChapterTask{
			Name:     chapterName,
			SiteName: novelTask.SiteName,
			Url:      utils.BuildUrl(novelTask.Url, a.Attr("href")),
		}
		chpTasks = append(chpTasks, chpTask)
	})

	if err := cly.Visit(novelTask.Url); err != nil {
		return nil, err
	}

	var novelId *primitive.ObjectID
	var err error

	if novelId, err = repository.NovelDao.FindIdByName(ctx, novel.Name); err != nil {
		return nil, err
	}

	if !skipSaveIfPresent || novelId == nil {
		//保存novel
		novel.HasChapters = len(chpTasks) > 0
		if novelId != nil {
			novel.Id = *novelId
		}
		if novelId, err = repository.NovelDao.Save(ctx, &novel); err != nil {
			return nil, err
		}
	}

	if novelId != nil {
		for i := 0; i < len(chpTasks); i++ {
			chpTasks[i].NovelId = *novelId
			chpTasks[i].Order = i + 1
		}
	}

	if len(chpTasks) == 0 {
		zap.L().Error("no chapters found for novel", zap.String("novelName", novel.Name))
	} else {
		zap.L().Info("number of chapters found for novel", zap.String("novelName", novel.Name),
			zap.Int("number", len(chpTasks)))
	}

	//create directory
	if novelDir, ok := siteCfg.Attributes["directory"]; ok {
		if err = os.MkdirAll(filepath.Join(novelDir, novel.Name), 0755); err != nil {
			return chpTasks, err
		}
	}

	return chpTasks, nil
}

func (c CartoonCrawler) CrawlChapterPage(ctx context.Context, chapterTask *entity.ChapterTask, skipSaveIfPresent bool) error {
	var err error
	var restyClient *resty.Client
	var novel *entity.Novel

	siteCfg := service.ConfigService.GetSiteConfig(base.Cartoon18)
	if siteCfg == nil {
		return errors.New("no site config found for site " + base.Cartoon18)
	}

	cly := c.colly.Clone()
	zap.L().Info("Got chapter message", zap.String("url", chapterTask.Url))

	if novel, err = repository.NovelDao.FindById(ctx, chapterTask.NovelId); err != nil {
		return err
	}

	//以novel名称为根目录，chapter目录为子目录
	var chapterDir string
	if novelDir, ok := siteCfg.Attributes["directory"]; ok {
		chapterDir = filepath.Join(novelDir, novel.Name, chapterTask.Name)
		if err = os.MkdirAll(chapterDir, 0755); err != nil {
			return err
		}
	}

	if chapterDir == "" {
		return fmt.Errorf("no chapter directory specified %v", siteCfg.Attributes["directory"])
	}

	var i = 1
	cly.OnHTML(".cartoon-image", func(img *colly.HTMLElement) {
		if err != nil {
			metrics.MetricsFailedComicPicTaskGauge.Inc()
			return
		}

		if i%100 == 0 {
			time.Sleep(4 * time.Second)
		}

		picUrl := img.Attr("data-src")
		restyClient, err = client.GetRestyClient(picUrl, true)
		if err != nil {
			return
		}

		var fileFormat = ".webp"
		if !strings.Contains(picUrl, ".webp") {
			fileFormat = ".jpg"
		}

		destFile := filepath.Join(chapterDir, fmt.Sprintf("%04d", i)+fileFormat)
		i++

		if fileutil.IsExist(destFile) {
			metrics.MetricsComicPicDownloaded.Inc()
			zap.L().Info("pic skipped since it exists in directory", zap.String("destFile", destFile))
			return
		}

		if _, err = restyClient.R().SetOutput(destFile).Get(picUrl); err != nil {
			metrics.MetricsFailedComicPicTaskGauge.Inc()
			zap.L().Error("failed to download picture", zap.String("url", picUrl), zap.Error(err))
			return
		} else {
			metrics.MetricsComicPicDownloaded.Inc()
			zap.L().Info("picture downloaded", zap.String("url", picUrl), zap.String("localFile", destFile))
		}
	})
	if err = cly.Visit(chapterTask.Url); err != nil {
		return err
	}
	return nil
}
