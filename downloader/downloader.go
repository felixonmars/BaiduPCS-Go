// Package downloader 多线程下载器
package downloader

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/iikira/BaiduPCS-Go/requester/downloader/cachepool"
	"io"
	"sync"
	"time"
)

// Downloader 下载
type Downloader struct {
	// 涉及原子操作, 兼容32位设备, 注意内存地址对齐

	OnExecute     func()
	OnFinish      func()
	OnPause       func()
	OnResume      func()
	OnCancel      func()                    // 手动取消
	OnCancelError func(code int, err error) // 中途遇到下载错误而取消的

	status    Status
	sinceTime time.Time
	writeMu   sync.Mutex
	monitorMu sync.Mutex

	URL    string
	Config Config

	checked bool
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
func (der *Downloader) Execute() (done <-chan struct{}, err error) {
	d := make(chan struct{}, 0)

	if !der.checked {
		err = der.Check()
		if err != nil {
			d <- struct{}{}
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

			pcsverbose.Verbosef("CREATED: parallel: %d, cache size: %d\n", der.Config.Parallel, der.Config.CacheSize)

			// 数据平均分配给各个线程
			var begin, end int64
			der.status.BlockList = make(BlockList, der.Config.Parallel)
			for k := range der.status.BlockList {
				end = int64(k+1) * blockSize
				der.status.BlockList[k] = &Block{
					Begin: begin,
					End:   end,
				}
				begin = end + 1
			}

			// 将余出数据分配给最后一个线程
			der.status.BlockList[der.Config.Parallel-1].End = der.status.StatusStat.TotalSize
			der.status.BlockList[der.Config.Parallel-1].IsFinal = true
		}
	}

	pcsverbose.Verbosef("DEBUG: download start\n")

	go func() {
		defer func() {
			d <- struct{}{}
		}()

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
				go der.addExecBlock(id)
			}

			// 开启监控
			<-der.blockMonitor()
		}

		// 下载结束
		der.status.done = true
		der.status.file.Close()
		trigger(der.OnFinish)
		pcsverbose.Verbosef("DEBUG: download finish\n")
	}()

	return d, nil
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
		go der.addExecBlock(id)
	}
}

// Cancel 取消下载
func (der *Downloader) Cancel() {
	der.cancel()
	trigger(der.OnCancel)
}

func (der *Downloader) cancel() {
	// 关闭所有连接
	for _, block := range der.status.BlockList {
		if block != nil && block.resp != nil && !block.resp.Close {
			block.resp.Body.Close()
		}
	}

	if resp := der.status.singleResp; resp != nil && !resp.Close {
		resp.Body.Close()
	}
}

func (der *Downloader) singleDownload() (err error) {
	der.status.singleResp, err = der.Config.Client.Req("GET", der.URL, nil, nil)
	if der.status.singleResp != nil {
		defer der.status.singleResp.Body.Close()
	}
	if err != nil {
		return err
	}

	switch der.status.singleResp.StatusCode / 100 {
	case 4, 5:
		return fmt.Errorf(der.status.singleResp.Status)
	}

	var (
		buf = cachepool.SetIfNotExist(0, der.Config.CacheSize)
		n   int
	)

	for {
		n, err = io.ReadFull(der.status.singleResp.Body, buf)
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
