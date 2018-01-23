package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	mu sync.Mutex
)

// Block 下载区块
type Block struct {
	Begin   int64 `json:"begin"`
	End     int64 `json:"end"`
	running int   // 线程的载入量
	Final   bool  `json:"isfinal"` // 最后线程, 因为最后的下载线程, 需要另外做处理
}

type blockList []Block

// isDone 判断线程是否完成下载任务
func (b *Block) isDone() bool {
	// 举个例子来演示 如何判断线程完成任务
	// 假设 文件的大小 (Content-Length) 为 300, 线程数为4
	// 则 Block 划分为
	// 线程0: 0-75; 线程1: 76-150; 线程2: 151-225; 线程3: 225-300
	// 正常情况下, 文件下载完成, 每个线程对应的 Block 会变为 (注意最后一个线程 (线程3))
	// 线程0: 76-75; 线程1: 151-150; 线程2: 226-225; 线程3: 300-300
	// 假设 线程0 出现异常, 调用 setDone 方法, 线程0 会变为
	// 线程0: 0-0
	// 即 Block 的 Begin 和 End 值都为 0 时, 返回 true
	//
	// 最后一个线程状态的判断方法, Begin 和 End 的值相等, 则返回 true
	// 其他, End 值 减去 Begin 值为 -1, 则返回 true
	//
	// 暂时先这么判断吧
	return b.End-b.Begin <= -1 || (b.Final == true && b.End-b.Begin <= 0) || (b.End == 0 && b.Begin == 0)
}

// setDone 设置线程为完成下载任务状态 (简单粗暴)
func (b *Block) setDone() {
	// 只操作 End 部分
	// 避免操作 Begin 部分, 否则可能写文件时, 会出现异常
	if b.Begin == 0 {
		b.End = 0
		return
	}
	b.End = b.Begin - 1
}

// isComplete 判断线程是否空闲,
// 即 线程已完成下载任务 且 线程的载入量 (running 值) 为 0
func (b *Block) isComplete() bool {
	return b.isDone() && b.running == 0
}

// avaliableThread 筛选空闲的线程,
// 返回值, 没有空闲的线程, bool 返回 false
// 找到空闲的线程, int 返回该线程的索引 index
func (bl *blockList) avaliableThread() (int, bool) {
	index := -1
	for k := range *bl {
		if (*bl)[k].isComplete() {
			index = k
			break
		}
	}
	return index, index != -1
}

// isAllDone 检查所有的线程, 是否都完成了下载任务
func (bl *blockList) isAllDone() bool {
	for k := range *bl {
		if (*bl)[k].isDone() {
			continue
		}
		return false
	}
	return true
}

// downloadBlockFn 线程控制器
func (f *FileDl) downloadBlockFn(id int) {
	f.BlockList[id].running++
for_2: // code 为 1 时, 不重试
	// 其他的 code, 无限重试
	for {
		code, err := f.downloadBlock(id)

		// 成功, 退出循环
		if code == 0 || err == nil {
			break
		}

		// 未成功(有错误), 继续
		switch code {
		case 1: //不重试
			break for_2 // break for循环
		case 2:
			// 连接太多, 可能会 connect refuse
			time.Sleep(3 * time.Second)
		case 10: // 无限重试
		default: // 休息 3 秒, 再无限重试
			time.Sleep(3 * time.Second)
		}

		// 重新下载
		f.touchOnError(code, err)
	}

	f.BlockList[id].running--
}

// downloadBlock 文件块下载
// 根据 id 对于的 Block, 创建下载任务
func (f *FileDl) downloadBlock(id int) (code int, err error) {
	if f.BlockList[id].isDone() {
		return 1, errors.New("already done")
	}

	request, err := http.NewRequest("GET", f.URL, nil)
	if err != nil {
		return 1, err
	}

	if f.BlockList[id].End != -1 {
		// 设置 Range 请求头, 给各线程分配内容
		request.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", f.BlockList[id].Begin, f.BlockList[id].End))
	}

	resp, err := f.client.Do(request) // 开始 http 请求
	if err != nil {
		return 2, err
	}

	switch resp.StatusCode {
	case 200, 206:
		// do nothing, continue
	case 416: //Requested Range Not Satisfiable
		// 可能是线程在等待响应时, 已被其他线程重载
		return 1, errors.New("thread already reload")
	case 406: // Not Acceptable
	case 429, 509: // Too Many Requests
		for f.status.Speeds >= f.status.MaxSpeeds/5 {
			// 下载速度若不减慢, 循环就不会退出
			time.Sleep(1 * time.Second)
		}
		return 3, errors.New(resp.Status)
	default:
		fmt.Println("不知道的错误", id, resp.StatusCode, resp.Status) // 调试
		return 2, errors.New(resp.Status)
	}

	defer resp.Body.Close()

	buf := make([]byte, cacheSize)
	loopSize := 0

	for {
		begin := f.BlockList[id].Begin // 用于下文比较

		n, err := resp.Body.Read(buf)

		bufSize := int64(n)
		loopSize += n
		if f.BlockList[id].End != -1 {
			// 检查下载的大小是否超出需要下载的大小
			// 这里End+1是因为http的Range的end是包括在需要下载的数据内的
			// 比如 0-1 的长度其实是2，所以这里end需要+1
			needSize := f.BlockList[id].End + 1 - f.BlockList[id].Begin

			// 已完成 (未雨绸缪)
			if needSize <= 0 {
				return 1, errors.New("already complete")
			}

			if bufSize > needSize {
				// 数据大小不正常
				// 一般是该线程已被重载

				// 也可能是因为网络环境不好导致
				// 比如用中国电信下载国外文件

				// 设置数据大小来去掉多余数据
				// 并结束这个线程的下载

				bufSize = needSize
				n = int(needSize)
				err = io.EOF
			}
		}

		// 将缓冲数据写入硬盘
		f.File.WriteAt(buf[:n], begin)

		// 两次 begin 不相等, 可能已有新的空闲线程参与
		// 旧线程应该被结束
		if begin != f.BlockList[id].Begin {
			return 1, errors.New("thread already reload")
		}

		// 更新已下载大小
		atomic.AddInt64(&f.status.Downloaded, bufSize)
		atomic.AddInt64(&f.BlockList[id].Begin, int64(n))

		// reload connection (百度的限制)
		if loopSize == 256*1024 {
			return 10, errors.New("reach to loop size, reload connection")
		}

		if err != nil {
			// 下载数据可能出现异常, 重新下载
			if f.BlockList[id].End != -1 && !f.BlockList[id].isDone() {
				return 11, fmt.Errorf("download failed, %s, reset", err)
			}
			switch {
			case err == io.EOF:
				// 数据已经下载完毕
				return 0, nil
			default:
				// 其他错误, 返回
				return 5, err
			}
		}
	}
}

// blockMonitor 延迟监控各线程状态,
// 管理空闲 (已完成下载任务) 的线程,
// 清除长时间无响应, 和下载速度为 0 的线程
func (f *FileDl) blockMonitor() <-chan struct{} {
	c := make(chan struct{})
	go func() {
		for {
			// 下载完毕, 线程全部完成下载任务, 发送结束信号
			if f.BlockList.isAllDone() {
				c <- struct{}{}
				os.Remove(f.File.Name() + DownloadingFileSuffix) // 删除断点信息

				if f.Size <= 0 {
					fileInfo, err := f.File.Stat()
					if err == nil && fileInfo.Size() > f.status.Downloaded {
						f.File.Truncate(f.status.Downloaded)
					}
				} else {
					f.File.Truncate(f.Size)
				}

				break
			}

			f.recordBreakPoint()

			for k := range f.BlockList {
				go func(k int) {
					// 过滤已完成下载任务的线程
					if f.BlockList[k].isDone() {
						return
					}

					// 清除长时间无响应, 和下载速度为 0 的线程
					go func(k int) {
						// 设 old 速度监测点, 2 秒后检查速度有无变化
						old := f.BlockList[k].Begin
						time.Sleep(2 * time.Second)
						// 过滤 速度有变化, 或 2 秒内完成了下载任务 的线程
						if old != f.BlockList[k].Begin || f.BlockList[k].isDone() {
							return
						}

						// 筛选出 长时间无响应, 和下载速度为 0 的线程
						// 然后尝试清除该线程, 选出其他 空闲的线程, 重新添加下载任务
						mu.Lock() // 加锁, 防止出现重复添加线程的状况 (实验阶段)

						// 筛选空闲的线程
						index, ok := f.BlockList.avaliableThread()
						if !ok { // 没有空的
							mu.Unlock() // 解锁
							return
						}

						// 复制 旧线程 的信息到 空闲的新线程
						f.BlockList[index].End = f.BlockList[k].End
						f.BlockList[index].Begin = f.BlockList[k].Begin
						f.BlockList[index].Final = f.BlockList[k].Final

						f.BlockList[k].setDone() // 清除旧线程

						mu.Unlock()              // 解锁
						f.downloadBlockFn(index) // 添加任务
					}(k)

					// 动态分配新线程
					go func(k int) {
						mu.Lock()

						// 筛选空闲的线程
						index, ok := f.BlockList.avaliableThread()
						if !ok { // 没有空的
							mu.Unlock() // 解锁
							return
						}

						middle := (f.BlockList[k].Begin + f.BlockList[k].End) / 2
						if f.BlockList[k].End-middle <= 102400 { // 如果线程剩余的下载量太少, 不分配空闲线程
							mu.Unlock()
							return
						}

						// 折半
						f.BlockList[index].Begin = middle + 1
						f.BlockList[index].End = f.BlockList[k].End
						f.BlockList[index].Final = f.BlockList[k].Final
						f.BlockList[k].End = middle

						mu.Unlock()
						// fmt.Println(k, f.BlockList[k], "=>", index, f.BlockList[index])
						f.downloadBlockFn(index)
					}(k)

				}(k)
			}
			time.Sleep(1 * time.Second) // 监测频率 1 秒
		}
	}()
	return c
}
