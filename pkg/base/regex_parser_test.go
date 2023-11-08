package base

import (
	"testing"
)

var urlRegex = "(?:page=)([^\\&]+)"
var nsfRegex = "([^/]+)\\.html"

func TestComplexUrlRegex(t *testing.T) {
	var url = "https://www.baidu.com?page=1,2-8&format=json"
	urls, err := GenPageUrls(urlRegex, url, "page=", "")
	if err != nil {
		t.Error(err)
	}
	if len(urls) != 8 {
		t.Error("urls length error")
	}
}

func TestSimpleUrlRegex(t *testing.T) {
	var url = "https://www.baidu.com?page=1&format=json"
	urls, err := GenPageUrls(urlRegex, url, "page=", "")
	if err != nil {
		t.Error(err)
	}
	if len(urls) != 1 {
		t.Error("urls length error")
	}
}

func TestInvalidUrlRegex(t *testing.T) {
	var url = "https://www.baidu.com?page=1,,8,4&format=json"
	urls, err := GenPageUrls(urlRegex, url, "page=", "")
	if err == nil {
		t.Error(err)
	}
	if len(urls) != 0 {
		t.Error("urls should be empty")
	}
}

func TestIndividualPageNumber(t *testing.T) {
	var url = "https://www.baidu.com?page=1,2,3,4&format=json"
	urls, err := GenPageUrls(urlRegex, url, "page=", "")
	if err != nil {
		t.Error(err)
	}
	if len(urls) != 4 {
		t.Error("urls length error")
	}
}

func TestNsfPage(t *testing.T) {
	var url = "https://abc.com/category/18/3-5.html"
	urls, err := GenPageUrls(nsfRegex, url, "", ".html")
	if err != nil {
		t.Error(err)
	}
	if len(urls) != 3 {
		t.Error("urls length error")
	}
}
