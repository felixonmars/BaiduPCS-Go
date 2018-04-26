package downloader

import (
	"context"
	"errors"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	//ErrNoWokers no workers
	ErrNoWokers = errors.New("no workers")
)

//Monitor 线程监控器
type Monitor struct {
	workers        []*Worker
	status         *DownloadStatus
	instanceState  *InstanceState
	completed      <-chan struct{}
	err            error
	dymanicMu      sync.Mutex
	isReloadWorker bool //是否重载worker
}

//NewMonitor 初始化Monitor
func NewMonitor() *Monitor {
	monitor := &Monitor{}
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

//InitMonitorCapacity 初始化workers, 用于Append
func (mt *Monitor) InitMonitorCapacity(capacity int) {
	mt.workers = make([]*Worker, 0, capacity)
}

//Append 增加Worker
func (mt *Monitor) Append(worker *Worker) {
	if worker == nil {
		return
	}
	mt.workers = append(mt.workers, worker)
}

//SetWorkers 设置workers, 此操作会覆盖原有的workers
func (mt *Monitor) SetWorkers(workers []*Worker) {
	mt.workers = workers
}

//SetStatus 设置DownloadStatus
func (mt *Monitor) SetStatus(status *DownloadStatus) {
	mt.status = status
}

//SetInstanceState 设置状态
func (mt *Monitor) SetInstanceState(instanceState *InstanceState) {
	mt.instanceState = instanceState
}

//Status 返回DownloadStatus
func (mt *Monitor) Status() *DownloadStatus {
	return mt.status
}

//Err 返回遇到的错误
func (mt *Monitor) Err() error {
	return mt.err
}

//CompletedChan 获取completed chan
func (mt *Monitor) CompletedChan() <-chan struct{} {
	return mt.completed
}

//GetSpeedsPerSecondFunc 获取每秒的速度, 返回获取速度的函数
func (mt *Monitor) GetSpeedsPerSecondFunc() func() int64 {
	if mt.status == nil {
		return nil
	}
	old := mt.status.Downloaded()
	nowTime := time.Now()
	return func() int64 {
		d := mt.status.Downloaded() - old
		s := time.Since(nowTime)

		old = mt.status.Downloaded()
		nowTime = time.Now()
		return int64(float64(d) / s.Seconds())
	}
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

		ranges = append(ranges, worker.GetRange())
	}
	return
}

//NumLeftWorkers 剩余的worker数量
func (mt *Monitor) NumLeftWorkers() (num int) {
	for _, worker := range mt.workers {
		if worker == nil {
			continue
		}

		if !worker.Completed() {
			num++
		}
	}
	return
}

//SetReloadWorker 是否重载worker
func (mt *Monitor) SetReloadWorker(b bool) {
	mt.isReloadWorker = b
}

//IsLeftWorkersAllFailed 剩下的线程是否全部失败
func (mt *Monitor) IsLeftWorkersAllFailed() bool {
	failedNum := 0
	for _, worker := range mt.workers {
		if worker == nil {
			continue
		}
		if worker.Completed() {
			continue
		}

		if !worker.Failed() {
			failedNum++
			return false
		}
	}
	return failedNum != 0
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

				switch worker.GetStatus().StatusCode() {
				case StatusCodeInternalError:
					mt.err = fmt.Errorf("ERROR: fatal internal error: %s", worker.Err())
					close(c)
					return
				case StatusCodeSuccessed, StatusCodeCanceled:
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

//ResetAllFailedAndNetErrorWorkers 重设所有网络错误的worker
func (mt *Monitor) ResetAllFailedAndNetErrorWorkers() {
	for k := range mt.workers {
		if mt.workers[k] == nil {
			continue
		}

		switch mt.workers[k].GetStatus().StatusCode() {
		case StatusCodeNetError:
			pcsverbose.Verbosef("DEBUG: monitor: ResetAllFailedAndNetErrorWorkers: reset StatusCodeNetError worker, id: %d\n", mt.workers[k].id)
			mt.workers[k].Reset()
		case StatusCodeFailed:
			pcsverbose.Verbosef("DEBUG: monitor: ResetAllFailedAndNetErrorWorkers: reset StatusCodeFailed worker, id: %d\n", mt.workers[k].id)
			mt.workers[k].Reset()
		}
	}
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
func (mt *Monitor) Execute(cancelCtx context.Context) {
	if len(mt.workers) == 0 {
		mt.err = ErrNoWokers
		return
	}

	mt.lazyInit()
	for _, worker := range mt.workers {
		if worker == nil {
			continue
		}
		worker.SetDownloadStatus(mt.status)
		go worker.Execute()
	}

	mt.completed = mt.AllCompleted()

	var (
		maxReloadNumFunc = func() int32 {
			total := len(mt.workers)
			num := int32(32 * float64(total-mt.NumLeftWorkers()) / float64(total))
			return num + 5
		}
		reloadNum int32
		// 重设所有网络错误的worker, 30s间隔
		nowTime                          = time.Now()
		resetAllFailedAndNetErrorWorkers = func() {
			if time.Since(nowTime) < 10*time.Second {
				return
			}
			pcsverbose.Verbosef("DEBUG: monitor: resetAllFailedAndNetErrorWorkers start\n")
			nowTime = time.Now()
			mt.ResetAllFailedAndNetErrorWorkers()
		}
	)

	//开始监控
	for {
		select {
		case <-cancelCtx.Done():
			for _, worker := range mt.workers {
				if worker == nil {
					continue
				}

				err := worker.Cancel()
				if err != nil {
					pcsverbose.Verbosef("DEBUG: cancel failed, worker id: %d, err: %s\n", worker.ID(), err)
				}
			}
			return
		case <-mt.completed:
			return
		default:
			time.Sleep(1 * time.Second)

			// 初始化监控工作
			resetAllFailedAndNetErrorWorkers()
			atomic.StoreInt32(&reloadNum, 0)
			mt.status.updateSpeeds()

			if mt.instanceState != nil {
				mt.instanceState.Put(&InstanceInfo{
					DlStatus: mt.status,
					Ranges:   mt.GetAllWorkersRange(),
				})
			}

			// 不重载worker
			if !mt.isReloadWorker {
				continue
			}

			// 速度减慢或者全部失败, 开始监控
			isLeftWorkersAllFailed := mt.IsLeftWorkersAllFailed()
			if mt.status.SpeedsPerSecond() < mt.status.MaxSpeeds()/20 || isLeftWorkersAllFailed {
				pcsverbose.Verbosef("DEBUG: monitor: start reload.\n")
				if isLeftWorkersAllFailed {
					pcsverbose.Verbosef("DEBUG: monitor: All workers failed\n")
				}
				mt.status.ResetMaxSpeeds() //清空统计

				for _, worker := range mt.workers {
					// 重设长时间无响应, 和下载速度为 0 的线程
					go func(worker *Worker) {
						if atomic.LoadInt32(&reloadNum) > maxReloadNumFunc() { //达到最大重载次数
							return
						}

						if worker.Completed() {
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
						case StatusCodePending, StatusCodeReseted:
							fallthrough
						case StatusCodeWaitToWrite: // 正在写入数据
							fallthrough
						case StatusCodePaused: // 已暂停
							// 忽略, 返回
							return
						}

						atomic.AddInt32(&reloadNum, 1) // 增加重载次数

						// 重设连接
						pcsverbose.Verbosef("MONITER: worker reload, worker id: %d\n", worker.ID())
						worker.Reset()
					}(worker)
				}

				//下载快完成了, 动态分配线程
				if float64(mt.status.Downloaded()) > float64(mt.status.totalSize)*0.8 {
					pcsverbose.Verbosef("DEBUG: monitor: start duplicate.\n")
					for k := range mt.workers {
						if mt.workers[k] == nil {
							continue
						}
						//动态分配线程
						go func(worker *Worker) {
							//过滤速度为0的worker
							if worker.GetSpeedsPerSecond() == 0 {
								return
							}

							mt.dymanicMu.Lock()
							defer mt.dymanicMu.Unlock()

							// 筛选空闲的Worker
							avaliableWorker := mt.GetAvaliableWorker()
							if avaliableWorker == nil || worker == avaliableWorker { // 没有空的
								return
							}

							workerRange := worker.GetRange()

							end := workerRange.LoadEnd()
							middle := (workerRange.LoadBegin() + end) / 2

							if end-middle < MinParallelSize/5 { // 如果线程剩余的下载量太少, 不分配空闲线程
								return
							}

							// 折半

							avaliableWorkerRange := avaliableWorker.GetRange()
							avaliableWorkerRange.StoreBegin(middle + 1)
							avaliableWorkerRange.StoreEnd(end)

							avaliableWorker.CleanStatus()

							workerRange.StoreEnd(middle)

							pcsverbose.Verbosef("MONITER: worker duplicated: %d <- %d\n", avaliableWorker.ID(), worker.ID())
							go avaliableWorker.Execute()
							time.Sleep(10 * time.Microsecond)
						}(mt.workers[k])
					} //end for
				} // end if 2
			} //end if 1
		} //end selete
	} //end for
}

//ShowWorkers 返回所有worker的状态
func (mt *Monitor) ShowWorkers() string {
	var (
		builder = &strings.Builder{}
		tb      = pcstable.NewTable(builder)
	)
	tb.SetHeader([]string{"#", "status", "range", "left", "speeds", "error"})
	mt.RangeWorker(func(key int, worker *Worker) bool {
		wrange := worker.GetRange()
		tb.Append([]string{fmt.Sprint(worker.ID()), worker.GetStatus().StatusText(), wrange.String(), strconv.FormatInt(wrange.Len(), 10), strconv.FormatInt(worker.GetSpeedsPerSecond(), 10), fmt.Sprint(worker.Err())})
		return true
	})
	tb.Render()
	return "\n" + builder.String()
}
