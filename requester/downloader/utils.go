package downloader

import (
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/iikira/BaiduPCS-Go/requester"
	mathrand "math/rand"
	"mime"
	"net/url"
	"path"
	"time"
)

// RandomNumber 生成指定区间随机数
func RandomNumber(min, max int) int {
	s := mathrand.NewSource(time.Now().UnixNano())
	r := mathrand.New(s)
	if min > max {
		min, max = max, min
	}
	return r.Intn(max-min) + min
}

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
		pcsverbose.Verbosef("DEBUG: GetFileName ParseMediaType error: %s\n", err)
		return path.Base(uri), nil
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
