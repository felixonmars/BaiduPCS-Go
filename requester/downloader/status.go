package downloader

import (
	"github.com/iikira/BaiduPCS-Go/requester/rio/speeds"
	"sync"
	"sync/atomic"
	"time"
)

type (
	//WorkerStatuser 状态
	WorkerStatuser interface {
		StatusCode() StatusCode //状态码
		StatusText() string
	}

	//StatusCode 状态码
	StatusCode int

	//WorkerStatus worker状态
	WorkerStatus struct {
		statusCode StatusCode
	}

	//DownloadStatuser 下载状态接口
	DownloadStatuser interface {
		TotalSize() int64
		Downloaded() int64
		SpeedsPerSecond() int64
		TimeElapsed() time.Duration // 已开始时间
		TimeLeft() time.Duration    // 预计剩余时间, 负数代表未知
	}

	//DownloadStatus 下载状态及统计信息
	DownloadStatus struct {
		totalSize        int64 // 总大小
		downloaded       int64 // 已下载的数据量
		speedsDownloaded int64 // 用于统计速度的downloaded
		maxSpeeds        int64 // 最大下载速度
		tmpSpeeds        int64 // 缓存的速度

		startTime  time.Time // 开始下载的时间
		speedsStat speeds.Speeds

		rateLimit *speeds.RateLimit // 限速控制

		gen *RangeListGen // Range生成状态
		mu  sync.Mutex
	}

	// DownloadStatusFunc 下载状态处理函数
	DownloadStatusFunc func(status DownloadStatuser, workersCallback func(RangeWorkerFunc))
)

const (
	//StatusCodeInit 初始化
	StatusCodeInit StatusCode = iota
	//StatusCodeSuccessed 成功
	StatusCodeSuccessed
	//StatusCodePending 等待响应
	StatusCodePending
	//StatusCodeDownloading 下载中
	StatusCodeDownloading
	//StatusCodeWaitToWrite 等待写入数据
	StatusCodeWaitToWrite
	//StatusCodeInternalError 内部错误
	StatusCodeInternalError
	//StatusCodeTooManyConnections 连接数太多
	StatusCodeTooManyConnections
	//StatusCodeNetError 网络错误
	StatusCodeNetError
	//StatusCodeFailed 下载失败
	StatusCodeFailed
	//StatusCodePaused 已暂停
	StatusCodePaused
	//StatusCodeReseted 已重设连接
	StatusCodeReseted
	//StatusCodeCanceled 已取消
	StatusCodeCanceled
)

//GetStatusText 根据状态码获取状态信息
func GetStatusText(sc StatusCode) string {
	switch sc {
	case StatusCodeInit:
		return "初始化"
	case StatusCodeSuccessed:
		return "成功"
	case StatusCodePending:
		return "等待响应"
	case StatusCodeDownloading:
		return "下载中"
	case StatusCodeWaitToWrite:
		return "等待写入数据"
	case StatusCodeInternalError:
		return "内部错误"
	case StatusCodeTooManyConnections:
		return "连接数太多"
	case StatusCodeNetError:
		return "网络错误"
	case StatusCodeFailed:
		return "下载失败"
	case StatusCodePaused:
		return "已暂停"
	case StatusCodeReseted:
		return "已重设连接"
	case StatusCodeCanceled:
		return "已取消"
	default:
		return "未知状态码"
	}
}

//NewWorkerStatus 初始化WorkerStatus
func NewWorkerStatus() *WorkerStatus {
	return &WorkerStatus{
		statusCode: StatusCodeInit,
	}
}

//SetStatusCode 设置worker状态码
func (ws *WorkerStatus) SetStatusCode(sc StatusCode) {
	ws.statusCode = sc
}

//StatusCode 返回状态码
func (ws *WorkerStatus) StatusCode() StatusCode {
	return ws.statusCode
}

//StatusText 返回状态信息
func (ws *WorkerStatus) StatusText() string {
	return GetStatusText(ws.statusCode)
}

//NewDownloadStatus 初始化DownloadStatus
func NewDownloadStatus() *DownloadStatus {
	return &DownloadStatus{
		startTime: time.Now(),
	}
}

// SetRateLimit 设置限速
func (ds *DownloadStatus) SetRateLimit(rl *speeds.RateLimit) {
	ds.rateLimit = rl
}

//AddDownloaded 增加已下载数据量
func (ds *DownloadStatus) AddDownloaded(d int64) {
	atomic.AddInt64(&ds.downloaded, d)
}

//AddSpeedsDownloaded 增加已下载数据量, 用于统计速度
func (ds *DownloadStatus) AddSpeedsDownloaded(d int64) {
	if ds.rateLimit != nil {
		ds.rateLimit.Add(d)
	}
	ds.speedsStat.Add(d)
}

//StoreMaxSpeeds 储存最大速度, 原子操作
func (ds *DownloadStatus) StoreMaxSpeeds(speeds int64) {
	atomic.StoreInt64(&ds.maxSpeeds, speeds)
}

//TotalSize 返回总大小
func (ds *DownloadStatus) TotalSize() int64 {
	return atomic.LoadInt64(&ds.totalSize)
}

//Downloaded 返回已下载数据量
func (ds *DownloadStatus) Downloaded() int64 {
	return atomic.LoadInt64(&ds.downloaded)
}

// UpdateSpeeds 更新speeds
func (ds *DownloadStatus) UpdateSpeeds() {
	atomic.StoreInt64(&ds.tmpSpeeds, ds.speedsStat.GetSpeeds())
}

//SpeedsPerSecond 返回每秒速度
func (ds *DownloadStatus) SpeedsPerSecond() int64 {
	return atomic.LoadInt64(&ds.tmpSpeeds)
}

//MaxSpeeds 返回最大速度
func (ds *DownloadStatus) MaxSpeeds() int64 {
	return atomic.LoadInt64(&ds.maxSpeeds)
}

//TimeElapsed 返回花费的时间
func (ds *DownloadStatus) TimeElapsed() (elapsed time.Duration) {
	return time.Since(ds.startTime)
}

//TimeLeft 返回预计剩余时间
func (ds *DownloadStatus) TimeLeft() (left time.Duration) {
	speeds := atomic.LoadInt64(&ds.tmpSpeeds)
	if speeds <= 0 {
		left = -1
	} else {
		left = time.Duration((ds.totalSize-ds.downloaded)/(speeds)) * time.Second
	}
	return
}

// RangeListGen 返回RangeListGen
func (ds *DownloadStatus) RangeListGen() *RangeListGen {
	return ds.gen
}
