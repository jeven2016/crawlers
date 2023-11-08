package onej

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model"
	"encoding/base64"
	"github.com/go-resty/resty/v2"
	"github.com/gocolly/colly/v2"
	"github.com/jeven2016/mylibs/cache"
	"github.com/jeven2016/mylibs/client"
	"github.com/jeven2016/mylibs/db"
	"github.com/jeven2016/mylibs/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"strings"
)

type SiteOnej struct {
	redis       *cache.Redis
	mongoClient *db.Mongo
	logger      *zap.Logger
	colly       *colly.Collector
	siteCfg     *base.SiteConfig
	client      *resty.Client
}

func NewSiteOnej() *SiteOnej {
	sys := base.GetSystem()
	cfg := base.GetSiteConfig(base.SiteOneJ)
	if cfg == nil {
		zap.L().Sugar().Warn("Could not find site config", zap.String("siteName", base.SiteNsf))
	}

	collyClient, err := client.NewCollector("", 3)
	if err != nil {
		zap.L().Warn("Could not create collector", zap.Error(err))
	}

	return &SiteOnej{
		redis:       sys.RedisClient,
		mongoClient: sys.MongoClient,
		logger:      zap.L(),
		colly:       collyClient,
		siteCfg:     cfg,
		client:      resty.New(),
	}
}

const maxActressNumberLimit = 4
const imgSrcKey = "imgSrc"
const attachmentUriKey = "attachmentUri"
const directory = "directory"

// HandleCatalogPage 解析每一页
func (s *SiteOnej) CrawlCatalogPage(ctx context.Context, catalogPageMsg *model.CatalogPageTask) ([]model.NovelTask, error) {
	collyCtx := colly.NewContext()

	url := catalogPageMsg.Url
	base64Url := base64.StdEncoding.EncodeToString([]byte(url))
	if result, err := s.redis.Client.Exists(ctx, base64Url).Result(); err != nil {
		return nil, err
	} else if result > 0 {
		s.logger.Info("url has been handled, just ignores", zap.String("url", url))
		return []model.NovelTask{}, nil
	}

	var novelMsgs []model.NovelTask

	//遍历每一个面板进行解析
	s.colly.OnHTML(".columns", func(element *colly.HTMLElement) {
		name := element.ChildText(".title.is-4.is-spaced>a")

		//只关心所允许的人员总数
		actressNum := 0
		element.ForEach(".panel-block", func(i int, element *colly.HTMLElement) {
			actressNum = i
		})
		if actressNum >= maxActressNumberLimit {
			s.logger.Info("actress number is large, just ignores", zap.Int("actressNum", actressNum),
				zap.String("name", name), zap.String("url", url))
		}

		imgSrc := element.ChildAttr(".column>.image", "src")

		//download button
		attachmentUri := element.ChildAttr(".button.is-primary.is-fullwidth", "href")
		if !strings.HasPrefix(attachmentUri, "http") {
			attachmentUri = utils.BuildUrl(url, attachmentUri)
		}
		novelMsgs = append(novelMsgs, model.NovelTask{
			Name:      name,
			CatalogId: catalogPageMsg.CatalogId,
			Url:       imgSrc, //使用图片地址作为novel的首页地址
			SiteName:  catalogPageMsg.SiteName,
			Attributes: map[string]interface{}{
				imgSrcKey:        imgSrc,
				attachmentUriKey: attachmentUri,
			},
		})
	})

	if err := s.colly.Request("GET", url, nil, collyCtx, nil); err != nil {
		ip := collyCtx.GetAny("inValidPage")
		println("ip=", ip)
		retries := collyCtx.GetAny("retries")
		println("retries=", retries)
		s.logger.Error("visit error", zap.String("url", url), zap.Error(err))
		return nil, err
	}

	return novelMsgs, nil
}

// HandleNovelPage 解析具体的Novel
func (s *SiteOnej) CrawlNovelPage(ctx context.Context, novelPageMsg *model.NovelTask, skipSaveIfPresent bool) ([]model.ChapterTask, error) {
	s.logger.Info("Got novel message", zap.String("name", novelPageMsg.Name))

	if picDir, ok := s.siteCfg.Attributes[directory]; ok {
		//获取catalog name
		catalogName, err := utils.GetAndSet(ctx, novelPageMsg.CatalogId.String(), func() (*string, error) {
			catlogCol := base.GetSystem().GetCollection(base.CollectionCatalog)
			var catalogMsg model.CatalogTask
			if err := catlogCol.FindOne(ctx, bson.M{base.ColumId: novelPageMsg.CatalogId}).Decode(&catalogMsg); err != nil {
				return nil, err
			} else {
				return &catalogMsg.Name, nil
			}
		})
		if err != nil {
			zap.L().Error("catalog not found", zap.String("catalogId", novelPageMsg.CatalogId.String()), zap.Error(err))
			return nil, err
		}
		if catalogName == nil {
			zap.L().Error("catalog not found", zap.String("catalog", "[nil]"))
			return nil, err
		}

		//以catalog name为根目录
		destDir := picDir + "/" + *catalogName

		//下载图片
		if imgUrl, ok := novelPageMsg.Attributes[imgSrcKey]; ok {
			imgUrlString := imgUrl.(string)
			localFile := strings.TrimRight(destDir, "/") + "/" + strings.ToLower(novelPageMsg.Name) + ".jpg"
			restyClient, err := client.GetRestyClient(imgUrlString, true)
			if err != nil {
				return nil, err
			}

			if _, err := restyClient.R().SetOutput(localFile).Get(imgUrlString); err != nil {
				s.logger.Error("download image error", zap.String("url", imgUrlString), zap.Error(err))
				return nil, err
			} else {
				s.logger.Info("image downloaded", zap.String("url", imgUrlString), zap.String("localFile", localFile))
			}
		}

		//下载附件
		if attachmentUrl, ok := novelPageMsg.Attributes[attachmentUriKey]; ok {
			attachUrlString := attachmentUrl.(string)
			restyAttClient, err := client.GetRestyClient(attachUrlString, true)
			if err != nil {
				return nil, err
			}

			lastSlashIndex := strings.LastIndex(attachUrlString, "/")
			attFile := strings.TrimRight(destDir, "/") + "/" + attachUrlString[lastSlashIndex+1:]
			attFile = strings.ReplaceAll(attFile, "onejav.com_", "")
			if _, err := restyAttClient.R().SetOutput(attFile).Get(attachUrlString); err != nil {
				s.logger.Error("download attachment error", zap.String("url", attachUrlString), zap.Error(err))
				return nil, err
			} else {
				s.logger.Info("attachment downloaded", zap.String("url", attachUrlString), zap.String("localFile", attFile))
			}
		}
	}

	return []model.ChapterTask{}, nil
}

func (s *SiteOnej) CrawlHomePage(ctx context.Context, url string) error {
	//TODO implement me
	panic("implement me")
}
func (s *SiteOnej) CrawlChapterPage(ctx context.Context, chapterMsg *model.ChapterTask, skipSaveIfPresent bool) error {
	panic("implement me")
}
