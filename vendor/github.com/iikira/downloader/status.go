package downloader

import (
	"errors"
	"github.com/json-iterator/go"
	"io/ioutil"
	"time"
)

// Status 下载状态
type Status struct {
	done bool // 是否已经结束, alignment, for 32-bit device

	TotalSize      int64 `json:"total_size"` // 总大小
	blockUnsupport bool  // 服务端是否支持断点续传, alignment

	Downloaded int64     `json:"downloaded"` // 已下载的数据量
	BlockList  BlockList `json:"block_list"` // 下载区块列表

	Speeds      int64         `json:"-"` // 下载速度, 每秒
	MaxSpeeds   int64         `json:"-"` // 最大下载速度
	TimeElapsed time.Duration `json:"-"` // 下载的时间
}

// GetStatusChan 返回 Status 对象的 channel
func (der *Downloader) GetStatusChan() <-chan *Status {
	c := make(chan *Status)

	go func() {
		var old = der.status.Downloaded
		for {
			time.Sleep(1 * time.Second) // 每秒统计

			der.status.Speeds = der.status.Downloaded - old
			old = der.status.Downloaded

			if der.status.Speeds > der.status.MaxSpeeds {
				der.status.MaxSpeeds = der.status.Speeds
			}

			der.status.TimeElapsed = time.Since(der.sinceTime) / 1e6 * 1e6

			// 下载结束, 关闭 chan
			if der.status.done {
				close(c)
				return
			}

			c <- &der.status
		}
	}()

	return c
}

// recordBreakPoint 保存下载断点到文件, 用于断点续传
func (der *Downloader) recordBreakPoint() error {
	if der.config.Testing {
		return errors.New("Testing not support record break points")
	}

	if der.status.blockUnsupport {
		return errors.New("服务端不支持断点续传, 不记录断点信息")
	}

	byt, err := jsoniter.Marshal(der.status)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(der.config.SavePath+DownloadingFileSuffix, byt, 0644)
}

// loadBreakPoint 尝试从文件载入下载断点
func (der *Downloader) loadBreakPoint() error {
	if der.config.Testing {
		return errors.New("Testing not support load break points")
	}

	if der.status.blockUnsupport {
		return errors.New("服务端不支持断点续传, 不载入断点信息")
	}

	byt, err := ioutil.ReadFile(der.config.SavePath + DownloadingFileSuffix)
	if err != nil {
		return err
	}

	s := &Status{}
	err = jsoniter.Unmarshal(byt, s)
	if err != nil {
		return err
	}

	der.status = *s
	return nil
}
