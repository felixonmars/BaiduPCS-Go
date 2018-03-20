package downloader

import (
	"fmt"
	"mime"
	"net/url"
	"os"
	"path/filepath"
)

// Check 检查配置, 环境, 准备下载条件
func (der *Downloader) Check() (err error) {
	der.Config.Fix()

	// 如果文件存在, 取消下载
	// 测试下载时, 则不检查
	if !der.Config.Testing {
		if der.Config.SavePath != "" {
			err = checkFileExist(der.Config.SavePath)
			if err != nil {
				return
			}
		}
	}

	// 获取文件信息
	resp, err := der.Config.Client.Req("HEAD", der.URL, nil, nil)
	if err != nil {
		return
	}

	// 检测网络错误
	switch resp.StatusCode / 100 {
	case 2: // succeed
	case 4, 5: // error
		return fmt.Errorf(resp.Status)
	}

	der.status.StatusStat.TotalSize = resp.ContentLength

	// 判断服务端是否支持断点续传
	if resp.ContentLength <= 0 {
		der.status.blockUnsupport = true
	}

	if !der.Config.Testing && der.Config.SavePath == "" {
		// 解析文件名, 通过 Content-Disposition
		_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
		if err == nil {
			der.Config.SavePath, _ = url.QueryUnescape(params["filename"])
		}

		if err != nil || der.Config.SavePath == "" {
			// 找不到文件名, 凑合吧
			der.Config.SavePath = filepath.Base(der.URL)
		}

		// 如果文件存在, 取消下载
		err = checkFileExist(der.Config.SavePath)
		if err != nil {
			return err
		}
	}

	if !der.Config.Testing {
		// 检测要保存下载内容的目录是否存在
		// 不存在则创建该目录
		if _, err = os.Stat(filepath.Dir(der.Config.SavePath)); err != nil {
			err = os.MkdirAll(filepath.Dir(der.Config.SavePath), 0777)
			if err != nil {
				return
			}
		}

		// 移除旧的断点续传文件
		if _, err = os.Stat(der.Config.SavePath); err != nil {
			if _, err = os.Stat(der.Config.SavePath + DownloadingFileSuffix); err == nil {
				os.Remove(der.Config.SavePath + DownloadingFileSuffix)
			}
		}

		// 检测要下载的文件是否存在
		// 如果存在, 则打开文件
		// 不存在则创建文件
		file, err := os.OpenFile(der.Config.SavePath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}

		der.status.file = file
	} else {
		der.status.file, _ = os.Open(os.DevNull)
	}

	der.checked = true
	return resp.Body.Close()
}
