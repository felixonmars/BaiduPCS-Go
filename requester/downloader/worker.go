package downloader

import (
	"context"
	"errors"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/BaiduPCS-Go/requester/downloader/cachepool"
	"github.com/iikira/BaiduPCS-Go/requester/rio/speeds"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type (
	//Worker 工作单元
	Worker struct {
		speedsPerSecond int64 //速度
		wrange          Range
		speedsStat      speeds.Speeds
		id              int    //id
		cacheSize       int    //下载缓存
		url             string //下载地址
		referer         string //来源地址
		acceptRanges    string
		client          *requester.HTTPClient
		firstResp       *http.Response // 第一个响应
		writerAt        io.WriterAt
		writeMu         *sync.Mutex
		execMu          sync.Mutex

		paused                 bool
		pauseChan              chan struct{}
		workerCancelFunc       context.CancelFunc
		resetFunc              context.CancelFunc
		readRespBodyCancelFunc func()
		err                    error //错误信息
		status                 WorkerStatus
		downloadStatus         *DownloadStatus //总的下载状态
	}

	// WorkerList worker列表
	WorkerList []*Worker
)

// Duplicate 构造新的列表
func (wl WorkerList) Duplicate() WorkerList {
	n := make(WorkerList, len(wl))
	copy(n, wl)
	return n
}

//NewWorker 初始化Worker
func NewWorker(id int, durl string, writerAt io.WriterAt) *Worker {
	return &Worker{
		id:       id,
		url:      durl,
		writerAt: writerAt,
	}
}

//ID 返回worker ID
func (wer *Worker) ID() int {
	return wer.id
}

func (wer *Worker) lazyInit() {
	if wer.client == nil {
		wer.client = requester.NewHTTPClient()
	}
	if wer.writeMu == nil {
		wer.writeMu = &sync.Mutex{}
	}
	if wer.pauseChan == nil {
		wer.pauseChan = make(chan struct{})
	}
	if wer.wrange.LoadBegin() == 0 && wer.wrange.LoadEnd() == 0 {
		// 取消多线程下载
		wer.acceptRanges = ""
		wer.wrange.StoreEnd(-2)
	}
}

//SetClient 设置http客户端
func (wer *Worker) SetClient(c *requester.HTTPClient) {
	wer.client = c
}

//SetCacheSize 设置下载缓存
func (wer *Worker) SetCacheSize(size int) {
	wer.cacheSize = size
	fixCacheSize(&wer.cacheSize)
}

//SetRange 设置请求范围
func (wer *Worker) SetRange(acceptRanges string, r Range) {
	wer.acceptRanges = acceptRanges
	wer.wrange = r
}

//SetReferer 设置来源
func (wer *Worker) SetReferer(referer string) {
	wer.referer = referer
}

//SetWriteMutex 设置数据写锁
func (wer *Worker) SetWriteMutex(mu *sync.Mutex) {
	wer.writeMu = mu
}

//SetDownloadStatus 增加其他需要统计的数据
func (wer *Worker) SetDownloadStatus(downloadStatus *DownloadStatus) {
	wer.downloadStatus = downloadStatus
}

//GetStatus 返回下载状态
func (wer *Worker) GetStatus() Status {
	// 空接口与空指针不等价
	return &wer.status
}

//GetRange 返回worker范围
func (wer *Worker) GetRange() *Range {
	return &wer.wrange
}

//GetSpeedsPerSecond 获取每秒的速度
func (wer *Worker) GetSpeedsPerSecond() int64 {
	return atomic.LoadInt64(&wer.speedsPerSecond)
}

//Pause 暂停下载
func (wer *Worker) Pause() {
	wer.lazyInit()
	if wer.acceptRanges == "" {
		pcsverbose.Verbosef("WARNING: worker unsupport pause")
		return
	}

	if wer.paused {
		return
	}
	wer.pauseChan <- struct{}{}
	wer.paused = true
}

//Resume 恢复下载
func (wer *Worker) Resume() {
	wer.paused = false
	go wer.Execute()
}

//Cancel 取消下载
func (wer *Worker) Cancel() error {
	if wer.workerCancelFunc == nil {
		return errors.New("cancelFunc not set")
	}
	wer.workerCancelFunc()
	if wer.readRespBodyCancelFunc != nil {
		wer.readRespBodyCancelFunc()
	}
	return nil
}

//Reset 重设连接
func (wer *Worker) Reset() {
	if wer.resetFunc == nil {
		pcsverbose.Verbosef("DEBUG: worker: resetFunc not set")
		return
	}
	wer.resetFunc()
	if wer.readRespBodyCancelFunc != nil {
		wer.readRespBodyCancelFunc()
	}
	wer.CleanStatus()
	go wer.Execute()
}

// Canceled 是否已经取消
func (wer *Worker) Canceled() bool {
	return wer.status.statusCode == StatusCodeCanceled
}

//Completed 是否已经完成
func (wer *Worker) Completed() bool {
	switch wer.status.statusCode {
	case StatusCodeSuccessed, StatusCodeCanceled:
		return true
	default:
		return false
	}
}

//Failed 是否失败
func (wer *Worker) Failed() bool {
	switch wer.status.statusCode {
	case StatusCodeFailed, StatusCodeInternalError, StatusCodeTooManyConnections, StatusCodeNetError:
		return true
	default:
		return false
	}
}

//CleanStatus 清空状态
func (wer *Worker) CleanStatus() {
	wer.status.statusCode = StatusCodeInit
}

// updateSpeeds 更新速度
func (wer *Worker) updateSpeeds(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				atomic.StoreInt64(&wer.speedsPerSecond, 0)
				return
			default:
				atomic.StoreInt64(&wer.speedsPerSecond, wer.speedsStat.GetSpeedsPerSecond())
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

//Err 返回worker错误
func (wer *Worker) Err() error {
	return wer.err
}

//Execute 执行任务
func (wer *Worker) Execute() {
	wer.lazyInit()

	wer.execMu.Lock()
	defer wer.execMu.Unlock()

	single := wer.acceptRanges == ""

	// 如果已暂停, 退出
	if wer.paused {
		wer.status.statusCode = StatusCodePaused
		return
	}

	if !single {
		// 已完成
		if rlen := wer.wrange.Len(); rlen <= 0 {
			if rlen < 0 {
				pcsverbose.Verbosef("DEBUG: RangeLen is negative at begin: %v, %d\n", wer.wrange, wer.wrange.Len())
			}
			wer.status.statusCode = StatusCodeSuccessed
			return
		}
	}

	workerCancelCtx, workerCancelFunc := context.WithCancel(context.Background())
	wer.workerCancelFunc = workerCancelFunc
	resetCtx, resetFunc := context.WithCancel(context.Background())
	wer.resetFunc = resetFunc

	header := map[string]string{}
	if wer.referer != "" {
		header["Referer"] = wer.referer
	}
	//检测是否支持range
	if wer.acceptRanges != "" && wer.wrange.Len() >= 0 {
		header["Range"] = fmt.Sprintf("%s=%d-%d", wer.acceptRanges, wer.wrange.LoadBegin(), wer.wrange.LoadEnd())
	}

	wer.status.statusCode = StatusCodePending

	var resp *http.Response
	if wer.firstResp != nil {
		resp = wer.firstResp // 使用第一个连接
	} else {
		resp, wer.err = wer.client.Req("GET", wer.url, nil, header)
	}
	if resp != nil {
		defer func() {
			resp.Body.Close()
			wer.firstResp = nil // 去掉第一个连接
		}()
		wer.readRespBodyCancelFunc = func() {
			resp.Body.Close()
		}
	}
	if wer.err != nil {
		wer.status.statusCode = StatusCodeNetError
		return
	}

	var (
		contentLength = resp.ContentLength
		rangeLength   = wer.wrange.Len()
	)

	if !single {
		if contentLength != rangeLength && wer.firstResp == nil { // 跳过检查第一个连接
			wer.status.statusCode = StatusCodeNetError
			wer.err = fmt.Errorf("Content-Length is unexpected: %d, need %d", contentLength, rangeLength)
			return
		}
	}

	switch resp.StatusCode {
	case 200, 206:
		// do nothing, continue
	case 416: //Requested Range Not Satisfiable
		fallthrough
	case 403: // Forbidden
		fallthrough
	case 406: // Not Acceptable
		wer.status.statusCode = StatusCodeNetError
		wer.err = errors.New(resp.Status)
		return
	case 429, 509: // Too Many Requests
		wer.status.SetStatusCode(StatusCodeTooManyConnections)
		wer.err = errors.New(resp.Status)
		return
	default:
		wer.status.statusCode = StatusCodeNetError
		wer.err = fmt.Errorf("unexpected http status code, %d, %s", resp.StatusCode, resp.Status)
		return
	}

	fixCacheSize(&wer.cacheSize)
	var (
		speedsCtx, speedsCancelFunc = context.WithCancel(context.Background())
		cache                       = cachepool.Require(wer.cacheSize)
		buf                         = cache.Bytes()
		n, nn                       int
		n64, nn64                   int64
	)

	wer.updateSpeeds(speedsCtx)
	defer func() {
		speedsCancelFunc() // 结束速度统计
		cache.Free()
	}()

	for {
		select {
		case <-workerCancelCtx.Done(): //取消
			wer.status.statusCode = StatusCodeCanceled
			return
		case <-resetCtx.Done(): //重设连接
			wer.status.statusCode = StatusCodeReseted
			return
		case <-wer.pauseChan: //暂停
			wer.status.statusCode = StatusCodePaused
			return
		default:
			wer.status.statusCode = StatusCodeDownloading

			// 初始化数据
			var readErr error
			n = 0

			// 读取数据
			for n < len(buf) && readErr == nil && (single || wer.wrange.Len() > 0) {
				nn, readErr = resp.Body.Read(buf[n:])
				nn64 = int64(nn)

				// 更新速度统计
				if wer.downloadStatus != nil {
					wer.downloadStatus.AddSpeedsDownloaded(nn64)
				}
				wer.speedsStat.Add(nn64)
				n += nn
			}

			if n > 0 && readErr == io.EOF {
				readErr = io.ErrUnexpectedEOF
			}

			n64 = int64(n)

			// 非单线程模式下
			if !single {
				rangeLength = wer.wrange.Len()

				// 已完成 (未雨绸缪)
				if rangeLength <= 0 {
					wer.status.statusCode = StatusCodeCanceled
					wer.err = errors.New("worker already complete")
					return
				}

				if n64 > rangeLength {
					// 数据大小不正常
					n64 = rangeLength
					n = int(rangeLength)
					readErr = io.EOF
				}
			}

			// 写入数据
			if wer.writerAt != nil {
				wer.status.statusCode = StatusCodeWaitToWrite
				wer.writeMu.Lock()                                           // 加锁, 减轻硬盘的压力
				_, wer.err = wer.writerAt.WriteAt(buf[:n], wer.wrange.Begin) // 写入数据
				if wer.err != nil {
					wer.writeMu.Unlock()
					wer.status.statusCode = StatusCodeInternalError
					return
				}

				wer.writeMu.Unlock() //解锁
				wer.status.statusCode = StatusCodeDownloading
			}

			// 更新数据
			wer.wrange.AddBegin(n64)
			if wer.downloadStatus != nil {
				wer.downloadStatus.AddDownloaded(n64)
			}

			if readErr != nil {
				rlen := wer.wrange.Len()
				switch {
				case single && readErr == io.ErrUnexpectedEOF:
					// 单线程判断下载成功
					fallthrough
				case readErr == io.EOF:
					fallthrough
				case rlen <= 0:
					// 下载完成
					// 小于0可能是因为 worker 被 duplicate
					wer.status.statusCode = StatusCodeSuccessed
					if rlen < 0 {
						pcsverbose.Verbosef("DEBUG: RangeLen is negative at end: %v, %d\n", wer.wrange, wer.wrange.Len())
					}
					return
				default:
					// 其他错误, 返回
					wer.status.statusCode = StatusCodeFailed
					wer.err = readErr
					return
				}
			}
		}
	}
}
