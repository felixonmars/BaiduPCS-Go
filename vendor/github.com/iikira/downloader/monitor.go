package downloader

import (
	"os"
	"sync"
	"time"
)

var (
	mu sync.Mutex
)

// blockMonitor 延迟监控各线程状态,
// 管理空闲 (已完成下载任务) 的线程,
// 清除长时间无响应, 和下载速度为 0 的线程
func (der *Downloader) blockMonitor() <-chan struct{} {
	c := make(chan struct{})
	go func() {
		for {
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

			for k := range der.status.BlockList {
				go func(k int) {
					// 过滤已完成下载任务的线程
					if der.status.BlockList[k].isDone() {
						return
					}

					// 清除长时间无响应, 和下载速度为 0 的线程
					go func(k int) {
						// 设 old 速度监测点, 2 秒后检查速度有无变化
						old := der.status.BlockList[k].Begin
						time.Sleep(2 * time.Second)
						// 过滤 速度有变化, 或 2 秒内完成了下载任务 的线程, 不过滤正在等待写入磁盘的线程
						if der.status.BlockList[k].waitingToWrite || old != der.status.BlockList[k].Begin || der.status.BlockList[k].isDone() {
							return
						}

						// 筛选出 长时间无响应, 和下载速度为 0 的线程
						// 然后尝试清除该线程, 选出其他 空闲的线程, 重新添加下载任务
						mu.Lock() // 加锁, 防止出现重复添加线程的状况 (实验阶段)

						// 筛选空闲的线程
						index, ok := der.status.BlockList.avaliableThread()
						if !ok { // 没有空的
							mu.Unlock() // 解锁
							return
						}

						// 复制 旧线程 的信息到 空闲的线程
						der.status.BlockList[index].End = der.status.BlockList[k].End
						der.status.BlockList[index].Begin = der.status.BlockList[k].Begin
						der.status.BlockList[index].IsFinal = der.status.BlockList[k].IsFinal

						der.status.BlockList[k].setDone() // 清除旧线程

						mu.Unlock()             // 解锁
						der.addExecBlock(index) // 添加任务
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

						middle := (der.status.BlockList[k].Begin + der.status.BlockList[k].End) / 2
						if der.status.BlockList[k].End-middle <= 102400 { // 如果线程剩余的下载量太少, 不分配空闲线程
							mu.Unlock()
							return
						}

						// 折半
						der.status.BlockList[index].Begin = middle + 1
						der.status.BlockList[index].End = der.status.BlockList[k].End
						der.status.BlockList[index].IsFinal = der.status.BlockList[k].IsFinal
						der.status.BlockList[k].End = middle

						// End 已变, 取消 Final
						der.status.BlockList[k].IsFinal = false

						mu.Unlock()

						der.addExecBlock(index)
					}(k)

				}(k)
			}
			time.Sleep(1 * time.Second) // 监测频率 1 秒
		}
	}()
	return c
}
