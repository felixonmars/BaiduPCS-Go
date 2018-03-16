// Package downloader 多线程下载器
package downloader

import (
	"fmt"
	"io"
	"time"
)

// Downloader 下载
type Downloader struct {
	OnExecute func()
	OnFinish  func()
	OnPause   func()
	OnResume  func()

	URL    string
	Config Config

	sinceTime time.Time
	status    Status
	checked   bool
}

// NewDownloader 创建新的文件下载
func NewDownloader(durl string, cfg Config) (der *Downloader, err error) {
	der = &Downloader{
		URL:    durl,
		Config: cfg,
	}

	err = der.Check()
	if err != nil {
		return nil, err
	}

	return der, nil
}

// Execute 开始执行下载
func (der *Downloader) Execute() (err error) {
	if !der.checked {
		err = der.Check()
		if err != nil {
			return
		}
	}

	if err = der.loadBreakPoint(); err != nil {
		if !der.status.blockUnsupport {
			// 控制线程
			// 如果文件不大, 或者线程数设置过高, 则调低线程数
			if int64(der.Config.Parallel) > der.status.StatusStat.TotalSize/int64(MinParallelSize) {
				der.Config.Parallel = int(der.status.StatusStat.TotalSize/int64(MinParallelSize)) + 1
			}

			blockSize := der.status.StatusStat.TotalSize / int64(der.Config.Parallel)

			// 如果 cache size 过高, 则调低
			if int64(der.Config.CacheSize) > blockSize {
				der.Config.CacheSize = int(blockSize)
			}

			var begin int64
			// 数据平均分配给各个线程
			der.status.BlockList = make(BlockList, der.Config.Parallel)
			for i := 0; i < der.Config.Parallel; i++ {
				var end = (int64(i) + 1) * blockSize
				der.status.BlockList[i] = &Block{
					Begin: begin,
					End:   end,
				}
				begin = end + 1
			}

			// 将余出数据分配给最后一个线程
			der.status.BlockList[der.Config.Parallel-1].End += der.status.StatusStat.TotalSize - der.status.BlockList[der.Config.Parallel-1].End
			der.status.BlockList[der.Config.Parallel-1].IsFinal = true
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
				// 分配缓存空间
				der.status.BlockList[id].buf = make([]byte, der.Config.CacheSize)
				go der.addExecBlock(id)
			}

			// 开启监控
			<-der.blockMonitor()
		}

		// 下载结束
		der.status.done = true
		der.status.file.Close()
		trigger(der.OnFinish)
	}()

	return err
}

// Pause 暂停下载, 不支持单线程暂停下载
func (der *Downloader) Pause() {
	defer trigger(der.OnPause)
	if der.status.paused { // 已经暂停, 退出
		return
	}

	der.status.paused = true
	for _, block := range der.status.BlockList {
		if block != nil && block.resp != nil && !block.resp.Close {
			block.resp.Body.Close()
		}
	}
}

// Resume 恢复下载, 不支持单线程
func (der *Downloader) Resume() {
	defer trigger(der.OnResume)
	if !der.status.paused { // 未被暂停, 退出
		return
	}

	der.status.paused = false
	for id := range der.status.BlockList {
		// 分配缓存空间
		if der.status.BlockList[id].buf == nil {
			der.status.BlockList[id].buf = make([]byte, der.Config.CacheSize)
		}
		go der.addExecBlock(id)
	}
}

func (der *Downloader) singleDownload() error {
	resp, err := der.Config.Client.Req("GET", der.URL, nil, nil)
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
		buf = make([]byte, der.Config.CacheSize)
		n   int
	)

	for {
		n, err = io.ReadFull(resp.Body, buf)
		n64 := int64(n)

		der.status.StatusStat.speedsStat.AddReaded(n64)
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		n, err = der.status.file.Write(buf[:n])
		if err != nil {
			return err
		}

		der.status.StatusStat.Downloaded += n64
	}

	return nil
}
