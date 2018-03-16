package downloader

import (
	"errors"
	"github.com/json-iterator/go"
	"io/ioutil"
	"time"
)

// StatusStat 统计状态数据
type StatusStat struct {
	TotalSize   int64 `json:"total_size"` // 总大小
	Downloaded  int64 `json:"downloaded"` // 已下载的数据量
	Speeds      int64 // 下载速度
	maxSpeeds   int64 // 最大下载速度
	speedsStat  SpeedsStat
	TimeElapsed time.Duration `json:"-"` // 下载的时间
}

// Status 下载状态
type Status struct {
	StatusStat

	file           Writer
	BlockList      BlockList `json:"block_list"` // 下载区块列表
	blockUnsupport bool      // 服务端是否支持断点续传
	paused         bool      // 是否暂停
	done           bool      // 是否已经结束
}

// GetStatusChan 返回 Status 对象的 channel
func (der *Downloader) GetStatusChan() <-chan StatusStat {
	c := make(chan StatusStat)

	go func() {
		for {
			// 针对单线程下载的速度统计
			if der.status.blockUnsupport {
				der.status.speedsStat.Start()
			}

			time.Sleep(1 * time.Second)
			der.status.TimeElapsed = time.Since(der.sinceTime) / 1e6 * 1e6

			if der.status.blockUnsupport {
				der.status.Speeds = der.status.speedsStat.EndAndGetSpeedsPerSecond()
			}

			// 下载结束, 关闭 chan
			if der.status.done {
				close(c)
				return
			}

			c <- der.status.StatusStat
		}
	}()

	return c
}

// recordBreakPoint 保存下载断点到文件, 用于断点续传
func (der *Downloader) recordBreakPoint() error {
	if der.Config.Testing {
		return errors.New("Testing not support record break points")
	}

	if der.status.blockUnsupport {
		return errors.New("服务端不支持断点续传, 不记录断点信息")
	}

	byt, err := jsoniter.Marshal(der.status)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(der.Config.SavePath+DownloadingFileSuffix, byt, 0644)
}

// loadBreakPoint 尝试从文件载入下载断点
func (der *Downloader) loadBreakPoint() error {
	if der.Config.Testing {
		return errors.New("Testing not support load break points")
	}

	if der.status.blockUnsupport {
		return errors.New("服务端不支持断点续传, 不载入断点信息")
	}

	byt, err := ioutil.ReadFile(der.Config.SavePath + DownloadingFileSuffix)
	if err != nil {
		return err
	}

	err = jsoniter.Unmarshal(byt, &der.status)
	if err != nil {
		return err
	}

	return nil
}
