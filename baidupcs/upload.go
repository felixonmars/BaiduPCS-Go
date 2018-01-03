package baidupcs

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/iikira/BaiduPCS-Go/requester"
	"net/http"
	"net/http/cookiejar"
)

// RapidUpload 秒传文件
func (p PCSApi) RapidUpload(targetPath, md5, smd5, crc32 string, length int64) (err error) {
	p.addItem("file", "rapidupload", map[string]string{
		"path":           targetPath,         // 上传文件的全路径名
		"content-length": fmt.Sprint(length), // 待秒传的文件长度
		"content-md5":    md5,                // 待秒传的文件的MD5
		"slice-md5":      smd5,               // 待秒传的文件的MD5
		"content-crc32":  crc32,              // 待秒传文件CRC32
		"ondup":          "overwrite",        // overwrite: 表示覆盖同名文件; newcopy: 表示生成文件副本并进行重命名，命名规则为“文件名_日期.后缀”
	})

	h := requester.NewHTTPClient()
	body, err := h.Fetch("POST", p.url.String(), nil, map[string]string{
		"Cookie": "BDUSS=" + p.bduss,
	})
	if err != nil {
		return
	}

	json, err := simplejson.NewJson(body)
	if err != nil {
		return
	}

	code, err := checkErr(json)

	switch code {
	case 31079:
		// file md5 not found, you should use upload api to upload the whole file.
	}

	if err != nil {
		return fmt.Errorf("秒传文件 遇到错误, 错误代码: %d, 消息: %s", code, err)
	}

	return nil
}

func (p PCSApi) Upload(targetPath string, uploadFunc func(uploadURL string, jar *cookiejar.Jar) error) (err error) {
	p.addItem("file", "upload", map[string]string{
		"path":  targetPath,
		"ondup": "overwrite",
	})

	jar, _ := cookiejar.New(nil)
	jar.SetCookies(&p.url, []*http.Cookie{
		&http.Cookie{
			Name:  "BDUSS",
			Value: p.bduss,
		},
	})

	return uploadFunc(p.url.String(), jar)
}
