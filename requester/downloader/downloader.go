// Package downloader 多线程下载器, 重构版
package downloader

import (
	"context"
	"errors"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/pcsutil/waitgroup"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/BaiduPCS-Go/requester/downloader/cachepool"
	"github.com/iikira/BaiduPCS-Go/requester/downloader/prealloc"
	"github.com/iikira/BaiduPCS-Go/requester/rio/speeds"
	"io"
	"net/http"
	"sync"
	"time"
)

type (
	// Downloader 下载
	Downloader struct {
		onExecuteEvent        requester.Event    //开始下载事件
		onSuccessEvent        requester.Event    //成功下载事件
		onFinishEvent         requester.Event    //结束下载事件
		onPauseEvent          requester.Event    //暂停下载事件
		onResumeEvent         requester.Event    //恢复下载事件
		onCancelEvent         requester.Event    //取消下载事件
		onDownloadStatusEvent DownloadStatusFunc //状态处理事件

		monitorCancelFunc context.CancelFunc

		durlCheckFunc           DURLCheckFunc
		statusCodeBodyCheckFunc StatusCodeBodyCheckFunc
		executeTime             time.Time
		durl                    string
		loadBalansers           []string
		writer                  io.WriterAt
		client                  *requester.HTTPClient
		config                  *Config
		monitor                 *Monitor
		instanceState           *InstanceState
	}

	// DURLCheckFunc 下载URL检测函数
	DURLCheckFunc func(client *requester.HTTPClient, durl string) (contentLength int64, resp *http.Response, err error)
	// StatusCodeBodyCheckFunc 响应状态码出错的检查函数
	StatusCodeBodyCheckFunc func(respBody io.Reader) error
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

// SetDURLCheckFunc 设置下载URL检测函数
func (der *Downloader) SetDURLCheckFunc(f DURLCheckFunc) {
	der.durlCheckFunc = f
}

//SetStatusCodeBodyCheckFunc 设置响应状态码出错的检查函数, 当FirstCheckMethod不为HEAD时才有效
func (der *Downloader) SetStatusCodeBodyCheckFunc(f StatusCodeBodyCheckFunc) {
	der.statusCodeBodyCheckFunc = f
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
	if der.durlCheckFunc == nil {
		der.durlCheckFunc = DefaultDURLCheckFunc
	}
}

// SelectParallel 获取合适的 parallel
func (der *Downloader) SelectParallel(acceptRanges string, maxParallel int, totalSize int64, instanceRangeList RangeList) (parallel int) {
	isRange := instanceRangeList != nil && len(instanceRangeList) > 0
	if acceptRanges == "" { //不支持多线程
		parallel = 1
	} else if isRange {
		parallel = len(instanceRangeList)
	} else {
		parallel = der.config.MaxParallel
		if int64(parallel) > totalSize/int64(MinParallelSize) {
			parallel = int(totalSize/int64(MinParallelSize)) + 1
		}
	}

	if parallel < 1 {
		parallel = 1
	}
	return
}

// SelectBlockSizeAndInitRangeGen 获取合适的 BlockSize, 和初始化 RangeGen
func (der *Downloader) SelectBlockSizeAndInitRangeGen(status *DownloadStatus, parallel int) (blockSize int64, initErr error) {
	// Range 生成器
	if status.gen == nil {
		switch der.config.Mode {
		case RangeGenMode_Default:
			status.gen = NewRangeListGenDefault(status.totalSize, 0, 0, parallel)
			blockSize = status.gen.LoadBlockSize()
		case RangeGenMode_BlockSize:
			b2 := status.totalSize / int64(parallel)
			if b2 > der.config.BlockSize { // 选小的BlockSize, 以更高并发
				blockSize = der.config.BlockSize
			} else {
				blockSize = b2
			}

			status.gen = NewRangeListGenBlockSize(status.totalSize, 0, blockSize)
		default:
			initErr = ErrUnknownRangeGenMode
			return
		}
	} else {
		blockSize = status.gen.blockSize
	}
	return
}

// SelectCacheSize 获取合适的 cacheSize
func (der *Downloader) SelectCacheSize(confCacheSize int, blockSize int64) (cacheSize int) {
	if int64(confCacheSize) > blockSize {
		// 如果 cache size 过高, 则调低
		cacheSize = int(blockSize)
	} else {
		cacheSize = confCacheSize
	}
	return
}

// DefaultDURLCheckFunc 默认的 DURLCheckFunc
func DefaultDURLCheckFunc(client *requester.HTTPClient, durl string) (contentLength int64, resp *http.Response, err error) {
	resp, err = client.Req("GET", durl, nil, nil)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return 0, nil, err
	}
	return resp.ContentLength, resp, nil
}

//Execute 开始任务
func (der *Downloader) Execute() error {
	der.lazyInit()

	// 检测
	contentLength, resp, err := der.durlCheckFunc(der.client, der.durl)
	if err != nil {
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
	if contentLength < 0 {
		acceptRanges = ""
	} else {
		acceptRanges = "bytes"
	}

	var (
		loadBalancerResponses = make([]*LoadBalancerResponse, 0, len(der.loadBalansers)+1)
		handleLoadBalancer    = func(req *http.Request) {
			if req != nil {
				if der.config.TryHTTP {
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

			subContentLength, subResp, subErr := der.durlCheckFunc(der.client, loadBalanser)
			if subResp != nil {
				subResp.Body.Close() // 不读Body, 马上关闭连接
			}
			if subErr != nil {
				pcsverbose.Verbosef("DEBUG: loadBalanser Error: %s\n", subErr)
				return
			}

			// 检测长度
			if contentLength != subContentLength {
				pcsverbose.Verbosef("DEBUG: loadBalanser Content-Length not equal to main server\n")
				return
			}

			if !ServerEqual(resp, subResp) {
				pcsverbose.Verbosef("DEBUG: loadBalanser not equal to main server\n")
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
	err = der.initInstanceState(der.config.InstanceStateStorageFormat)
	if err != nil {
		return err
	}

	var (
		bii        = der.instanceState.Get()
		isInstance = bii != nil // 是否存在断点信息
		status     *DownloadStatus
	)
	if !isInstance {
		bii = &InstanceInfo{}
	}

	if bii.DownloadStatus != nil {
		// 使用断点信息的状态
		status = bii.DownloadStatus
	} else {
		// 新建状态
		status = NewDownloadStatus()
		status.totalSize = contentLength
	}

	// 设置限速
	if der.config.MaxRate > 0 {
		status.rateLimit = speeds.NewRateLimit(der.config.MaxRate)
		defer status.rateLimit.Stop()
	}

	// 数据处理
	parallel := der.SelectParallel(acceptRanges, der.config.MaxParallel, status.totalSize, bii.Ranges) // 实际的下载并行量
	blockSize, err := der.SelectBlockSizeAndInitRangeGen(status, parallel)                             // 实际的BlockSize
	if err != nil {
		return err
	}

	cacheSize := der.SelectCacheSize(der.config.CacheSize, blockSize) // 实际下载缓存
	cachepool.SetSyncPoolSize(cacheSize)                              // 调整pool大小

	pcsverbose.Verbosef("DEBUG: download task CREATED: parallel: %d, cache size: %d\n", parallel, cacheSize)

	der.monitor.InitMonitorCapacity(parallel)

	var writer Writer
	if !der.config.IsTest {
		// 尝试修剪文件
		if fder, ok := der.writer.(Fder); ok {
			err = prealloc.PreAlloc(fder.Fd(), status.totalSize)
			if err != nil {
				pcsverbose.Verbosef("DEBUG: truncate file error: %s\n", err)
			}
		}
		writer = der.writer // 非测试模式, 赋值writer
	}

	// 数据平均分配给各个线程
	isRange := bii.Ranges != nil && len(bii.Ranges) > 0
	if !isRange {
		bii.Ranges = make(RangeList, 0, parallel)
		for i := 0; i < cap(bii.Ranges); i++ {
			_, r := status.gen.GenRange()
			if r == nil { // 没有了（不正常）
				break
			}
			bii.Ranges = append(bii.Ranges, r)
		}
	}

	var (
		writeMu = &sync.Mutex{}
	)
	for k, r := range bii.Ranges {
		loadBalancer := loadBalancerResponseList.SequentialGet()
		if loadBalancer == nil {
			continue
		}

		worker := NewWorker(k, loadBalancer.URL, writer)
		worker.SetClient(der.client)
		worker.SetCacheSize(cacheSize)
		worker.SetWriteMutex(writeMu)
		worker.SetReferer(loadBalancer.Referer)

		// 使用第一个连接
		// 断点续传时不使用
		if k == 0 && !isInstance {
			worker.firstResp = resp
		}

		worker.SetAcceptRange(acceptRanges)
		worker.SetRange(r) // 分配Range
		der.monitor.Append(worker)
	}

	der.monitor.SetStatus(status)

	// 服务器不支持断点续传, 或者单线程下载, 都不重载worker
	der.monitor.SetReloadWorker(parallel > 1)

	moniterCtx, moniterCancelFunc := context.WithCancel(context.Background())
	der.monitorCancelFunc = moniterCancelFunc

	der.monitor.SetInstanceState(der.instanceState)

	// 开始执行
	der.executeTime = time.Now()
	pcsutil.Trigger(der.onExecuteEvent)
	der.downloadStatusEvent() // 启动执行状态处理事件
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

//downloadStatusEvent 执行状态处理事件
func (der *Downloader) downloadStatusEvent() {
	if der.onDownloadStatusEvent == nil {
		return
	}

	status := der.monitor.Status()
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-der.monitor.completed:
				return
			case <-ticker.C:
				der.onDownloadStatusEvent(status, der.monitor.RangeWorker)
			}
		}
	}()
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

//OnDownloadStatusEvent 设置状态处理函数
func (der *Downloader) OnDownloadStatusEvent(f DownloadStatusFunc) {
	der.onDownloadStatusEvent = f
}
