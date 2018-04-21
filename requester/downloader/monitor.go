package downloader

import (
	"context"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Monitor 线程监控器
type Monitor struct {
	workers       []*Worker
	status        *DownloadStatus
	instanceState *InstanceState
	completed     <-chan struct{}
	dymanicMu     sync.Mutex
}

//NewMonitor 初始化Monitor
func NewMonitor(parallel int) *Monitor {
	monitor := &Monitor{
		workers: make([]*Worker, 0, parallel),
	}
	monitor.lazyInit()
	return monitor
}

func (mt *Monitor) lazyInit() {
	if mt.workers == nil {
		mt.workers = make([]*Worker, 0, 100)
	}
	if mt.status == nil {
		mt.status = NewDownloadStatus()
	}
}

//Append 增加Worker
func (mt *Monitor) Append(worker *Worker) {
	if worker == nil {
		return
	}
	mt.workers = append(mt.workers, worker)
}

//SetInstanceState 设置状态
func (mt *Monitor) SetInstanceState(instanceState *InstanceState) {
	mt.instanceState = instanceState
}

//GetSpeedsPerSecond 获取每秒的速度
func (mt *Monitor) GetSpeedsPerSecond() int64 {
	var speeds int64
	for _, worker := range mt.workers {
		if worker == nil {
			continue
		}

		speeds += worker.GetSpeedsPerSecond()
	}
	return speeds
}

//GetAvaliableWorker 获取空闲的worker
func (mt *Monitor) GetAvaliableWorker() *Worker {
	for _, worker := range mt.workers {
		if worker == nil {
			continue
		}

		if worker.Completed() {
			return worker
		}
	}
	return nil
}

//GetAllWorkersRange 获取所以worker的范围
func (mt *Monitor) GetAllWorkersRange() (ranges []*Range) {
	ranges = make([]*Range, 0, len(mt.workers))
	for _, worker := range mt.workers {
		if worker == nil {
			continue
		}

		ranges = append(ranges, worker.wrange)
	}
	return
}

//AllFailed 是否全部失败
func (mt *Monitor) AllFailed() bool {
	for _, worker := range mt.workers {
		if worker == nil {
			continue
		}

		if !worker.Failed() {
			return false
		}
	}
	return true
}

//AllCompleted 全部完成则发送消息
func (mt *Monitor) AllCompleted() <-chan struct{} {
	var (
		c           = make(chan struct{}, 0)
		workerNum   = len(mt.workers)
		completeNum = 0
	)

	go func() {
		for {
			completeNum = 0
			for _, worker := range mt.workers {
				if worker == nil {
					continue
				}

				if worker.Completed() {
					completeNum++
				}
			}
			if completeNum >= workerNum {
				close(c)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()
	return c
}

//RangeWorker 遍历worker
func (mt *Monitor) RangeWorker(f func(key int, worker *Worker) bool) {
	for k := range mt.workers {
		if mt.workers[k] == nil {
			continue
		}
		if !f(k, mt.workers[k]) {
			break
		}
	}
}

//Pause 暂停所有的下载
func (mt *Monitor) Pause() {
	for k := range mt.workers {
		if mt.workers[k] == nil {
			continue
		}

		mt.workers[k].Pause()
	}
}

//Resume 恢复所有的下载
func (mt *Monitor) Resume() {
	for k := range mt.workers {
		if mt.workers[k] == nil {
			continue
		}

		mt.workers[k].Resume()
	}
}

//Execute 执行任务
func (mt *Monitor) Execute(ctx context.Context) {
	mt.lazyInit()
	for _, worker := range mt.workers {
		if worker == nil {
			continue
		}
		worker.AppendOthersAdd(&mt.status.downloaded)
		go worker.Execute()
	}

	mt.completed = mt.AllCompleted()
	ranges := mt.GetAllWorkersRange()

	exitErr := make(chan error)
	//开始监控
	for {
		select {
		case err := <-exitErr:
			pcsverbose.Verbosef("ERROR: fatal error: %s\n", err)
			return
		case <-ctx.Done():
			var err error
			for _, worker := range mt.workers {
				if worker == nil {
					continue
				}

				err = worker.Cancel()
				if err != nil {
					pcsverbose.Verbosef("DEBUG: cancel failed, worker id: %d, err: %s\n", worker.id, err)
				}
			}
			return
		case <-mt.completed:
			return
		default:
			if mt.instanceState != nil {
				mt.instanceState.Put(&InstanceInfo{
					DlStatus: mt.status,
					Ranges:   ranges,
				})
			}
			speeds := mt.GetSpeedsPerSecond()
			mt.status.StoreSpeedsPerSecond(speeds)
			if speeds > mt.status.maxSpeeds {
				mt.status.StoreMaxSpeeds(speeds)
			}

			for _, worker := range mt.workers {
				if worker == nil {
					continue
				}

				switch worker.GetStatus().StatusCode() {
				case StatusCodeInternalError:
					exitErr <- worker.err
					return
				}
			}

			// 速度减慢或者全部失败, 开始监控
			if speeds < mt.status.maxSpeeds/10 || mt.AllFailed() {
				mt.status.StoreMaxSpeeds(0) //清空统计
				for k := range mt.workers {
					if mt.workers[k] == nil {
						continue
					}

					// 重设长时间无响应, 和下载速度为 0 的线程
					go func(worker *Worker) {
						if !worker.inited || worker.Completed() || !worker.Resetable() {
							return
						}

						// 忽略正在写入数据到硬盘的
						// 过滤速度有变化的线程
						status := worker.GetStatus()
						speeds := worker.GetSpeedsPerSecond()
						if speeds != 0 {
							return
						}

						switch status.StatusCode() {
						case StatusCodeWaitToWrite: // 正在写入数据
							fallthrough
						case StatusCodePaused: // 已暂停
							// 忽略, 返回
							return
						case StatusCodeNetError:
							fallthrough
						case StatusCodeTooManyConnections:
							// do nothing
						}

						// 重设连接
						pcsverbose.Verbosef("MONITER: worker reload, worker id: %d\n", worker.id)
						worker.Reset()
					}(mt.workers[k])

					//动态分配线程
					go func(worker *Worker) {
						if !worker.Resetable() { // 不可重设
							return
						}
						mt.dymanicMu.Lock()
						defer mt.dymanicMu.Unlock()

						// 筛选空闲的Worker
						avaliableWorker := mt.GetAvaliableWorker()
						if avaliableWorker == nil || worker == avaliableWorker { // 没有空的
							return
						}

						end := worker.wrange.LoadEnd()
						middle := (worker.wrange.LoadBegin() + end) / 2

						if end-middle <= MinParallelSize { // 如果线程剩余的下载量太少, 不分配空闲线程
							return
						}

						// 折半
						avaliableWorker.wrange = &Range{
							Begin: middle + 1,
							End:   end,
						}
						avaliableWorker.CleanStatus()

						worker.wrange.StoreEnd(middle)

						pcsverbose.Verbosef("MONITER: thread duplicated: %d <- %d\n", avaliableWorker.id, worker.id)
						go avaliableWorker.Execute()
						<-avaliableWorker.InitedChan()
					}(mt.workers[k])
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}

//ShowWorkers 返回所有worker的状态
func (mt *Monitor) ShowWorkers() string {
	builder := &strings.Builder{}
	tb := pcstable.NewTable(builder)
	tb.SetHeader([]string{"#", "status", "range", "speeds", "error"})
	mt.RangeWorker(func(key int, worker *Worker) bool {
		tb.Append([]string{fmt.Sprint(worker.id), worker.GetStatus().StatusText(), worker.wrange.String(), strconv.FormatInt(worker.GetSpeedsPerSecond(), 10), fmt.Sprint(worker.err)})
		return true
	})
	tb.Render()
	return builder.String()
}
