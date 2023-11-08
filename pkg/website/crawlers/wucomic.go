package crawlers

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/dao"
	"crawlers/pkg/metrics"
	"crawlers/pkg/model"
	"crawlers/pkg/model/entity"
	"fmt"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/go-creed/sat"
	"github.com/go-resty/resty/v2"
	"github.com/gocolly/colly/v2"
	"github.com/jeven2016/mylibs/cache"
	"github.com/jeven2016/mylibs/client"
	"github.com/jeven2016/mylibs/db"
	"github.com/jeven2016/mylibs/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type wucomicCrawler struct {
	redis       *cache.Redis
	mongoClient *db.Mongo
	colly       *colly.Collector
	siteCfg     *base.SiteConfig
	client      *resty.Client
	zhConvertor sat.Dicter
}

func NewWucomicCrawler() *wucomicCrawler {
	sys := base.GetSystem()
	cfg := base.GetSiteConfig(base.Cartoon18)
	if cfg == nil {
		zap.L().Sugar().Warn("Could not find site config", zap.String("siteName", base.SiteNsf))
	}

	collyClient, err := client.NewCollector("", 3)
	if err != nil {
		zap.L().Warn("Could not create collector", zap.Error(err))
	}

	return &wucomicCrawler{
		redis:       sys.RedisClient,
		mongoClient: sys.MongoClient,
		colly:       collyClient,
		siteCfg:     cfg,
		client:      resty.New(),
		zhConvertor: sat.DefaultDict(),
	}
}

func (c wucomicCrawler) CrawlHomePage(ctx context.Context, url string) error {
	//TODO implement me
	panic("implement me")
}

func (c wucomicCrawler) CrawlCatalogPage(ctx context.Context, catalogPageTask *model.CatalogPageTask) ([]model.NovelTask, error) {
	zap.L().Info("[wucomic] Got CatalogPageTask message", zap.String("url", catalogPageTask.Url))
	var novelTasks []model.NovelTask
	cly := c.colly.Clone()
	cly.OnHTML(".cartoon-cover", func(element *colly.HTMLElement) {
		href := element.Attr("href")
		novelUrl := utils.BuildUrl(catalogPageTask.Url, href)
		novelTasks = append(novelTasks, model.NovelTask{
			Url:      novelUrl,
			SiteName: catalogPageTask.SiteName,
		})
	})

	if err := cly.Visit(catalogPageTask.Url); err != nil {
		return nil, err
	}
	zap.L().Info("[wucomic] the number of novel tasks shall be processed", zap.Int("count", len(novelTasks)))
	return novelTasks, nil
}

func (c wucomicCrawler) CrawlNovelPage(ctx context.Context, novelTask *model.NovelTask, skipSaveIfPresent bool) ([]model.ChapterTask, error) {
	zap.L().Info("[wucomic] Got novel message", zap.String("url", novelTask.Url))
	var createdTime = time.Now()
	var novel = entity.Novel{Attributes: make(map[string]interface{}), CreatedTime: &createdTime}
	var chpTasks []model.ChapterTask
	cly := c.colly.Clone()
	//获取名称
	cly.OnHTML(".detail-tit", func(element *colly.HTMLElement) {
		novel.Name = strings.TrimSpace(c.zhConvertor.Read(element.Text))
	})

	//多章节情况：获取每一页上面的chapter内容
	cly.OnHTML(".chapter-container a", func(a *colly.HTMLElement) {
		chapterName := c.zhConvertor.Read(a.Attr("title"))
		chpTask := model.ChapterTask{
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

	if novelId, err = dao.NovelDao.FindIdByName(ctx, novel.Name); err != nil {
		return nil, err
	}

	if !skipSaveIfPresent || novelId == nil {
		//保存novel
		novel.HasChapters = len(chpTasks) > 0
		if novelId != nil {
			novel.Id = *novelId
		}
		if novelId, err = dao.NovelDao.Save(ctx, &novel); err != nil {
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
		zap.L().Error("[wucomic] no chapters found for novel", zap.String("novelName", novel.Name))
	} else {
		zap.L().Info("[wucomic] number of chapters found for novel", zap.String("novelName", novel.Name),
			zap.Int("number", len(chpTasks)))
	}

	//create directory
	if novelDir, ok := c.siteCfg.Attributes["directory"]; ok {
		novelFolder := filepath.Join(novelDir, novel.Name)

		if fileutil.IsExist(novelFolder) {
			zap.L().Info("[wucomic] duplicated novel and no need to create directory", zap.String("novelName", novel.Name))
			return []model.ChapterTask{}, nil
		}
		if err = os.MkdirAll(novelFolder, 0755); err != nil {
			return chpTasks, err
		}
	}

	return chpTasks, nil
}

func (c wucomicCrawler) CrawlChapterPage(ctx context.Context, chapterTask *model.ChapterTask, skipSaveIfPresent bool) error {
	var err error
	var restyClient *resty.Client
	var novel *entity.Novel

	cly := c.colly.Clone()
	zap.L().Info("[wucomic] Got chapter message", zap.String("url", chapterTask.Url))

	if novel, err = dao.NovelDao.FindById(ctx, chapterTask.NovelId); err != nil {
		return err
	}

	//以novel名称为根目录，chapter目录为子目录
	var chapterDir string
	if novelDir, ok := c.siteCfg.Attributes["directory"]; ok {
		chapterDir = filepath.Join(novelDir, novel.Name, chapterTask.Name)
		if err = os.MkdirAll(chapterDir, 0755); err != nil {
			return err
		}
	}

	if chapterDir == "" {
		return fmt.Errorf("no chapter directory specified %v", c.siteCfg.Attributes["directory"])
	}

	var i = 1
	cly.OnHTML(".cropped", func(img *colly.HTMLElement) {
		if err != nil {
			metrics.MetricsFailedComicPicTaskGauge.Inc()
			return
		}

		if i%100 == 0 {
			time.Sleep(4 * time.Second)
		}

		picUrl := img.Attr("src")
		restyClient, err = client.GetRestyClient(picUrl, true)
		if err != nil {
			return
		}

		var fileFormat = ".jpg"
		destFile := filepath.Join(chapterDir, fmt.Sprintf("%04d", i)+fileFormat)
		i++

		if fileutil.IsExist(destFile) {
			metrics.MetricsComicPicDownloaded.Inc()
			zap.L().Info("[wucomic] pic skipped since it exists in directory", zap.String("destFile", destFile))
			return
		}

		if _, err = restyClient.R().SetOutput(destFile).Get(picUrl); err != nil {
			metrics.MetricsFailedComicPicTaskGauge.Inc()
			zap.L().Error("[wucomic] failed to download picture", zap.String("url", picUrl), zap.Error(err))
			return
		} else {
			metrics.MetricsComicPicDownloaded.Inc()
			zap.L().Info("[wucomic] picture downloaded", zap.String("url", picUrl), zap.String("localFile", destFile))
		}
	})
	if err = cly.Visit(chapterTask.Url); err != nil {
		return err
	}
	return nil
}
