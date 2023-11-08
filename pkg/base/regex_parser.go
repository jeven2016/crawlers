package base

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var regexMap = map[string]*regexp.Regexp{}
var invalidPageParameterErr = errors.New("invalid page parameter")

// GenPageUrls 解析page=1, page=1,2-8两类分页参数，并返回拼装好的URL
func GenPageUrls(pageRegex, url, pagePrefix, pageSuffix string) ([]string, error) {
	var err error
	var regex = regexMap[pageRegex]

	if regex == nil {
		if regex, err = regexp.Compile(pageRegex); err != nil {
			return nil, nil
		}
		regexMap[pageRegex] = regex
	}

	submatch := regex.FindStringSubmatch(url)
	if len(submatch) == 2 {
		pageString := submatch[1]

		var pageNoArray []uint32
		itemArray := strings.Split(pageString, ",")
		if len(itemArray) == 0 {
			return nil, invalidPageParameterErr
		}
		for _, item := range itemArray {
			if item == "" {
				return nil, invalidPageParameterErr
			}
			pages := strings.Split(item, "-")
			pagesLen := len(pages)
			if pagesLen == 1 {
				pageNo, err := strconv.ParseUint(item, 10, 32)
				if err != nil {
					return nil, invalidPageParameterErr
				}
				pageNoArray = append(pageNoArray, uint32(pageNo))
			} else if pagesLen == 2 {
				pageStart, err := strconv.ParseUint(pages[0], 10, 32)
				if err != nil {
					return nil, invalidPageParameterErr
				}
				pageEnd, err := strconv.ParseUint(pages[1], 10, 32)
				if err != nil {
					return nil, invalidPageParameterErr
				}
				for i := pageStart; i <= pageEnd; i++ {
					pageNoArray = append(pageNoArray, uint32(i))
				}
			} else {
				return nil, invalidPageParameterErr
			}

		}

		var pageUrls []string
		for _, pageNo := range pageNoArray {
			pageUrl := regex.ReplaceAllString(url, pagePrefix+strconv.Itoa(int(pageNo))+pageSuffix)
			pageUrls = append(pageUrls, pageUrl)
		}
		return pageUrls, nil
	}
	return []string{url}, nil
}
