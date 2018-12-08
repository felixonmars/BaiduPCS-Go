// Package downloader 多线程下载器, 重构版
package downloader

import (
	"context"
	"errors"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/pcsutil/waitgroup"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/iikira/BaiduPCS-Go/requester"
	"io"
	"net/http"
	"sync"
	"time"
)

type (
	// Downloader 下载
	Downloader struct {
		onExecuteEvent    requester.Event //开始下载事件
		onSuccessEvent    requester.Event //成功下载事件
		onFinishEvent     requester.Event //结束下载事件
		onPauseEvent      requester.Event //暂停下载事件
		onResumeEvent     requester.Event //恢复下载事件
		onCancelEvent     requester.Event //取消下载事件
		monitorCancelFunc context.CancelFunc

		statusCodeBodyCheckFunc func(respBody io.Reader) error
		executeTime             time.Time
		executed                bool
		durl                    string
		loadBalansers           []string
		tryHTTP                 bool
		writer                  io.WriterAt
		client                  *requester.HTTPClient
		config                  *Config
		monitor                 *Monitor
		instanceState           *InstanceState
	}
)

//NewDownloader 初始化Downloader
func NewDownloader(durl string, writer io.WriterAt, config *Config) (der *Downloader) {
	der = &Downloader{
		durl:   durl,
		config: config,
		writer: writer,
	}
	return
}

//SetClient 设置http客户端
func (der *Downloader) SetClient(client *requester.HTTPClient) {
	der.client = client
}

//SetStatusCodeBodyCheckFunc 设置响应状态码出错的检查函数, 当FirstCheckMethod不为HEAD时才有效
func (der *Downloader) SetStatusCodeBodyCheckFunc(f func(respBody io.Reader) error) {
	der.statusCodeBodyCheckFunc = f
}

//TryHTTP 尝试使用 http 连接
func (der *Downloader) TryHTTP(t bool) {
	der.tryHTTP = t
}

func (der *Downloader) lazyInit() {
	if der.config == nil {
		der.config = NewConfig()
	}
	if der.client == nil {
		der.client = requester.NewHTTPClient()
		der.client.SetTimeout(20 * time.Minute)
	}
	if der.monitor == nil {
		der.monitor = NewMonitor()
	}
}

//Execute 开始任务
func (der *Downloader) Execute() error {
	der.lazyInit()

	// 检测
	resp, err := der.client.Req("GET", der.durl, nil, nil)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return err
	}

	// 检测网络错误
	switch resp.StatusCode / 100 {
	case 2: // succeed
	case 4, 5: // error
		if der.statusCodeBodyCheckFunc != nil {
			err = der.statusCodeBodyCheckFunc(resp.Body)
			resp.Body.Close() // 关闭连接
			if err != nil {
				return err
			}
		}
		return errors.New(resp.Status)
	}

	acceptRanges := resp.Header.Get("Accept-Ranges")
	if resp.ContentLength < 0 {
		acceptRanges = ""
	} else {
		acceptRanges = "bytes"
	}

	status := NewDownloadStatus()
	status.totalSize = resp.ContentLength

	var (
		loadBalancerResponses = make([]*LoadBalancerResponse, 0, len(der.loadBalansers)+1)
		handleLoadBalancer    = func(req *http.Request) {
			if req != nil {
				if der.tryHTTP {
					req.URL.Scheme = "http"
				}

				loadBalancer := &LoadBalancerResponse{
					URL:     req.URL.String(),
					Referer: req.Referer(),
				}

				loadBalancerResponses = append(loadBalancerResponses, loadBalancer)
				pcsverbose.Verbosef("DEBUG: download task: URL: %s, Referer: %s\n", loadBalancer.URL, loadBalancer.Referer)
			}
		}
	)

	handleLoadBalancer(resp.Request) // 加入第一个

	// 负载均衡
	wg := waitgroup.NewWaitGroup(10)
	privTimeout := der.client.Client.Timeout
	der.client.SetTimeout(5 * time.Second)
	for _, loadBalanser := range der.loadBalansers {
		wg.AddDelta()
		go func(loadBalanser string) {
			defer wg.Done()

			subResp, subErr := der.client.Req("GET", loadBalanser, nil, nil)
			if subResp != nil {
				subResp.Body.Close() // 不读Body, 马上关闭连接
			}
			if subErr != nil {
				pcsverbose.Verbosef("DEBUG: loadBalanser Error: %s\n", subErr)
				return
			}

			if !ServerEqual(resp, subResp) {
				pcsverbose.Verbosef("DEBUG: loadBalanser not equal to main server: %s\n", subErr)
				return
			}

			if subResp.Request != nil {
				loadBalancerResponses = append(loadBalancerResponses, &LoadBalancerResponse{
					URL: subResp.Request.URL.String(),
				})
			}
			handleLoadBalancer(subResp.Request)

		}(loadBalanser)
	}
	wg.Wait()
	der.client.SetTimeout(privTimeout)

	loadBalancerResponseList := NewLoadBalancerResponseList(loadBalancerResponses)

	//load breakpoint
	err = der.initInstanceState()
	if err != nil {
		return err
	}

	instanceInfo := der.instanceState.Get()
	var (
		dlStatus *DownloadStatus
		ranges   []*Range
	)
	if instanceInfo != nil {
		dlStatus = instanceInfo.DlStatus
		ranges = instanceInfo.Ranges
	}

	if dlStatus != nil {
		status = dlStatus
	}

	// 数据处理
	isRange := ranges != nil && len(ranges) > 0
	if acceptRanges == "" { //不支持多线程
		der.config.parallel = 1
	} else if isRange {
		der.config.parallel = len(ranges)
	} else {
		der.config.parallel = der.config.MaxParallel
		if int64(der.config.parallel) > status.totalSize/int64(MinParallelSize) {
			der.config.parallel = int(status.totalSize/int64(MinParallelSize)) + 1
		}
	}

	if der.config.parallel < 1 {
		der.config.parallel = 1
	}

	der.config.cacheSize = der.config.CacheSize
	blockSize := status.totalSize / int64(der.config.parallel)

	// 如果 cache size 过高, 则调低
	if int64(der.config.cacheSize) > blockSize {
		der.config.cacheSize = int(blockSize)
	}

	pcsverbose.Verbosef("DEBUG: download task CREATED: parallel: %d, cache size: %d\n", der.config.parallel, der.config.cacheSize)

	der.monitor.InitMonitorCapacity(der.config.parallel)

	// 数据平均分配给各个线程
	var (
		begin, end int64
		writeMu    = &sync.Mutex{}
		writerAt   io.WriterAt
	)
	if der.writer == nil {
		writerAt = nil
	} else {
		writerAt = der.writer
	}

	for i := 0; i < der.config.parallel; i++ {
		loadBalancer := loadBalancerResponseList.SequentialGet()
		if loadBalancer == nil {
			continue
		}

		worker := NewWorker(i, loadBalancer.URL, writerAt)
		worker.SetClient(der.client)
		worker.SetCacheSize(der.config.cacheSize)
		worker.SetWriteMutex(writeMu)
		worker.SetReferer(loadBalancer.Referer)

		// 使用第一个连接
		// 断点续传时不使用
		if i == 0 && instanceInfo == nil {
			worker.firstResp = resp
		}

		// 分配线程
		if isRange {
			worker.SetRange(acceptRanges, *ranges[i])
		} else {
			end = int64(i+1) * blockSize
			worker.SetRange(acceptRanges, Range{
				Begin: begin,
				End:   end,
			})
			begin = end + 1
			if i == der.config.parallel-1 {
				worker.wrange.End = status.totalSize - 1
			}
		}
		der.monitor.Append(worker)
	}

	der.monitor.SetStatus(status)
	der.monitor.SetReloadWorker(acceptRanges != "")

	moniterCtx, moniterCancelFunc := context.WithCancel(context.Background())
	der.monitorCancelFunc = moniterCancelFunc

	der.monitor.SetInstanceState(der.instanceState)

	// 开始执行
	der.executeTime = time.Now()
	der.executed = true
	pcsutil.Trigger(der.onExecuteEvent)
	der.monitor.Execute(moniterCtx)

	// 检查错误
	err = der.monitor.Err()
	if err == nil { // 成功
		pcsutil.Trigger(der.onSuccessEvent)
		der.removeInstanceState() // 移除断点续传文件
	}

	// 执行结束
	pcsutil.Trigger(der.onFinishEvent)
	return err
}

//GetDownloadStatusChan 获取下载统计信息
func (der *Downloader) GetDownloadStatusChan() <-chan DlStatus {
	if der.monitor == nil {
		pcsverbose.Verbosef("DEBUG: GetDownloadStatusChan: monitor is nil\n")
		return nil
	}

	status := der.monitor.Status()
	if status == nil {
		pcsverbose.Verbosef("DEBUG: GetDownloadStatusChan: monitor.status is nil\n")
		return nil
	}

	c := make(chan DlStatus)
	go func() {
		for {
			select {
			case <-der.monitor.CompletedChan():
				close(c)
				return
			default:
				if der.executed {
					status.timeElapsed = time.Since(der.executeTime)
					c <- status
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()
	return c
}

//Pause 暂停
func (der *Downloader) Pause() {
	if der.monitor == nil {
		return
	}
	pcsutil.Trigger(der.onPauseEvent)
	der.monitor.Pause()
}

//Resume 恢复
func (der *Downloader) Resume() {
	if der.monitor == nil {
		return
	}
	pcsutil.Trigger(der.onResumeEvent)
	der.monitor.Resume()
}

//Cancel 取消
func (der *Downloader) Cancel() {
	if der.monitor == nil {
		return
	}
	pcsutil.Trigger(der.onCancelEvent)
	pcsutil.Trigger(der.monitorCancelFunc)
}

//PrintAllWorkers 输出所有的worker
func (der *Downloader) PrintAllWorkers() {
	if der.monitor == nil {
		return
	}
	fmt.Println(der.monitor.ShowWorkers())
}

//OnExecute 设置开始下载事件
func (der *Downloader) OnExecute(onExecuteEvent requester.Event) {
	der.onExecuteEvent = onExecuteEvent
}

//OnSuccess 设置成功下载事件
func (der *Downloader) OnSuccess(onSuccessEvent requester.Event) {
	der.onSuccessEvent = onSuccessEvent
}

//OnFinish 设置结束下载事件
func (der *Downloader) OnFinish(onFinishEvent requester.Event) {
	der.onFinishEvent = onFinishEvent
}

//OnPause 设置暂停下载事件
func (der *Downloader) OnPause(onPauseEvent requester.Event) {
	der.onPauseEvent = onPauseEvent
}

//OnResume 设置恢复下载事件
func (der *Downloader) OnResume(onResumeEvent requester.Event) {
	der.onResumeEvent = onResumeEvent
}

//OnCancel 设置取消下载事件
func (der *Downloader) OnCancel(onCancelEvent requester.Event) {
	der.onCancelEvent = onCancelEvent
}
