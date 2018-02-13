/*
 Copyright 2015 Bluek404

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.

 补充: 此 implement 基于 https://github.com/Bluek404/downloader
 针对百度网盘下载, 做出一些修改

 增加功能: 线程控制等
 删去功能: 暂停下载, 恢复下载
*/

package downloader

import (
	"github.com/iikira/BaiduPCS-Go/requester"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var (
	// FileNameRE 正则表达式: 匹配文件名
	FileNameRE = regexp.MustCompile("filename=\"(.*?)\"")
)

// Downloader 下载详情
type Downloader struct {
	URL       string    // 下载地址
	BlockList blockList // 用于记录未下载的文件块起始位置
	Options   *Options

	file *os.File // 要写入的文件
	size int64    // 文件大小

	onStart  func()
	onFinish func()
	onError  func(int, error)

	sinceTime time.Time
	status    DownloadStatus // 下载状态
}

// NewDownloader 创建新的文件下载
func NewDownloader(url, savePath string, o *Options) (der *Downloader, err error) {
	if o == nil {
		o = NewOptions()
	}

	if !o.Testing {
		// 如果文件存在, 取消下载
		// 测试下载时, 则不检查
		if savePath != "" {
			err = checkFileExist(savePath)
			if err != nil {
				return nil, err
			}
		}
	}

	if o.Client == nil {
		o.Client = requester.NewHTTPClient()
	}

	// 获取文件信息
	resp, err := o.Client.Req("HEAD", url, nil, nil)
	if err != nil {
		return nil, err
	}

	der = &Downloader{
		URL:     url,
		Options: o,
		size:    resp.ContentLength,
	}

	if !o.Testing && savePath == "" {
		finds := FileNameRE.FindStringSubmatch(
			resp.Header.Get("Content-Disposition"),
		)
		if len(finds) >= 2 {
			savePath = finds[1]
		} else {
			// 找不到文件名, 凑合吧
			savePath = filepath.Base(url)
		}

		// 如果文件存在, 取消下载
		err = checkFileExist(savePath)
		if err != nil {
			return nil, err
		}
	}

	if !o.Testing {
		// 检测要保存下载内容的目录是否存在
		// 不存在则创建该目录
		if _, err = os.Stat(filepath.Dir(savePath)); err != nil {
			err = os.MkdirAll(filepath.Dir(savePath), 0777)
			if err != nil {
				return nil, err
			}
		}

		// 移除旧的断点续传文件
		if _, err = os.Stat(savePath); err != nil {
			if _, err = os.Stat(savePath + DownloadingFileSuffix); err == nil {
				os.Remove(savePath + DownloadingFileSuffix)
			}
		}

		// 检测要下载的文件是否存在
		// 如果存在, 则打开文件
		// 不存在则创建文件
		file, err := os.OpenFile(savePath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return nil, err
		}

		der.file = file
	}

	resp.Body.Close()

	return der, nil
}

// StartDownload 开始下载
func (der *Downloader) StartDownload() {
	if der.Options == nil {
		der.Options = NewOptions()
	}

	// 控制线程
	// 如果文件不大, 或者线程数设置过高, 则调低线程数
	if int64(der.Options.Parallel) > der.size/int64(102400) {
		der.Options.Parallel = int(der.size/int64(102400)) + 1
	}

	blockSize := der.size / int64(der.Options.Parallel)

	// 如果 cache size 过高, 则调低
	if int64(der.Options.CacheSize) > blockSize {
		der.Options.CacheSize = int(blockSize)
	}

	if err := der.loadBreakPoint(); err != nil {
		if der.size <= 0 { // 获取不到文件的大小, 关闭多线程下载 (暂时)
			der.BlockList = append(der.BlockList, &Block{
				Begin: 0,
				End:   -1,
			})
		} else {
			var begin int64
			// 数据平均分配给各个线程
			for i := 0; i < der.Options.Parallel; i++ {
				var end = (int64(i) + 1) * blockSize
				der.BlockList = append(der.BlockList, &Block{
					Begin: begin,
					End:   end,
				})
				begin = end + 1
			}
			// 将余出数据分配给最后一个线程
			der.BlockList[der.Options.Parallel-1].End += der.size - der.BlockList[der.Options.Parallel-1].End
			der.BlockList[der.Options.Parallel-1].Final = true
		}
	}

	go func() {
		touch(der.onStart)

		// 开始下载
		der.sinceTime = time.Now()
		der.status.Total = der.size

		err := der.download()
		if err != nil {
			der.touchOnError(0, err)
			return
		}
	}()
}

func (der *Downloader) download() error {
	for i := range der.BlockList {
		go func(id int) {
			// 分配缓存空间
			der.BlockList[id].buf = make([]byte, der.Options.CacheSize)
			der.downloadBlockFn(id)
		}(i)
	}

	// 开启监控
	<-der.blockMonitor()

	der.status.done = true
	touch(der.onFinish)

	if !der.Options.Testing {
		der.file.Close()
	}

	return nil
}

// OnStart 任务开始时触发的事件
func (der *Downloader) OnStart(fn func()) {
	der.onStart = fn
}

// OnFinish 任务完成时触发的事件
func (der *Downloader) OnFinish(fn func()) {
	der.onFinish = fn
}

// OnError 任务出错时触发的事件
//
// errCode为错误码，errStr为错误描述
func (der *Downloader) OnError(fn func(int, error)) {
	der.onError = fn
}

// 用于触发事件
func touch(fn func()) {
	if fn != nil {
		go fn()
	}
}

// 触发Error事件
func (der *Downloader) touchOnError(errCode int, err error) {
	if der.onError != nil {
		go der.onError(errCode, err)
	}
}
