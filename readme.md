
### install
```shell
go get github.com/gocolly/colly/v2 latest
msgp 序列化
```
https://github.com/prometheus/client_golang/blob/main/prometheus/examples_test.go
https://github.com/lao-siji/lao-siji

``
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build main.go
``

### Build by taskfile
Using taskfile to build

```shell
# install 
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d

chmod -R 777 ./bin/task
mv ./bin/task /usr/local/bin/

task -version

```

### swagger
```text
参数说明 ：@Param 参数名 位置（query / path / body / header） 类型 是否必需 注释

// HandleCatalogPage handle catalog page request and parse the novel links for further processing
// @Tags 测试
// @Summary  处理目录页面请求
// @Description 处理目录页面请求,解析出Novel的地址并发送到消息对列中去
// @Param name	query string true "Bearer 31a165baebe6dec616b1f8f3207b4273"
// @Accept  json
// @Product json
// @Param   id     query    int     true        "用户id"
// @Param   name   query    string  false        "用户name"
// @Success 200 {object} string	"{"code": 200, "data": [...]}"
// @Router /getUser/:id [get]
```

libs
* github.com/go-co-op/gocron
* chromedp(浏览器抓取)