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

type kxkmCrawler struct {
	redis       *cache.Redis
	mongoClient *db.Mongo
	colly       *colly.Collector
	siteCfg     *base.SiteConfig
	client      *resty.Client
	zhConvertor sat.Dicter
}

func NewKxkmCrawler() *kxkmCrawler {
	sys := base.GetSystem()
	cfg := base.GetSiteConfig(base.Cartoon18)
	if cfg == nil {
		zap.L().Sugar().Warn("Could not find site config", zap.String("siteName", base.SiteNsf))
	}

	collyClient, err := client.NewCollector("", 3)
	if err != nil {
		zap.L().Warn("Could not create collector", zap.Error(err))
	}

	return &kxkmCrawler{
		redis:       sys.RedisClient,
		mongoClient: sys.MongoClient,
		colly:       collyClient,
		siteCfg:     cfg,
		client:      resty.New(),
		zhConvertor: sat.DefaultDict(),
	}
}

func (c kxkmCrawler) CrawlHomePage(ctx context.Context, url string) error {
	//TODO implement me
	panic("implement me")
}

func (c kxkmCrawler) CrawlCatalogPage(ctx context.Context, catalogPageTask *model.CatalogPageTask) ([]model.NovelTask, error) {
	zap.L().Info("[kxkm] Got CatalogPageTask message", zap.String("url", catalogPageTask.Url))
	var novelTasks []model.NovelTask
	cly := c.colly.Clone()
	cly.OnHTML(".product__item__text > h6 > a", func(element *colly.HTMLElement) {
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
	zap.L().Info("[kxkm] the number of novel tasks shall be processed", zap.Int("count", len(novelTasks)))
	return novelTasks, nil
}

func (c kxkmCrawler) CrawlNovelPage(ctx context.Context, novelTask *model.NovelTask, skipSaveIfPresent bool) ([]model.ChapterTask, error) {
	zap.L().Info("[kxkm] Got novel message", zap.String("url", novelTask.Url))
	var createdTime = time.Now()
	var novel = entity.Novel{Attributes: make(map[string]interface{}), CreatedTime: &createdTime}
	var chpTasks []model.ChapterTask
	var novelFolder string

	cly := c.colly.Clone()
	//获取名称
	cly.OnHTML(".anime__details__title  h3", func(element *colly.HTMLElement) {
		novel.Name = strings.TrimSpace(c.zhConvertor.Read(element.Text))
	})

	//获取封面图片
	var coverImageUrl string
	cly.OnHTML(".anime__details__pic.set-bg", func(img *colly.HTMLElement) {
		coverImageUrl = img.Attr("data-setbg")
	})

	//多章节情况：获取每一页上面的chapter内容
	cly.OnHTML(".chapter_list a", func(a *colly.HTMLElement) {
		chapterName := c.zhConvertor.Read(a.Text)
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
		zap.L().Error("[kxkm] no chapters found for novel", zap.String("novelName", novel.Name))
	} else {
		zap.L().Info("[kxkm] number of chapters found for novel", zap.String("novelName", novel.Name),
			zap.Int("number", len(chpTasks)))
	}

	//create directory
	if novelDir, ok := c.siteCfg.Attributes["directory"]; ok {
		novelFolder = filepath.Join(novelDir, novel.Name)

		if fileutil.IsExist(novelFolder) {
			zap.L().Info("[kxkm] duplicated novel and no need to create directory", zap.String("novelName", novel.Name))
		} else {
			err = os.MkdirAll(novelFolder, 0755)
		}

		//下载封面图片
		if err == nil && coverImageUrl != "" {
			destFile := filepath.Join(novelFolder, "cover.jpg")
			if exist := fileutil.IsExist(destFile); !exist {
				client, err := client.GetRestyClient(novelTask.Url, true)
				if err != nil {
					return chpTasks, err
				}
				if _, err = client.R().SetOutput(destFile).Get(coverImageUrl); err != nil {
					metrics.MetricsFailedComicPicTaskGauge.Inc()
					zap.L().Error("[kxkm] failed to download cover picture", zap.String("url", coverImageUrl), zap.Error(err))
					return chpTasks, err
				} else {
					metrics.MetricsComicPicDownloaded.Inc()
					zap.L().Info("[kxkm] cover picture downloaded", zap.String("url", coverImageUrl), zap.String("localFile", destFile))
				}
			}
		}
	}

	return chpTasks, err
}

func (c kxkmCrawler) CrawlChapterPage(ctx context.Context, chapterTask *model.ChapterTask, skipSaveIfPresent bool) error {
	var err error
	var restyClient *resty.Client
	var novel *entity.Novel

	cly := c.colly.Clone()
	zap.L().Info("[kxkm] Got chapter message", zap.String("url", chapterTask.Url))

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

	var fileFormat string
	var i = 0
	cly.OnHTML(".blog__details__content>img", func(img *colly.HTMLElement) {
		i++
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

		fileFormat, err = utils.GetFileExtFromUrl(picUrl)
		if err != nil {
			return
		}
		destFile := filepath.Join(chapterDir, fmt.Sprintf("%04d", i)+fileFormat)

		if fileutil.IsExist(destFile) {
			metrics.MetricsComicPicDownloaded.Inc()
			zap.L().Info("[kxkm] pic skipped since it exists in directory", zap.String("destFile", destFile))
			return
		}

		if _, err = restyClient.R().SetOutput(destFile).Get(picUrl); err != nil {
			metrics.MetricsFailedComicPicTaskGauge.Inc()
			zap.L().Error("[kxkm] failed to download picture", zap.String("url", picUrl), zap.Error(err))
			return
		} else {
			metrics.MetricsComicPicDownloaded.Inc()
			zap.L().Info("[kxkm] picture downloaded", zap.String("url", picUrl), zap.String("localFile", destFile))
		}
	})
	if err = cly.Visit(chapterTask.Url); err != nil {
		return err
	}
	return nil
}
