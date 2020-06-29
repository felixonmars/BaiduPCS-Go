package pcsdownload

import (
	"errors"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/internal/pcsfunctions"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"github.com/iikira/BaiduPCS-Go/pcsutil/taskframework"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/BaiduPCS-Go/requester/downloader"
	"github.com/iikira/BaiduPCS-Go/requester/transfer"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type (
	// DownloadMode 下载模式
	DownloadMode int

	// DownloadTaskUnit 下载的任务单元
	DownloadTaskUnit struct {
		taskInfo *taskframework.TaskInfo // 任务信息

		Cfg                *downloader.Config
		PCS                *baidupcs.BaiduPCS
		ParentTaskExecutor *taskframework.TaskExecutor

		DownloadStatistic *DownloadStatistic // 下载统计

		// 可选项
		VerbosePrinter       *pcsverbose.PCSVerbose
		PrintFormat          string
		IsPrintStatus        bool // 是否输出各个下载线程的详细信息
		IsExecutedPermission bool // 下载成功后是否加上执行权限
		IsOverwrite          bool // 是否覆盖已存在的文件
		NoCheck              bool // 不校验文件

		DownloadMode DownloadMode // 下载模式

		PcsPath  string // 要下载的网盘文件路径
		SavePath string // 保存的路径

		fileInfo *baidupcs.FileDirectory // 文件或目录详情
	}
)

const (
	// DefaultPrintFormat 默认的下载进度输出格式
	DefaultPrintFormat = "\r[%s] ↓ %s/%s %s/s in %s, left %s ............"
	//DownloadSuffix 文件下载后缀
	DownloadSuffix = ".BaiduPCS-Go-downloading"
	//StrDownloadInitError 初始化下载发生错误
	StrDownloadInitError = "初始化下载发生错误"
	// StrDownloadFailed 下载文件失败
	StrDownloadFailed = "下载文件失败"
	// StrDownloadGetDlinkFailed 获取下载链接失败
	StrDownloadGetDlinkFailed = "获取下载链接失败"
	// StrDownloadChecksumFailed 检测文件有效性失败
	StrDownloadChecksumFailed = "检测文件有效性失败"
	// DefaultDownloadMaxRetry 默认下载失败最大重试次数
	DefaultDownloadMaxRetry = 3
)

const (
	DownloadModeLocate DownloadMode = iota
	DownloadModePCS
	DownloadModeStreaming
)

func (dtu *DownloadTaskUnit) SetTaskInfo(info *taskframework.TaskInfo) {
	dtu.taskInfo = info
}

func (dtu *DownloadTaskUnit) verboseInfof(format string, a ...interface{}) {
	if dtu.VerbosePrinter != nil {
		dtu.VerbosePrinter.Infof(format, a...)
	}
}

// download 执行下载
func (dtu *DownloadTaskUnit) download(downloadURL string, client *requester.HTTPClient) (err error) {
	var (
		writer downloader.Writer
		file   *os.File
	)

	if !dtu.Cfg.IsTest {
		// 非测试下载
		dtu.Cfg.InstanceStatePath = dtu.SavePath + DownloadSuffix

		// 创建下载的目录
		// 获取SavePath所在的目录
		dir := filepath.Dir(dtu.SavePath)
		fileInfo, err := os.Stat(dir)
		if err != nil {
			// 目录不存在, 创建
			err = os.MkdirAll(dir, 0777)
			if err != nil {
				return err
			}
		} else if !fileInfo.IsDir() {
			// SavePath所在的目录不是目录
			return fmt.Errorf("%s, path %s: not a directory", StrDownloadInitError, dir)
		}

		// 打开文件
		writer, file, err = downloader.NewDownloaderWriterByFilename(dtu.SavePath, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return fmt.Errorf("%s, %s", StrDownloadInitError, err)
		}
		defer file.Close()
	}

	der := downloader.NewDownloader(downloadURL, writer, dtu.Cfg)
	der.SetClient(client)
	der.SetDURLCheckFunc(BaiduPCSURLCheckFunc)
	der.SetStatusCodeBodyCheckFunc(func(respBody io.Reader) error {
		// 返回的错误可能是pcs的json
		// 解析错误
		return pcserror.DecodePCSJSONError(baidupcs.OperationDownloadFile, respBody)
	})

	// 检查输出格式
	if dtu.PrintFormat == "" {
		dtu.PrintFormat = DefaultPrintFormat
	}

	// 这里用共享变量的方式
	isComplete := false
	der.OnDownloadStatusEvent(func(status transfer.DownloadStatuser, workersCallback func(downloader.RangeWorkerFunc)) {
		// 这里可能会下载结束了, 还会输出内容
		builder := &strings.Builder{}
		if dtu.IsPrintStatus {
			// 输出所有的worker状态
			var (
				tb      = pcstable.NewTable(builder)
			)
			tb.SetHeader([]string{"#", "status", "range", "left", "speeds", "error"})
			workersCallback(func(key int, worker *downloader.Worker) bool {
				wrange := worker.GetRange()
				tb.Append([]string{fmt.Sprint(worker.ID()), worker.GetStatus().StatusText(), wrange.ShowDetails(), strconv.FormatInt(wrange.Len(), 10), strconv.FormatInt(worker.GetSpeedsPerSecond(), 10), fmt.Sprint(worker.Err())})
				return true
			})

			// 先空两行
			builder.WriteString("\n\n")
			tb.Render()
		}

		// 如果下载速度为0, 剩余下载时间未知, 则用 - 代替
		var leftStr string
		left := status.TimeLeft()
		if left < 0 {
			leftStr = "-"
		} else {
			leftStr = left.String()
		}

		fmt.Fprintf(builder,dtu.PrintFormat, dtu.taskInfo.Id(),
			converter.ConvertFileSize(status.Downloaded(), 2),
			converter.ConvertFileSize(status.TotalSize(), 2),
			converter.ConvertFileSize(status.SpeedsPerSecond(), 2),
			status.TimeElapsed()/1e7*1e7, leftStr,
		)

		if !isComplete {
			// 如果未完成下载, 就输出
			fmt.Print(builder.String())
		}
	})

	der.OnExecute(func() {
		if dtu.Cfg.IsTest {
			fmt.Printf("[%s] 测试下载开始\n\n", dtu.taskInfo.Id())
		}
	})

	err = der.Execute()
	isComplete = true
	fmt.Print("\n")

	if err != nil {
		// 下载发生错误
		if !dtu.Cfg.IsTest {
			// 下载失败, 删去空文件
			if info, infoErr := file.Stat(); infoErr == nil {
				if info.Size() == 0 {
					// 空文件, 应该删除
					dtu.verboseInfof("[%s] remove empty file: %s\n", dtu.taskInfo.Id(), dtu.SavePath)
					removeErr := os.Remove(dtu.SavePath)
					if removeErr != nil {
						dtu.verboseInfof("[%s] remove file error: %s\n", dtu.taskInfo.Id(), removeErr)
					}
				}
			}
		}
		return err
	}

	// 下载成功
	if !dtu.Cfg.IsTest {
		if dtu.IsExecutedPermission {
			err = file.Chmod(0766)
			if err != nil {
				fmt.Printf("[%s] 警告, 加执行权限错误: %s\n", dtu.taskInfo.Id(), err)
			}
		}

		fmt.Printf("[%s] 下载完成, 保存位置: %s\n", dtu.taskInfo.Id(), dtu.SavePath)
	} else {
		fmt.Printf("[%s] 测试下载结束\n", dtu.taskInfo.Id())
	}

	return nil
}

//panHTTPClient 获取包含特定User-Agent的HTTPClient
func (dtu *DownloadTaskUnit) panHTTPClient() (client *requester.HTTPClient) {
	client = pcsconfig.Config.PanHTTPClient()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		// 去掉 Referer
		if !pcsconfig.Config.EnableHTTPS {
			req.Header.Del("Referer")
		}
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}
	client.SetTimeout(20 * time.Minute)
	client.SetKeepAlive(true)
	return client
}

func (dtu *DownloadTaskUnit) handleError(result *taskframework.TaskUnitRunResult) {
	switch value := result.Err.(type) {
	case pcserror.Error: // pcserror 接口
		switch value.GetErrType() {
		case pcserror.ErrTypeRemoteError:
		// 远程服务器错误
		case 31045: // user not exists
			fallthrough
		case 31066: // file does not exist
			result.NeedRetry = false
		case 31626: // user is not authorized
			//可能是User-Agent不对
			//重试
			fallthrough
		default:
			result.NeedRetry = true
		}
	case *os.PathError:
		// 系统级别的错误, 可能是权限问题
		result.NeedRetry = false
	default:
		// 其他错误, 需要重试
		result.NeedRetry = true
	}
}

func (dtu *DownloadTaskUnit) execPanDownload(dlink string, result *taskframework.TaskUnitRunResult, okPtr *bool) {
	dtu.verboseInfof("[%s] 获取到下载链接: %s\n", dtu.taskInfo.Id(), dlink)

	client := dtu.panHTTPClient()
	err := dtu.download(dlink, client)
	if err != nil {
		result.ResultMessage = StrDownloadFailed
		result.Err = err
		dtu.handleError(result)
		return
	}
	*okPtr = true
}

func (dtu *DownloadTaskUnit) locateDownload(result *taskframework.TaskUnitRunResult) (ok bool) {
	rawDlinks, err := GetLocateDownloadLinks(dtu.PCS, dtu.PcsPath)
	if err != nil {
		result.ResultMessage = StrDownloadGetDlinkFailed
		result.Err = err
		dtu.handleError(result)
		return
	}

	// 更新链接的协议
	FixHTTPLinkURL(rawDlinks[0])
	// 至少重试一次 HTTP 下载
	if dtu.taskInfo.Retry() == 2 {
		rawDlinks[0].Scheme = "http"
	}
	dlink := rawDlinks[0].String()

	dtu.execPanDownload(dlink, result, &ok)
	return
}

func (dtu *DownloadTaskUnit) pcsOrStreamingDownload(mode DownloadMode, result *taskframework.TaskUnitRunResult) (ok bool) {
	dfunc := func(downloadURL string, jar http.CookieJar) error {
		client := pcsconfig.Config.PCSHTTPClient()
		client.SetCookiejar(jar)
		client.SetKeepAlive(true)
		client.SetTimeout(10 * time.Minute)

		return dtu.download(downloadURL, client)
	}

	var err error
	switch mode {
	case DownloadModePCS:
		err = dtu.PCS.DownloadFile(dtu.PcsPath, dfunc)
	case DownloadModeStreaming:
		err = dtu.PCS.DownloadStreamFile(dtu.PcsPath, dfunc)
	default:
		panic("unreachable")
	}

	if err != nil {
		result.ResultMessage = StrDownloadFailed
		result.Err = err
		dtu.handleError(result)
		return
	}
	return true // 下载成功
}

//checkFileValid 检测文件有效性
func (dtu *DownloadTaskUnit) checkFileValid(result *taskframework.TaskUnitRunResult) (ok bool) {
	if dtu.Cfg.IsTest || dtu.NoCheck {
		// 不检测文件有效性
		return
	}

	if dtu.fileInfo.Size >= 128*converter.MB {
		// 大文件, 输出一句提示消息
		fmt.Printf("[%s] 开始检验文件有效性, 请稍候...\n", dtu.taskInfo.Id())
	}

	// 就在这里处理校验出错
	err := CheckFileValid(dtu.SavePath, dtu.fileInfo)
	if err != nil {
		result.ResultMessage = StrDownloadChecksumFailed
		result.Err = err
		switch err {
		case ErrDownloadNotSupportChecksum:
			// 文件不支持校验
			result.ResultMessage = "检验文件有效性"
			result.Err = err
			fmt.Printf("[%s] 检验文件有效性: %s\n", dtu.taskInfo.Id(), err)
			return true
		case ErrDownloadFileBanned:
			// 违规文件
			result.NeedRetry = false
			return
		case ErrDownloadChecksumFailed:
			// 校验失败, 需要重新下载
			result.NeedRetry = true
			// 设置允许覆盖
			dtu.IsOverwrite = true
			return
		default:
			result.NeedRetry = false
			return
		}
	}

	fmt.Printf("[%s] 检验文件有效性成功: %s\n", dtu.taskInfo.Id(), dtu.SavePath)
	return true
}

func (dtu *DownloadTaskUnit) OnRetry(lastRunResult *taskframework.TaskUnitRunResult) {
	// 输出错误信息
	if lastRunResult.Err == nil {
		// result中不包含Err, 忽略输出
		fmt.Printf("[%s] %s, 重试 %d/%d\n", dtu.taskInfo.Id(), lastRunResult.ResultMessage, dtu.taskInfo.Retry(), dtu.taskInfo.MaxRetry())
		return
	}
	fmt.Printf("[%s] %s, %s, 重试 %d/%d\n", dtu.taskInfo.Id(), lastRunResult.ResultMessage, lastRunResult.Err, dtu.taskInfo.Retry(), dtu.taskInfo.MaxRetry())
}

func (dtu *DownloadTaskUnit) OnSuccess(lastRunResult *taskframework.TaskUnitRunResult) {
}

func (dtu *DownloadTaskUnit) OnFailed(lastRunResult *taskframework.TaskUnitRunResult) {
	// 失败
	if lastRunResult.Err == nil {
		// result中不包含Err, 忽略输出
		fmt.Printf("[%s] %s\n", dtu.taskInfo.Id(), lastRunResult.ResultMessage)
		return
	}
	fmt.Printf("[%s] %s, %s\n", dtu.taskInfo.Id(), lastRunResult.ResultMessage, lastRunResult.Err)
}

func (dtu *DownloadTaskUnit) OnComplete(lastRunResult *taskframework.TaskUnitRunResult) {
}

func (dtu *DownloadTaskUnit) RetryWait() time.Duration {
	return pcsfunctions.RetryWait(dtu.taskInfo.Retry())
}

func (dtu *DownloadTaskUnit) Run() (result *taskframework.TaskUnitRunResult) {
	result = &taskframework.TaskUnitRunResult{}
	// 获取文件信息
	var err error
	if dtu.fileInfo == nil || dtu.taskInfo.Retry() > 0 {
		// 没有获取文件信息
		// 如果是动态添加的下载任务, 是会写入文件信息的
		// 如果该任务重试过, 则应该再获取一次文件信息
		dtu.fileInfo, err = dtu.PCS.FilesDirectoriesMeta(dtu.PcsPath)
		if err != nil {
			// 如果不是未登录或文件不存在, 则不重试
			result.ResultMessage = "获取下载路径信息错误"
			result.Err = err
			dtu.handleError(result)
			return
		}
	}

	// 输出文件信息
	fmt.Print("\n")
	fmt.Printf("[%s] ----\n%s\n", dtu.taskInfo.Id(), dtu.fileInfo.String())

	// 如果是一个目录, 将子文件和子目录加入队列
	if dtu.fileInfo.Isdir {
		if !dtu.Cfg.IsTest { // 测试下载, 不建立空目录
			os.MkdirAll(dtu.SavePath, 0777) // 首先在本地创建目录, 保证空目录也能被保存
		}

		// 获取该目录下的文件列表
		fileList, err := dtu.PCS.FilesDirectoriesList(dtu.PcsPath, baidupcs.DefaultOrderOptions)
		if err != nil {
			result.ResultMessage = "获取目录信息错误"
			result.Err = err
			result.NeedRetry = true
			return
		}

		for k := range fileList {
			// 添加子任务
			subUnit := *dtu
			newCfg := *dtu.Cfg
			subUnit.Cfg = &newCfg
			subUnit.fileInfo = fileList[k] // 保存文件信息
			subUnit.PcsPath = fileList[k].Path
			subUnit.SavePath = filepath.Join(dtu.SavePath, fileList[k].Filename) // 保存位置

			// 加入父队列
			info := dtu.ParentTaskExecutor.Append(&subUnit, dtu.taskInfo.MaxRetry())
			fmt.Printf("[%s] 加入下载队列: %s\n", info.Id(), fileList[k].Path)
		}

		result.Succeed = true // 执行成功
		return
	}

	fmt.Printf("[%s] 准备下载: %s\n", dtu.taskInfo.Id(), dtu.PcsPath)

	if !dtu.Cfg.IsTest && !dtu.IsOverwrite && FileExist(dtu.SavePath) {
		fmt.Printf("[%s] 文件已经存在: %s, 跳过...\n", dtu.taskInfo.Id(), dtu.SavePath)
		result.Succeed = true // 执行成功
		return
	}

	if !dtu.Cfg.IsTest {
		// 不是测试下载, 输出下载路径
		fmt.Printf("[%s] 将会下载到路径: %s\n\n", dtu.taskInfo.Id(), dtu.SavePath)
	}

	var ok bool
	// 获取下载链接
	switch dtu.DownloadMode {
	case DownloadModeLocate:
		ok = dtu.locateDownload(result)
	case DownloadModePCS, DownloadModeStreaming:
		ok = dtu.pcsOrStreamingDownload(dtu.DownloadMode, result)
	}

	if !ok {
		// 以上执行不成功, 返回
		return result
	}

	// 检测文件有效性
	ok = dtu.checkFileValid(result)
	if !ok {
		// 校验不成功, 返回结果
		return result
	}

	// 统计下载
	dtu.DownloadStatistic.AddTotalSize(dtu.fileInfo.Size)
	// 下载成功
	result.Succeed = true
	return
}
