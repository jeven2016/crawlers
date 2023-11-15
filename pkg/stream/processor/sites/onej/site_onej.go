package onej

import (
	"context"
	"crawlers/pkg/base"
	"crawlers/pkg/model/entity"
	"encoding/base64"
	"errors"
	"github.com/gocolly/colly/v2"
	"github.com/jeven2016/mylibs/client"
	"github.com/jeven2016/mylibs/system"
	"github.com/jeven2016/mylibs/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"strings"
)

type SiteOnej struct {
	colly *colly.Collector
}

func NewSiteOnej() *SiteOnej {
	collyClient, err := client.NewCollector("", 3)
	if err != nil {
		zap.L().Warn("Could not create collector", zap.Error(err))
	}

	return &SiteOnej{
		colly: collyClient,
	}
}

const maxActressNumberLimit = 4
const imgSrcKey = "imgSrc"
const attachmentUriKey = "attachmentUri"
const directory = "directory"

// CrawlCatalogPage 解析每一页
func (s *SiteOnej) CrawlCatalogPage(ctx context.Context, catalogPageMsg *entity.CatalogPageTask) ([]entity.NovelTask, error) {
	sys := system.GetSystem()
	collyCtx := colly.NewContext()

	url := catalogPageMsg.Url
	base64Url := base64.StdEncoding.EncodeToString([]byte(url))
	if result, err := sys.RedisClient.Client.Exists(ctx, base64Url).Result(); err != nil {
		return nil, err
	} else if result > 0 {
		zap.L().Info("url has been handled, just ignores", zap.String("url", url))
		return []entity.NovelTask{}, nil
	}

	var novelMsgs []entity.NovelTask

	//遍历每一个面板进行解析
	s.colly.OnHTML(".columns", func(element *colly.HTMLElement) {
		name := element.ChildText(".title.is-4.is-spaced>a")

		//只关心所允许的人员总数
		actressNum := 0
		element.ForEach(".panel-block", func(i int, element *colly.HTMLElement) {
			actressNum = i
		})
		if actressNum >= maxActressNumberLimit {
			zap.L().Info("actress number is large, just ignores", zap.Int("actressNum", actressNum),
				zap.String("name", name), zap.String("url", url))
		}

		imgSrc := element.ChildAttr(".column>.image", "src")

		//download button
		attachmentUri := element.ChildAttr(".button.is-primary.is-fullwidth", "href")
		if !strings.HasPrefix(attachmentUri, "http") {
			attachmentUri = utils.BuildUrl(url, attachmentUri)
		}
		novelMsgs = append(novelMsgs, entity.NovelTask{
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
		zap.L().Error("visit error", zap.String("url", url), zap.Error(err))
		return nil, err
	}

	return novelMsgs, nil
}

// CrawlNovelPage 解析具体的Novel
func (s *SiteOnej) CrawlNovelPage(ctx context.Context, novelPageMsg *entity.NovelTask, skipSaveIfPresent bool) ([]entity.ChapterTask, error) {
	zap.L().Info("Got novel message", zap.String("name", novelPageMsg.Name))
	siteCfg := base.GetSiteConfig(base.SiteOneJ)
	if siteCfg == nil {
		return nil, errors.New("no site config found for site " + base.SiteOneJ)
	}

	if picDir, ok := siteCfg.Attributes[directory]; ok {
		//获取catalog name
		catalogName, err := utils.GetAndSet(ctx, novelPageMsg.CatalogId.String(), func() (*string, error) {
			catlogCol := system.GetSystem().GetCollection(base.CollectionCatalog)
			var catalogMsg entity.CatalogTask
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
				zap.L().Error("download image error", zap.String("url", imgUrlString), zap.Error(err))
				return nil, err
			} else {
				zap.L().Info("image downloaded", zap.String("url", imgUrlString), zap.String("localFile", localFile))
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
				zap.L().Error("download attachment error", zap.String("url", attachUrlString), zap.Error(err))
				return nil, err
			} else {
				zap.L().Info("attachment downloaded", zap.String("url", attachUrlString), zap.String("localFile", attFile))
			}
		}
	}

	return []entity.ChapterTask{}, nil
}

func (s *SiteOnej) CrawlHomePage(ctx context.Context, url string) error {
	//TODO implement me
	panic("implement me")
}
func (s *SiteOnej) CrawlChapterPage(ctx context.Context, chapterMsg *entity.ChapterTask, skipSaveIfPresent bool) error {
	panic("implement me")
}
