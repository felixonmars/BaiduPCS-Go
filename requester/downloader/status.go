package downloader

import (
	"sync"
	"sync/atomic"
	"time"
)

//Status 状态
type Status interface {
	StatusCode() StatusCode //状态码
	StatusText() string
}

//StatusCode 状态码
type StatusCode int

//WorkerStatus worker状态
type WorkerStatus struct {
	statusCode StatusCode
}

//DlStatus 下载状态接口
type DlStatus interface {
	TotalSize() int64
	Downloaded() int64
	SpeedsPerSecond() int64
	TimeElapsed() time.Duration
}

//DownloadStatus 下载状态及统计信息
type DownloadStatus struct {
	totalSize       int64 // 总大小
	downloaded      int64 // 已下载的数据量
	speedsPerSecond int64 // 下载速度
	maxSpeeds       int64 // 最大下载速度

	speedsDownloaded int64 // 用于统计数据的downloaded
	oldDownloaded    int64
	timeElapsed      time.Duration // 下载的时间
	nowTime          time.Time
	sinceNowTime     time.Duration
	mu               sync.Mutex
}

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
		nowTime: time.Now(),
	}
}

//Add 实现Adder接口
func (ds *DownloadStatus) Add(i int64) {
	ds.AddDownloaded(i)
}

//AddDownloaded 增加已下载数据量
func (ds *DownloadStatus) AddDownloaded(d int64) {
	atomic.AddInt64(&ds.downloaded, d)
}

//AddSpeedsDownloaded 增加已下载数据量, 用于统计速度
func (ds *DownloadStatus) AddSpeedsDownloaded(d int64) {
	atomic.AddInt64(&ds.speedsDownloaded, d)
	ds.updateSpeeds()
}

func (ds *DownloadStatus) updateSpeeds() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.nowTime.Unix() <= 0 {
		ds.nowTime = time.Now()
	}

	ds.sinceNowTime = time.Since(ds.nowTime)
	seconds := ds.sinceNowTime.Seconds()
	if seconds < 0.5 {
		return
	}

	downloaded := ds.SpeedsDownloaded() - ds.oldDownloaded
	speeds := int64(float64(downloaded) / seconds)
	ds.StoreSpeedsPerSecond(speeds)
	if speeds > ds.MaxSpeeds() {
		ds.StoreMaxSpeeds(ds.speedsPerSecond)
	}

	ds.nowTime = time.Now()
	ds.oldDownloaded = ds.SpeedsDownloaded()
}

//ResetMaxSpeeds 清空最大速度统计
func (ds *DownloadStatus) ResetMaxSpeeds() {
	ds.StoreMaxSpeeds(0)
}

//StoreMaxSpeeds 储存最大速度, 原子操作
func (ds *DownloadStatus) StoreMaxSpeeds(speeds int64) {
	atomic.StoreInt64(&ds.maxSpeeds, speeds)
}

//StoreSpeedsPerSecond 储存速度, 原子操作
func (ds *DownloadStatus) StoreSpeedsPerSecond(speeds int64) {
	atomic.StoreInt64(&ds.speedsPerSecond, speeds)
}

//TotalSize 返回总大小
func (ds *DownloadStatus) TotalSize() int64 {
	return atomic.LoadInt64(&ds.totalSize)
}

//Downloaded 返回已下载数据量
func (ds *DownloadStatus) Downloaded() int64 {
	return atomic.LoadInt64(&ds.downloaded)
}

//SpeedsDownloaded 返回用于统计速度的已下载数据量
func (ds *DownloadStatus) SpeedsDownloaded() int64 {
	return atomic.LoadInt64(&ds.speedsDownloaded)
}

//SpeedsPerSecond 返回每秒速度
func (ds *DownloadStatus) SpeedsPerSecond() int64 {
	return atomic.LoadInt64(&ds.speedsPerSecond)
}

//MaxSpeeds 返回最大速度
func (ds *DownloadStatus) MaxSpeeds() int64 {
	return atomic.LoadInt64(&ds.maxSpeeds)
}

//TimeElapsed 返回花费的时间
func (ds *DownloadStatus) TimeElapsed() time.Duration {
	return ds.timeElapsed
}
