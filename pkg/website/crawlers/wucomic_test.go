package crawlers

import (
	"bytes"
	"context"
	"crawlers/pkg/base"
	"encoding/base64"
	"github.com/chromedp/chromedp"
	"image/jpeg"
	"os"
	"testing"
)

func TestDownloadPic(t *testing.T) {
	url := "https://www.wuwucomic.xyz/chapter/22906"

	chromeCtx, cleanFunc := base.OpenChrome(context.Background())
	defer cleanFunc()

	var data string
	err := chromedp.Run(chromeCtx,
		chromedp.Navigate(url),
		chromedp.ScrollIntoView(`body`, chromedp.ByQuery),
		//chromedp.WaitNotPresent("//p[contains(text(),'内容未加载完成')]", chromedp.BySearch),
		chromedp.Evaluate(`
				$('#canvas_689512')[0].toDataURL("image/jpeg")
			`,
			&data,
		),
	)

	// 滚动到页面的最底部
	err = chromedp.Run(chromeCtx, chromedp.ScrollIntoView(`body`, chromedp.ByQuery))
	if err != nil {
		panic(err)
	}

	err = chromedp.Run(chromeCtx,
		chromedp.Navigate(url),
		chromedp.Evaluate(`
				$('#canvas_689512')[0].toDataURL("image/jpeg")
			`,
			&data,
		),
	)
	if err != nil {
		panic(err)
	}

	// 解码 base64 字符串
	byteData, err := base64.StdEncoding.DecodeString(data[:22])
	if err != nil {
		panic(err)
	}

	// 创建 io.Reader 对象
	reader := bytes.NewReader([]byte(byteData))

	// 创建 Image 对象
	img, err := jpeg.Decode(reader)
	if err != nil {
		panic(err)
	}

	file, err := os.Create("/home/cloud/Desktop/wang.jpg")
	if err != nil {
		panic(err)
	}

	// 保存图片
	err = jpeg.Encode(file, img, nil)
	if err != nil {
		panic(err)
	}
}
