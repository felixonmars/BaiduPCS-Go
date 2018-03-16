package downloader

import (
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	mu sync.Mutex
)

// blockMonitor 延迟监控各线程状态,
// 重设长时间无响应, 和下载速度为 0 的线程
// 动态分配新线程
func (der *Downloader) blockMonitor() <-chan struct{} {
	c := make(chan struct{})
	go func() {
		for {
			// 下载暂停, 不开启监控
			if der.paused {
				time.Sleep(2 * time.Second)
				continue
			}

			// 下载完毕, 线程全部完成下载任务, 发送结束信号
			if der.status.BlockList.isAllDone() {
				c <- struct{}{}

				if !der.config.Testing {
					os.Remove(der.config.SavePath + DownloadingFileSuffix) // 删除断点信息
				}

				return
			}

			if !der.config.Testing {
				der.recordBreakPoint()
			}

			// 统计各线程的速度
			go func() {
				for k := range der.status.BlockList {
					go func(k int) {
						old := atomic.LoadInt64(&der.status.BlockList[k].Begin)
						time.Sleep(1 * time.Second)
						atomic.StoreInt64(&der.status.BlockList[k].speed, old)
					}(k)
				}
			}()

			// 速度减慢, 开启监控
			if atomic.LoadInt64(&der.status.Speeds) < atomic.LoadInt64(&der.status.MaxSpeeds)/10 {
				atomic.StoreInt64(&der.status.MaxSpeeds, 0)
				for k := range der.status.BlockList {
					go func(k int) {
						// 重设长时间无响应, 和下载速度为 0 的线程
						// 过滤速度有变化的线程, 不过滤正在等待写入磁盘的线程
						if der.status.BlockList[k].waitingToWrite || atomic.LoadInt64(&der.status.BlockList[k].speed) == 0 {
							return
						}

						// 重设连接
						if r := der.status.BlockList[k].resp; r != nil {
							r.Body.Close()
						}

					}(k)

					// 动态分配新线程
					go func(k int) {
						mu.Lock()

						// 筛选空闲的线程
						index, ok := der.status.BlockList.avaliableThread()
						if !ok { // 没有空的
							mu.Unlock() // 解锁
							return
						}

						end := atomic.LoadInt64(&der.status.BlockList[k].End)
						middle := (atomic.LoadInt64(&der.status.BlockList[k].Begin) + end) / 2

						if end-middle <= 128*1024 { // 如果线程剩余的下载量太少, 不分配空闲线程
							mu.Unlock()
							return
						}

						// 折半
						atomic.StoreInt64(&der.status.BlockList[index].Begin, middle+1)
						atomic.StoreInt64(&der.status.BlockList[index].End, end)

						der.status.BlockList[index].IsFinal = der.status.BlockList[k].IsFinal
						atomic.StoreInt64(&der.status.BlockList[k].End, middle)

						// End 已变, 取消 Final
						der.status.BlockList[k].IsFinal = false

						mu.Unlock()

						der.addExecBlock(index)
					}(k)
				}
			}
			time.Sleep(1 * time.Second) // 监测频率 1 秒
		}
	}()
	return c
}
