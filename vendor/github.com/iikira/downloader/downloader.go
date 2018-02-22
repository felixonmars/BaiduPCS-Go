// Package downloader 多线程下载器
package downloader

import (
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// Downloader 下载
type Downloader struct {
	OnExecute func()
	OnFinish  func()

	url    string
	config *Config

	file Writer

	sinceTime time.Time
	status    Status
}

// NewDownloader 创建新的文件下载
func NewDownloader(durl string, cfg *Config) (der *Downloader, err error) {
	if cfg == nil {
		cfg = NewConfig()
	}

	cfg.Fix()

	// 如果文件存在, 取消下载
	// 测试下载时, 则不检查
	if !cfg.Testing {
		if cfg.SavePath != "" {
			err = checkFileExist(cfg.SavePath)
			if err != nil {
				return nil, err
			}
		}
	}

	// 获取文件信息
	resp, err := cfg.Client.Req("HEAD", durl, nil, nil)
	if err != nil {
		return nil, err
	}

	// 检测网络错误
	switch resp.StatusCode / 100 {
	case 2: // succeed
	case 4, 5: // error
		return nil, fmt.Errorf(resp.Status)
	}

	// 设置新的url, 如果网页存在跳转
	if resp.Request != nil {
		if resp.Request.URL != nil {
			durl = resp.Request.URL.String()
		}
	}

	der = &Downloader{
		url:    durl,
		config: cfg,
		status: Status{
			TotalSize: resp.ContentLength,
		},
	}

	// 判断服务端是否支持断点续传
	if resp.ContentLength <= 0 {
		der.status.blockUnsupport = true
	}

	if !cfg.Testing && cfg.SavePath == "" {
		// 解析文件名, 通过 Content-Disposition
		_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
		if err == nil {
			cfg.SavePath, _ = url.QueryUnescape(params["filename"])
		}

		if err != nil || cfg.SavePath == "" {
			// 找不到文件名, 凑合吧
			cfg.SavePath = filepath.Base(durl)
		}

		// 如果文件存在, 取消下载
		err = checkFileExist(cfg.SavePath)
		if err != nil {
			return nil, err
		}
	}

	if !cfg.Testing {
		// 检测要保存下载内容的目录是否存在
		// 不存在则创建该目录
		if _, err = os.Stat(filepath.Dir(cfg.SavePath)); err != nil {
			err = os.MkdirAll(filepath.Dir(cfg.SavePath), 0777)
			if err != nil {
				return nil, err
			}
		}

		// 移除旧的断点续传文件
		if _, err = os.Stat(cfg.SavePath); err != nil {
			if _, err = os.Stat(cfg.SavePath + DownloadingFileSuffix); err == nil {
				os.Remove(cfg.SavePath + DownloadingFileSuffix)
			}
		}

		// 检测要下载的文件是否存在
		// 如果存在, 则打开文件
		// 不存在则创建文件
		file, err := os.OpenFile(cfg.SavePath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return nil, err
		}

		der.file = file
	} else {
		der.file, _ = os.Open(os.DevNull)
	}

	resp.Body.Close()

	return der, nil
}

// Execute 开始执行下载
func (der *Downloader) Execute() (err error) {
	if der.config == nil {
		return fmt.Errorf("请先通过NewDownloader初始化")
	}

	if err = der.loadBreakPoint(); err != nil {
		if !der.status.blockUnsupport {
			// 控制线程
			// 如果文件不大, 或者线程数设置过高, 则调低线程数
			if int64(der.config.Parallel) > der.status.TotalSize/int64(MinParallelSize) {
				der.config.Parallel = int(der.status.TotalSize/int64(MinParallelSize)) + 1
			}

			blockSize := der.status.TotalSize / int64(der.config.Parallel)

			// 如果 cache size 过高, 则调低
			if int64(der.config.CacheSize) > blockSize {
				der.config.CacheSize = int(blockSize)
			}

			var begin int64
			// 数据平均分配给各个线程
			der.status.BlockList = make(BlockList, der.config.Parallel)
			for i := 0; i < der.config.Parallel; i++ {
				var end = (int64(i) + 1) * blockSize
				der.status.BlockList[i] = &Block{
					Begin: begin,
					End:   end,
				}
				begin = end + 1
			}

			// 将余出数据分配给最后一个线程
			der.status.BlockList[der.config.Parallel-1].End += der.status.TotalSize - der.status.BlockList[der.config.Parallel-1].End
			der.status.BlockList[der.config.Parallel-1].IsFinal = true
		}
	}

	go func() {
		trigger(der.OnExecute)

		// 开始下载
		der.sinceTime = time.Now()

		if der.status.blockUnsupport {
			// 不支持断点续传
			serr := der.singleDownload()
			if serr != nil {
				err = serr
			}
		} else {
			for id := range der.status.BlockList {
				go func(id int) {
					// 分配缓存空间
					der.status.BlockList[id].buf = make([]byte, der.config.CacheSize)
					der.addExecBlock(id)
				}(id)
			}

			// 开启监控
			<-der.blockMonitor()
		}

		// 下载结束
		der.status.done = true
		der.file.Close()
		trigger(der.OnFinish)
	}()

	return err
}

func (der *Downloader) singleDownload() error {
	resp, err := der.config.Client.Req("GET", der.url, nil, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	switch resp.StatusCode / 100 {
	case 4, 5:
		return fmt.Errorf(resp.Status)
	}

	var (
		buf = make([]byte, der.config.CacheSize)
		n   int
	)

	for {
		n, err = resp.Body.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		n, err = der.file.Write(buf[:n])
		if err != nil {
			return err
		}

		der.status.Downloaded += int64(n)
	}

	return nil
}
