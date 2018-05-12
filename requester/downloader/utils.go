package downloader

import (
	"github.com/iikira/BaiduPCS-Go/requester"
	"mime"
	"net/url"
	"path"
)

// GetFileName 获取文件名
func GetFileName(uri string, client *requester.HTTPClient) (filename string, err error) {
	if client == nil {
		client = requester.NewHTTPClient()
	}

	resp, err := client.Req("HEAD", uri, nil, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}

	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	if err != nil {
		return "", err
	}

	filename, err = url.QueryUnescape(params["filename"])
	if err != nil {
		return
	}

	if filename == "" {
		filename = path.Base(uri)
	}

	return
}

//trigger 用于触发事件
func trigger(f func()) {
	if f == nil {
		return
	}
	go f()
}

func fixCacheSize(size *int) {
	if *size < 1024 {
		*size = 1024
	}
}
