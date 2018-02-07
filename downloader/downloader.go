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
	"fmt"
	"github.com/iikira/BaiduPCS-Go/requester"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var parallel int

// FileDl 下载详情
type FileDl struct {
	URL  string   // 下载地址
	Size int64    // 文件大小
	File *os.File // 要写入的文件

	BlockList blockList // 用于记录未下载的文件块起始位置

	client *requester.HTTPClient // http client

	onStart  func()
	onFinish func()
	onError  func(int, error)

	status status // 下载状态
}

// NewFileDl 创建新的文件下载
//
// 如果 size <= 0 则自动获取文件大小
func NewFileDl(h *requester.HTTPClient, url, savePath string) (*FileDl, error) {
	// 获取文件信息
	request, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.Do(request)
	if err != nil {
		return nil, err
	}

	if savePath == "" {
		finds := FileNameRE.FindStringSubmatch(
			resp.Header.Get("Content-Disposition"),
		)
		if len(finds) >= 2 {
			savePath = finds[1]
		} else {
			// 找不到文件名, 凑合吧
			savePath = filepath.Base(url)
		}
	}

	// 如果文件存在, 取消下载
	if _, err = os.Stat(savePath); err == nil {
		if _, err = os.Stat(savePath + DownloadingFileSuffix); err != nil {
			return nil, fmt.Errorf("文件已存在: %s", savePath)
		}
	}

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

	resp.Body.Close()

	f := &FileDl{
		URL:    url,
		Size:   resp.ContentLength,
		File:   file,
		client: h,
	}

	return f, nil
}

// Start 开始下载
func (f *FileDl) Start() {
	// 控制线程
	parallel = maxParallel

	// 如果文件不大, 或者线程数设置过高, 则调低线程数
	if int64(maxParallel) > f.Size/int64(102400) {
		parallel = int(f.Size/int64(102400)) + 1
	}

	blockSize := f.Size / int64(parallel)

	// 如果 cache size 过高, 则调低
	if int64(cacheSize) > blockSize {
		cacheSize = int(blockSize)
	}

	if err := f.loadBreakPoint(); err != nil {
		if f.Size <= 0 { // 获取不到文件的大小, 关闭多线程下载 (暂时)
			f.BlockList = append(f.BlockList, Block{
				Begin: 0,
				End:   -1,
			})
		} else {
			var begin int64
			// 数据平均分配给各个线程
			for i := 0; i < parallel; i++ {
				var end = (int64(i) + 1) * blockSize
				f.BlockList = append(f.BlockList, Block{
					Begin: begin,
					End:   end,
				})
				begin = end + 1
			}
			// 将余出数据分配给最后一个线程
			f.BlockList[parallel-1].End += f.Size - f.BlockList[parallel-1].End
			f.BlockList[parallel-1].Final = true
		}
	}

	go func() {
		f.touch(f.onStart)
		// 开始下载
		err := f.download()
		if err != nil {
			f.touchOnError(0, err)
			return
		}
	}()
}

func (f *FileDl) download() error {
	f.startGetSpeeds() // 启用速度监测

	for i := range f.BlockList {
		go func(id int) {
			f.downloadBlockFn(id)
		}(i)
	}
	<-f.blockMonitor()

	f.touch(f.onFinish)

	f.File.Close()

	return nil
}

func (f *FileDl) startGetSpeeds() {
	go func() {
		var old = f.status.Downloaded
		for {
			time.Sleep(time.Second * 1)
			f.status.Speeds = f.status.Downloaded - old
			old = f.status.Downloaded

			if f.status.Speeds > f.status.MaxSpeeds {
				f.status.MaxSpeeds = f.status.Speeds
			}
		}
	}()
}

// GetStatus 获取下载统计信息
func (f FileDl) GetStatus() status {
	return f.status
}

// OnStart 任务开始时触发的事件
func (f *FileDl) OnStart(fn func()) {
	f.onStart = fn
}

// OnFinish 任务完成时触发的事件
func (f *FileDl) OnFinish(fn func()) {
	f.onFinish = fn
}

// OnError 任务出错时触发的事件
//
// errCode为错误码，errStr为错误描述
func (f *FileDl) OnError(fn func(int, error)) {
	f.onError = fn
}

// 用于触发事件
func (f FileDl) touch(fn func()) {
	if fn != nil {
		go fn()
	}
}

// 触发Error事件
func (f FileDl) touchOnError(errCode int, err error) {
	if f.onError != nil {
		go f.onError(errCode, err)
	}
}

// status 状态
type status struct {
	Downloaded int64 `json:"downloaded"`
	Speeds     int64
	MaxSpeeds  int64
}
