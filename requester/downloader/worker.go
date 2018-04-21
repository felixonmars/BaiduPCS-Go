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

//Worker 工作单元
type Worker struct {
	speedsPerSecond int64 //速度
	wrange          *Range
	speedsStat      speeds.Speeds
	id              int32  //id
	cacheSize       int    //下载缓存
	url             string //下载地址
	referer         string //来源地址
	acceptRanges    string
	client          *requester.HTTPClient
	resp            *http.Response
	writerAt        io.WriterAt
	writeMu         *sync.Mutex
	execMu          sync.Mutex

	resetable  bool
	inited     bool
	paused     bool
	pauseChan  chan struct{}
	cancelFunc context.CancelFunc
	resetFunc  context.CancelFunc
	err        error //错误信息
	status     *WorkerStatus
	othersAdd  []*int64
}

//NewWorker 初始化Worker
func NewWorker(id int32, durl string, writerAt io.WriterAt) *Worker {
	return &Worker{
		id:       id,
		url:      durl,
		writerAt: writerAt,
	}
}

//MustCheck 遇到错误则panic
func (wer *Worker) MustCheck() {
	if wer.client == nil {
		panic("client is nil")
	}
	if wer.status == nil {
		panic("status is nil")
	}
}

//InitedChan 是否已经完全的初始化, 是则发送chan
func (wer *Worker) InitedChan() <-chan struct{} {
	c := make(chan struct{})
	go func() {
		for {
			time.Sleep(1e8)
			if wer.inited {
				c <- struct{}{}
				return
			}
		}
	}()
	return c
}

func (wer *Worker) lazyInit() {
	if wer.client == nil {
		wer.client = requester.NewHTTPClient()
	}
	if wer.status == nil {
		wer.status = NewWorkerStatus()
	}
	if wer.writeMu == nil {
		wer.writeMu = &sync.Mutex{}
	}
	wer.inited = true
}

//SetClient 设置http客户端
func (wer *Worker) SetClient(c *requester.HTTPClient) {
	if c != nil {
		wer.client = c
	}
}

//SetCacheSize 设置下载缓存
func (wer *Worker) SetCacheSize(size int) {
	wer.cacheSize = size
	fixCacheSize(&wer.cacheSize)
}

//SetRange 设置请求范围
func (wer *Worker) SetRange(acceptRanges string, r *Range) {
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

//AppendOthersAdd 增加其他需要统计的数据
func (wer *Worker) AppendOthersAdd(ptrI *int64) {
	wer.othersAdd = append(wer.othersAdd, ptrI)
}

//GetStatus 返回下载状态
func (wer *Worker) GetStatus() Status {
	if wer.status == nil {
		wer.status = NewWorkerStatus()
	}
	return wer.status
}

//GetSpeedsPerSecond 获取每秒的速度
func (wer *Worker) GetSpeedsPerSecond() int64 {
	return atomic.LoadInt64(&wer.speedsPerSecond)
}

//Pause 暂停下载
func (wer *Worker) Pause() {
	if wer.acceptRanges == "" {
		pcsverbose.Verbosef("WARNING: worker unsupport pause")
		return
	}

	if wer.paused {
		return
	}
	wer.pauseChan = make(chan struct{}, 1)
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
	if wer.cancelFunc == nil {
		return errors.New("cancelFunc not set")
	}
	wer.cancelFunc()
	if wer.resp != nil {
		wer.resp.Body.Close()
	}
	return nil
}

//Resetable 是否可以重设
func (wer *Worker) Resetable() bool {
	return wer.resetable
}

//Reset 重设连接
func (wer *Worker) Reset() {
	if !wer.resetable {
		return
	}

	if wer.resetFunc == nil {
		pcsverbose.Verbosef("DEBUG: worker: resetFunc not set")
		return
	}
	wer.resetFunc()
	if wer.resp != nil {
		wer.resp.Body.Close()
	}
	wer.CleanStatus()
	go wer.Execute()
}

// Canceled 是否已经取消
func (wer *Worker) Canceled() bool {
	return wer.GetStatus().StatusCode() == StatusCodeCanceled
}

//Completed 是否已经完成
func (wer *Worker) Completed() bool {
	switch wer.GetStatus().StatusCode() {
	case StatusCodeSuccessed, StatusCodeCanceled:
		return true
	default:
		return false
	}
}

//Failed 是否失败
func (wer *Worker) Failed() bool {
	switch wer.GetStatus().StatusCode() {
	case StatusCodeFailed, StatusCodeInternalError, StatusCodeTooManyConnections, StatusCodeNetError:
		return true
	default:
		return false
	}
}

//CleanStatus 清空状态
func (wer *Worker) CleanStatus() {
	wer.status = NewWorkerStatus()
	wer.resetable = false
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

//Execute 执行任务
func (wer *Worker) Execute() {
	wer.lazyInit()

	wer.execMu.Lock()
	defer wer.execMu.Unlock()

	wer.err = nil
	single := wer.acceptRanges == ""

	go func() {
		time.Sleep(5e9)
		wer.resetable = true
	}()

	// 如果已暂停, 退出
	if wer.paused {
		wer.status.SetStatusCode(StatusCodePaused)
		return
	}

	// 已完成
	if wer.wrange.Len() == 0 {
		wer.status.SetStatusCode(StatusCodeSuccessed)
		return
	}

	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	wer.cancelFunc = cancelFunc
	resetCtx, resetFunc := context.WithCancel(context.Background())
	wer.resetFunc = resetFunc

	setNetError := func() {
		wer.status.SetStatusCode(StatusCodeNetError)
	}

	header := map[string]string{}
	if wer.referer != "" {
		header["Referer"] = wer.referer
	}
	//检测是否支持range
	if wer.acceptRanges != "" && wer.wrange != nil {
		header["Range"] = fmt.Sprintf("%s=%d-%d", wer.acceptRanges, wer.wrange.LoadBegin(), wer.wrange.LoadEnd())
	}

	wer.status.SetStatusCode(StatusCodePending)

	var resp *http.Response

	resp, wer.err = wer.client.Req("GET", wer.url, nil, header)
	wer.resetable = true
	if resp != nil {
		defer resp.Body.Close()
	}
	if wer.err != nil {
		setNetError()
		return
	}

	wer.resp = resp

	var (
		contentLength = resp.ContentLength
		rangeLength   = wer.wrange.Len()
	)

	if !single {
		if contentLength != rangeLength {
			setNetError()
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
		setNetError()
		wer.err = errors.New(resp.Status)
		return
	case 429, 509: // Too Many Requests
		wer.status.SetStatusCode(StatusCodeTooManyConnections)
		wer.err = errors.New(resp.Status)
		return
	default:
		setNetError()
		wer.err = fmt.Errorf("unexpected http status code, %d, %s", resp.StatusCode, resp.Status)
		return
	}

	fixCacheSize(&wer.cacheSize)
	var (
		speedsCtx, speedsCancel = context.WithCancel(context.Background())
		buf                     = cachepool.SetIfNotExist(wer.id, wer.cacheSize)
		n                       int
		n64                     int64
		readErr                 error
	)
	wer.updateSpeeds(speedsCtx)
	defer speedsCancel()

	for {
		select {
		case <-cancelCtx.Done():
			wer.status.SetStatusCode(StatusCodeCanceled)
			return
		case <-resetCtx.Done():
			wer.status.SetStatusCode(StatusCodeReseted)
			return
		case <-wer.pauseChan:
			wer.status.SetStatusCode(StatusCodePaused)
			return
		default:
			wer.status.SetStatusCode(StatusCodeDownloading)
			n, readErr = readFullFrom(resp.Body, buf, &wer.speedsStat)
			n64 = int64(n)

			// 非单线程模式下
			if !single {
				rangeLength = wer.wrange.Len()

				// 已完成 (未雨绸缪)
				if rangeLength <= 0 {
					wer.status.SetStatusCode(StatusCodeCanceled)
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
				wer.status.SetStatusCode(StatusCodeWaitToWrite)
				wer.writeMu.Lock()                                           // 加锁, 减轻硬盘的压力
				_, wer.err = wer.writerAt.WriteAt(buf[:n], wer.wrange.Begin) // 写入数据
				if wer.err != nil {
					wer.writeMu.Unlock()
					wer.status.SetStatusCode(StatusCodeInternalError)
					return
				}

				wer.writeMu.Unlock() //解锁
				wer.status.SetStatusCode(StatusCodeDownloading)
			}

			// 更新数据
			wer.wrange.AddBegin(n64)
			for k := range wer.othersAdd {
				if wer.othersAdd[k] == nil {
					continue
				}
				atomic.AddInt64(wer.othersAdd[k], n64)
			}

			if readErr != nil {
				switch {
				case single && readErr == io.ErrUnexpectedEOF:
					// 单线程判断下载成功
					fallthrough
				case readErr == io.EOF:
					fallthrough
				case wer.wrange.Len() == 0:
					// 下载完成
					wer.status.SetStatusCode(StatusCodeSuccessed)
					return
				default:
					// 其他错误, 返回
					wer.status.SetStatusCode(StatusCodeFailed)
					wer.err = readErr
					return
				}
			}
		}
	}
}
