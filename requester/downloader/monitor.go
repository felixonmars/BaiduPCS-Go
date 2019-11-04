package downloader

import (
	"context"
	"errors"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/iikira/BaiduPCS-Go/requester/transfer"
	"sort"
	"sync"
	"time"
)

var (
	//ErrNoWokers no workers
	ErrNoWokers = errors.New("no workers")
)

type (
	//Monitor 线程监控器
	Monitor struct {
		workers         WorkerList
		status          *transfer.DownloadStatus
		instanceState   *InstanceState
		completed       chan struct{}
		err             error
		resetController *ResetController
		isReloadWorker  bool //是否重载worker, 单线程模式不重载

		// 临时变量
		lastAvaliableIndex int
		allWorkerRanges    transfer.RangeList // worker 的 Range 内存地址, 必须不变
		allWorkerRangesMu  sync.Mutex
	}

	// RangeWorkerFunc 遍历workers的函数
	RangeWorkerFunc func(key int, worker *Worker) bool
)

//NewMonitor 初始化Monitor
func NewMonitor() *Monitor {
	monitor := &Monitor{}
	return monitor
}

func (mt *Monitor) lazyInit() {
	if mt.workers == nil {
		mt.workers = make(WorkerList, 0, 100)
	}
	if mt.status == nil {
		mt.status = transfer.NewDownloadStatus()
	}
	if mt.resetController == nil {
		mt.resetController = NewResetController(80)
	}
}

//InitMonitorCapacity 初始化workers, 用于Append
func (mt *Monitor) InitMonitorCapacity(capacity int) {
	mt.workers = make(WorkerList, 0, capacity)
}

//Append 增加Worker
func (mt *Monitor) Append(worker *Worker) {
	if worker == nil {
		return
	}
	mt.workers = append(mt.workers, worker)
}

//SetWorkers 设置workers, 此操作会覆盖原有的workers
func (mt *Monitor) SetWorkers(workers WorkerList) {
	mt.workers = workers
}

//SetStatus 设置DownloadStatus
func (mt *Monitor) SetStatus(status *transfer.DownloadStatus) {
	mt.status = status
}

//SetInstanceState 设置状态
func (mt *Monitor) SetInstanceState(instanceState *InstanceState) {
	mt.instanceState = instanceState
}

//Status 返回DownloadStatus
func (mt *Monitor) Status() *transfer.DownloadStatus {
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

//GetAvaliableWorker 获取空闲的worker
func (mt *Monitor) GetAvaliableWorker() *Worker {
	workerCount := len(mt.workers)
	for i := mt.lastAvaliableIndex; i < mt.lastAvaliableIndex+workerCount; i++ {
		index := i % workerCount
		worker := mt.workers[index]
		if worker.Completed() {
			mt.lastAvaliableIndex = index
			return worker
		}
	}
	return nil
}

//GetAllWorkersRange 获取所有worker的范围
func (mt *Monitor) GetAllWorkersRange() transfer.RangeList {
	mt.allWorkerRangesMu.Lock()
	defer mt.allWorkerRangesMu.Unlock()

	if mt.allWorkerRanges != nil && len(mt.allWorkerRanges) == len(mt.workers) {
		return mt.allWorkerRanges
	}
	mt.allWorkerRanges = make(transfer.RangeList, 0, len(mt.workers))
	for _, worker := range mt.workers {
		mt.allWorkerRanges = append(mt.allWorkerRanges, worker.GetRange())
	}
	return mt.allWorkerRanges
}

//NumLeftWorkers 剩余的worker数量
func (mt *Monitor) NumLeftWorkers() (num int) {
	for _, worker := range mt.workers {
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

//registerAllCompleted 全部完成则发送消息
func (mt *Monitor) registerAllCompleted() {
	mt.completed = make(chan struct{}, 0)
	var (
		workerNum   = len(mt.workers)
		completeNum = 0
	)

	go func() {
		for {
			time.Sleep(1 * time.Second)

			completeNum = 0
			for _, worker := range mt.workers {
				switch worker.GetStatus().StatusCode() {
				case StatusCodeInternalError:
					mt.err = fmt.Errorf("ERROR: fatal internal error: %s", worker.Err())
					close(mt.completed)
					return
				case StatusCodeSuccessed, StatusCodeCanceled:
					completeNum++
				}
			}
			// status 在 lazyInit 之后, 不可能为空
			// 完成条件: 所有worker 都已经完成, 且 rangeGen 已生成完毕
			gen := mt.status.RangeListGen()
			if completeNum >= workerNum && (gen == nil || gen.IsDone()) { // 已完成
				close(mt.completed)
				return
			}
		}
	}()
}

//ResetFailedAndNetErrorWorkers 重设部分网络错误的worker
func (mt *Monitor) ResetFailedAndNetErrorWorkers() {
	for k := range mt.workers {
		if !mt.resetController.CanReset() {
			continue
		}

		switch mt.workers[k].GetStatus().StatusCode() {
		case StatusCodeNetError:
			pcsverbose.Verbosef("DEBUG: monitor: ResetFailedAndNetErrorWorkers: reset StatusCodeNetError worker, id: %d\n", mt.workers[k].id)
			goto reset
		case StatusCodeFailed:
			pcsverbose.Verbosef("DEBUG: monitor: ResetFailedAndNetErrorWorkers: reset StatusCodeFailed worker, id: %d\n", mt.workers[k].id)
			goto reset
		default:
			continue
		}

	reset:
		mt.workers[k].Reset()
		mt.resetController.AddResetNum()
	}
}

//RangeWorker 遍历worker
func (mt *Monitor) RangeWorker(f RangeWorkerFunc) {
	for k := range mt.workers {
		if !f(k, mt.workers[k]) {
			break
		}
	}
}

//Pause 暂停所有的下载
func (mt *Monitor) Pause() {
	for k := range mt.workers {
		mt.workers[k].Pause()
	}
}

//Resume 恢复所有的下载
func (mt *Monitor) Resume() {
	for k := range mt.workers {
		mt.workers[k].Resume()
	}
}

// TryAddNewWork 尝试加入新range
func (mt *Monitor) TryAddNewWork() {
	if mt.status == nil {
		return
	}
	gen := mt.status.RangeListGen()
	if gen == nil || gen.IsDone() {
		return
	}

	if !mt.resetController.CanReset() { //能否建立新连接
		return
	}

	avaliableWorker := mt.GetAvaliableWorker()
	if avaliableWorker == nil {
		return
	}

	// 有空闲的range, 执行
	_, r := gen.GenRange()
	if r == nil {
		// 没有range了
		return
	}

	avaliableWorker.SetRange(r)
	avaliableWorker.CleanStatus()

	mt.resetController.AddResetNum()
	pcsverbose.Verbosef("MONITER: worker[%d] add new range: %s\n", avaliableWorker.ID(), r.ShowDetails())
	go avaliableWorker.Execute()
}

// DymanicSplitWorker 动态分配线程
func (mt *Monitor) DymanicSplitWorker(worker *Worker) {
	if !mt.resetController.CanReset() {
		return
	}

	switch worker.status.statusCode {
	case StatusCodeDownloading, StatusCodeFailed, StatusCodeNetError:
	//pass
	default:
		return
	}

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

	mt.resetController.AddResetNum()
	pcsverbose.Verbosef("MONITER: worker duplicated: %d <- %d\n", avaliableWorker.ID(), worker.ID())
	go avaliableWorker.Execute()
}

// ResetWorker 重设长时间无响应, 和下载速度为 0 的 Worker
func (mt *Monitor) ResetWorker(worker *Worker) {
	if !mt.resetController.CanReset() { //达到最大重载次数
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

	mt.resetController.AddResetNum()

	// 重设连接
	pcsverbose.Verbosef("MONITER: worker[%d] reload\n", worker.ID())
	worker.Reset()
}

//Execute 执行任务
func (mt *Monitor) Execute(cancelCtx context.Context) {
	if len(mt.workers) == 0 {
		mt.err = ErrNoWokers
		return
	}

	mt.lazyInit()
	for _, worker := range mt.workers {
		worker.SetDownloadStatus(mt.status)
		go worker.Execute()
	}

	mt.registerAllCompleted() // 注册completed
	ticker := time.NewTicker(990 * time.Millisecond)
	defer ticker.Stop()

	//开始监控
	for {
		select {
		case <-cancelCtx.Done():
			for _, worker := range mt.workers {
				err := worker.Cancel()
				if err != nil {
					pcsverbose.Verbosef("DEBUG: cancel failed, worker id: %d, err: %s\n", worker.ID(), err)
				}
			}
			return
		case <-mt.completed:
			return
		case <-ticker.C:
			// 初始化监控工作
			mt.ResetFailedAndNetErrorWorkers()

			mt.status.UpdateSpeeds() // 更新速度

			// 保存断点信息到文件
			if mt.instanceState != nil {
				mt.instanceState.Put(&transfer.DownloadInstanceInfo{
					DownloadStatus: mt.status,
					Ranges:         mt.GetAllWorkersRange(),
				})
			}

			// 加入新range
			mt.TryAddNewWork()

			// 不重载worker
			if !mt.isReloadWorker {
				continue
			}

			// 速度减慢或者全部失败, 开始监控
			// 只有一个worker时不重设连接
			isLeftWorkersAllFailed := mt.IsLeftWorkersAllFailed()
			if mt.status.SpeedsPerSecond() < mt.status.MaxSpeeds()/5 || isLeftWorkersAllFailed {
				if isLeftWorkersAllFailed {
					pcsverbose.Verbosef("DEBUG: monitor: All workers failed\n")
				}
				mt.status.StoreMaxSpeeds(0) //清空统计

				// 先进行动态分配线程
				pcsverbose.Verbosef("DEBUG: monitor: start duplicate.\n")
				sort.Sort(ByLeftDesc{mt.workers})
				for _, worker := range mt.workers {
					//动态分配线程
					mt.DymanicSplitWorker(worker)
				}

				// 重设长时间无响应, 和下载速度为 0 的线程
				pcsverbose.Verbosef("DEBUG: monitor: start reload.\n")
				for _, worker := range mt.workers {
					mt.ResetWorker(worker)
				}
			} // end if 2
		} //end select
	} //end for
}
