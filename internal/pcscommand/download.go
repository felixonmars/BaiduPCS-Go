package pcscommand

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/baidupcs/dlinkclient"
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/internal/pcsfunctions/pcsdownload"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/iikira/BaiduPCS-Go/pcsutil/checksum"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"github.com/iikira/BaiduPCS-Go/pcsutil/waitgroup"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/BaiduPCS-Go/requester/downloader"
	"github.com/oleiade/lane"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

const (
	//DownloadSuffix 文件下载后缀
	DownloadSuffix = ".BaiduPCS-Go-downloading"
	//StrDownloadInitError 初始化下载发生错误
	StrDownloadInitError = "初始化下载发生错误"
	// StrDownloadFailed 下载文件错误
	StrDownloadFailed = "下载文件错误"
	// DefaultDownloadMaxRetry 默认下载失败最大重试次数
	DefaultDownloadMaxRetry = 3
)

var (
	// ErrDownloadNotSupportChecksum 文件不支持校验
	ErrDownloadNotSupportChecksum = errors.New("该文件不支持校验")
	// ErrDownloadChecksumFailed 文件校验失败
	ErrDownloadChecksumFailed = errors.New("该文件校验失败, 文件md5值与服务器记录的不匹配")
	// ErrDownloadFileBanned 违规文件
	ErrDownloadFileBanned = errors.New("该文件可能是违规文件, 不支持校验")
	// ErrDlinkNotFound 未取得下载链接
	ErrDlinkNotFound = errors.New("未取得下载链接")
)

type (
	// dtask 下载任务
	dtask struct {
		ListTask
		path         string                  // 下载的路径
		savePath     string                  // 保存的路径
		downloadInfo *baidupcs.FileDirectory // 文件或目录详情
	}

	//DownloadOptions 下载可选参数
	DownloadOptions struct {
		IsTest                 bool
		IsPrintStatus          bool
		IsExecutedPermission   bool
		IsOverwrite            bool
		IsShareDownload        bool
		IsLocateDownload       bool
		IsLocatePanAPIDownload bool
		IsStreaming            bool
		SaveTo                 string
		Parallel               int
		Load                   int
		MaxRetry               int
		NoCheck                bool
		Out                    io.Writer
	}

	// LocateDownloadOption 获取下载链接可选参数
	LocateDownloadOption struct {
		FromPan bool
	}
)

func downloadPrintFormat(load int) string {
	if load <= 1 {
		return "\r[%d] ↓ %s/%s %s/s in %s, left %s ............"
	}
	return "[%d] ↓ %s/%s %s/s in %s, left %s ...\n"
}

func download(id int, downloadURL, savePath string, loadBalansers []string, client *requester.HTTPClient, newCfg downloader.Config, downloadOptions *DownloadOptions) error {
	var (
		file     *os.File
		writerAt io.WriterAt
		err      error
		exitChan chan struct{}
	)

	if !newCfg.IsTest {
		newCfg.InstanceStatePath = savePath + DownloadSuffix

		// 创建下载的目录
		dir := filepath.Dir(savePath)
		fileInfo, err := os.Stat(dir)
		if err != nil {
			err = os.MkdirAll(dir, 0777)
			if err != nil {
				return err
			}
		} else if !fileInfo.IsDir() {
			return fmt.Errorf("%s, path %s: not a directory", StrDownloadInitError, dir)
		}

		file, err = os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY, 0666)
		if file != nil {
			defer file.Close()
		}
		if err != nil {
			return fmt.Errorf("%s, %s", StrDownloadInitError, err)
		}

		// 空指针和空接口不等价
		if file != nil {
			writerAt = file
		}
	}

	download := downloader.NewDownloader(downloadURL, writerAt, &newCfg)
	download.SetClient(client)
	download.TryHTTP(!pcsconfig.Config.EnableHTTPS())
	download.AddLoadBalanceServer(loadBalansers...)
	download.SetStatusCodeBodyCheckFunc(func(respBody io.Reader) error {
		return pcserror.DecodePCSJSONError(baidupcs.OperationDownloadFile, respBody)
	})

	exitChan = make(chan struct{})

	download.OnExecute(func() {
		if downloadOptions.IsPrintStatus {
			go func() {
				for {
					time.Sleep(1 * time.Second)
					select {
					case <-exitChan:
						return
					default:
						download.PrintAllWorkers()
					}
				}
			}()
		}

		if newCfg.IsTest {
			fmt.Fprintf(downloadOptions.Out, "[%d] 测试下载开始\n\n", id)
		}

		var (
			ds                            = download.GetDownloadStatusChan()
			format                        = downloadPrintFormat(downloadOptions.Load)
			downloaded, totalSize, speeds int64
			leftStr                       string
		)
		for {
			select {
			case <-exitChan:
				return
			case v, ok := <-ds:
				if !ok { // channel 已经关闭
					return
				}

				downloaded, totalSize, speeds = v.Downloaded(), v.TotalSize(), v.SpeedsPerSecond()
				if speeds <= 0 {
					leftStr = "-"
				} else {
					leftStr = (time.Duration((totalSize-downloaded)/(speeds)) * time.Second).String()
				}

				fmt.Fprintf(downloadOptions.Out, format, id,
					converter.ConvertFileSize(v.Downloaded(), 2),
					converter.ConvertFileSize(v.TotalSize(), 2),
					converter.ConvertFileSize(v.SpeedsPerSecond(), 2),
					v.TimeElapsed()/1e7*1e7, leftStr,
				)
			}
		}
	})

	err = download.Execute()
	close(exitChan)
	fmt.Fprintf(downloadOptions.Out, "\n")
	if err != nil {
		// 下载失败, 删去空文件
		if info, infoErr := file.Stat(); infoErr == nil {
			if info.Size() == 0 {
				pcsCommandVerbose.Infof("[%d] remove empty file: %s\n", id, savePath)
				os.Remove(savePath)
			}
		}
		return err
	}

	if downloadOptions.IsExecutedPermission {
		err = file.Chmod(0766)
		if err != nil {
			fmt.Fprintf(downloadOptions.Out, "[%d] 警告, 加执行权限错误: %s\n", id, err)
		}
	}

	if !newCfg.IsTest {
		fmt.Fprintf(downloadOptions.Out, "[%d] 下载完成, 保存位置: %s\n", id, savePath)
	} else {
		fmt.Fprintf(downloadOptions.Out, "[%d] 测试下载结束\n", id)
	}

	return nil
}

// checkFileValid 检测文件有效性
func checkFileValid(filePath string, fileInfo *baidupcs.FileDirectory) error {
	if len(fileInfo.BlockList) != 1 {
		return ErrDownloadNotSupportChecksum
	}

	f := checksum.NewLocalFileInfo(filePath, int(256*converter.KB))
	err := f.OpenPath()
	if err != nil {
		return err
	}

	defer f.Close()

	f.Md5Sum()
	md5Str := hex.EncodeToString(f.MD5)

	if strings.Compare(md5Str, fileInfo.MD5) != 0 {
		// 检测是否为违规文件
		if pcsdownload.IsSkipMd5Checksum(f.Length, md5Str) {
			return ErrDownloadFileBanned
		}
		return ErrDownloadChecksumFailed
	}
	return nil
}

// RunDownload 执行下载网盘内文件
func RunDownload(paths []string, options *DownloadOptions) {
	if options == nil {
		options = &DownloadOptions{}
	}

	if options.Out == nil {
		options.Out = os.Stdout
	}

	if options.Load <= 0 {
		options.Load = pcsconfig.Config.MaxDownloadLoad()
	}

	if options.MaxRetry < 0 {
		options.MaxRetry = DefaultDownloadMaxRetry
	}

	// 设置下载配置
	cfg := &downloader.Config{
		IsTest:    options.IsTest,
		CacheSize: pcsconfig.Config.CacheSize(),
	}

	// 设置下载最大并发量
	if options.Parallel < 1 {
		options.Parallel = pcsconfig.Config.MaxParallel()
	}

	paths, err := matchPathByShellPattern(paths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Fprintf(options.Out, "\n")
	fmt.Fprintf(options.Out, "[0] 提示: 当前下载最大并发量为: %d, 下载缓存为: %d\n", options.Parallel, cfg.CacheSize)

	var (
		pcs       = GetBaiduPCS()
		dlist     = lane.NewDeque()
		lastID    = 0
		loadCount = 0
	)

	// 预测要下载的文件数量
	// TODO: pcscache
	for k := range paths {
		pcs.FilesDirectoriesRecurseList(paths[k], baidupcs.DefaultOrderOptions, func(depth int, _ string, fd *baidupcs.FileDirectory, pcsError pcserror.Error) bool {
			if pcsError != nil {
				pcsCommandVerbose.Warnf("%s\n", pcsError)
				return true
			}

			if !fd.Isdir {
				loadCount++
				if loadCount >= options.Load {
					return false
				}
			}
			return true
		})

		if loadCount >= options.Load {
			break
		}
	}
	// 修改Load, 设置MaxParallel
	if loadCount > 0 {
		options.Load = loadCount
		// 取平均值
		cfg.MaxParallel = pcsconfig.AverageParallel(options.Parallel, loadCount)
	} else {
		cfg.MaxParallel = options.Parallel
	}

	// 处理队列
	for k := range paths {
		lastID++
		ptask := &dtask{
			ListTask: ListTask{
				ID:       lastID,
				MaxRetry: options.MaxRetry,
			},
			path: paths[k],
		}
		if options.SaveTo != "" {
			ptask.savePath = filepath.Join(options.SaveTo, filepath.Base(paths[k]))
		} else {
			ptask.savePath = GetActiveUser().GetSavePath(paths[k])
		}
		dlist.Append(ptask)
		fmt.Fprintf(options.Out, "[%d] 加入下载队列: %s\n", lastID, paths[k])
	}

	var (
		totalSize     int64
		failedList    []string
		handleTaskErr = func(task *dtask, errManifest string, err error) {
			if task == nil {
				panic("task is nil")
			}

			if err == nil {
				return
			}

			// 不重试的情况
			switch {
			case err == ErrDownloadNotSupportChecksum:
				fallthrough
			case strings.Compare(errManifest, StrDownloadFailed) == 0 && strings.Contains(err.Error(), StrDownloadInitError):
				fmt.Fprintf(options.Out, "[%d] %s, %s\n", task.ID, errManifest, err)
				return
			}

			// 未达到失败重试最大次数, 将任务推送到队列末尾
			if task.retry < task.MaxRetry {
				task.retry++
				fmt.Fprintf(options.Out, "[%d] %s, %s, 重试 %d/%d\n", task.ID, errManifest, err, task.retry, task.MaxRetry)
				dlist.Append(task)
				time.Sleep(3 * time.Duration(task.retry) * time.Second)
			} else {
				fmt.Fprintf(options.Out, "[%d] %s, %s\n", task.ID, errManifest, err)
				failedList = append(failedList, task.path)
			}

			switch err {
			case ErrDownloadChecksumFailed:
				// 删去旧的文件, 重新下载
				rerr := os.Remove(task.savePath)
				if rerr != nil {
					fmt.Fprintf(options.Out, "[%d] 删除校验失败的文件出错, %s\n", task.ID, rerr)
					failedList = append(failedList, task.path)
					return
				}

				fmt.Fprintf(options.Out, "[%d] 已删除校验失败的文件\n", task.ID)
			}
		}
		startTime = time.Now()
		wg        = waitgroup.NewWaitGroup(options.Load)
	)

	for {
		e := dlist.Shift()
		if e == nil { // 任务为空
			if wg.Parallel() == 0 { // 结束
				break
			} else {
				time.Sleep(1e9)
				continue
			}
		}

		task := e.(*dtask)
		wg.AddDelta()
		go func() {
			defer wg.Done()

			if task.downloadInfo == nil {
				task.downloadInfo, err = pcs.FilesDirectoriesMeta(task.path)
				if err != nil {
					// 不重试
					fmt.Printf("[%d] 获取路径信息错误, %s\n", task.ID, err)
					return
				}
			}

			fmt.Fprintf(options.Out, "\n")
			fmt.Fprintf(options.Out, "[%d] ----\n%s\n", task.ID, task.downloadInfo.String())

			// 如果是一个目录, 将子文件和子目录加入队列
			if task.downloadInfo.Isdir {
				if !options.IsTest { // 测试下载, 不建立空目录
					os.MkdirAll(task.savePath, 0777) // 首先在本地创建目录, 保证空目录也能被保存
				}

				fileList, err := pcs.FilesDirectoriesList(task.path, baidupcs.DefaultOrderOptions)
				if err != nil {
					// 不重试
					fmt.Fprintf(options.Out, "[%d] 获取目录信息错误, %s\n", task.ID, err)
					return
				}

				for k := range fileList {
					lastID++
					subTask := &dtask{
						ListTask: ListTask{
							ID:       lastID,
							MaxRetry: options.MaxRetry,
						},
						path:         fileList[k].Path,
						downloadInfo: fileList[k],
					}

					if options.SaveTo != "" {
						subTask.savePath = filepath.Join(task.savePath, fileList[k].Filename)
					} else {
						subTask.savePath = GetActiveUser().GetSavePath(subTask.path)
					}

					dlist.Append(subTask)
					fmt.Fprintf(options.Out, "[%d] 加入下载队列: %s\n", lastID, fileList[k].Path)
				}
				return
			}

			fmt.Fprintf(options.Out, "[%d] 准备下载: %s\n", task.ID, task.path)

			if !options.IsTest && !options.IsOverwrite && fileExist(task.savePath) {
				fmt.Fprintf(options.Out, "[%d] 文件已经存在: %s, 跳过...\n", task.ID, task.savePath)
				return
			}

			if !options.IsTest {
				fmt.Fprintf(options.Out, "[%d] 将会下载到路径: %s\n\n", task.ID, task.savePath)
			}

			// 获取直链, 或者以分享文件的方式获取下载链接来下载
			var (
				dlink  string
				dlinks []string
			)

			switch {
			case options.IsLocateDownload:
				// 获取直链下载
				var rawDlinks []*url.URL
				rawDlinks, err = getLocateDownloadLinks(task.path)
				if err == nil {
					handleHTTPLinkURL(rawDlinks[0])
					dlink = rawDlinks[0].String()
					dlinks = make([]string, 0, len(rawDlinks)-1)
					for _, rawDlink := range rawDlinks[1:len(rawDlinks)] {
						handleHTTPLinkURL(rawDlink)
						dlinks = append(dlinks, rawDlink.String())
					}
				}
			case options.IsShareDownload: // 分享下载
				dlink, err = getShareDLink(task.path)
				switch err {
				case nil, ErrShareInfoNotFound: // 未分享, 采用默认下载方式
				default:
					handleTaskErr(task, StrDownloadFailed, err)
					return
				}
			case options.IsLocatePanAPIDownload: // 由第三方服务器处理
				dlink, err = getLocatePanLink(pcs, task.downloadInfo.FsID)
				if err != nil {
					handleTaskErr(task, StrDownloadFailed, err)
					return
				}
			}

			if (options.IsShareDownload || options.IsLocateDownload || options.IsLocatePanAPIDownload) && err == nil {
				pcsCommandVerbose.Infof("[%d] 获取到下载链接: %s\n", task.ID, dlink)
				client := pcsconfig.Config.HTTPClient()
				client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					// 去掉 Referer
					if !pcsconfig.Config.EnableHTTPS() {
						req.Header.Del("Referer")
					}
					if len(via) >= 10 {
						return errors.New("stopped after 10 redirects")
					}
					return nil
				}
				client.SetTimeout(20 * time.Minute)
				client.SetKeepAlive(true)
				err = download(task.ID, dlink, task.savePath, dlinks, client, *cfg, options)
			} else {
				if options.IsShareDownload || options.IsLocateDownload || options.IsLocatePanAPIDownload {
					fmt.Fprintf(options.Out, "[%d] 错误: %s, 将使用默认的下载方式\n", task.ID, err)
				}

				dfunc := func(downloadURL string, jar http.CookieJar) error {
					h := pcsconfig.Config.HTTPClient()
					h.SetCookiejar(jar)
					h.SetKeepAlive(true)
					h.SetTimeout(10 * time.Minute)

					err := download(task.ID, downloadURL, task.savePath, dlinks, h, *cfg, options)
					if err != nil {
						return err
					}

					return nil
				}
				if options.IsStreaming {
					err = pcs.DownloadStreamFile(task.path, dfunc)
				} else {
					err = pcs.DownloadFile(task.path, dfunc)
				}
			}

			if err != nil {
				handleTaskErr(task, StrDownloadFailed, err)
				return
			}

			// 检验文件有效性
			if !cfg.IsTest && !options.NoCheck {
				if task.downloadInfo.Size >= 128*converter.MB {
					fmt.Fprintf(options.Out, "[%d] 开始检验文件有效性, 稍后...\n", task.ID)
				}
				err = checkFileValid(task.savePath, task.downloadInfo)
				if err != nil {
					switch err {
					case ErrDownloadFileBanned:
						fmt.Fprintf(options.Out, "[%d] 检验文件有效性: %s\n", task.ID, err)
						return
					default:
						handleTaskErr(task, "检验文件有效性出错", err)
						return
					}
				}

				fmt.Fprintf(options.Out, "[%d] 检验文件有效性成功\n", task.ID)
			}

			atomic.AddInt64(&totalSize, task.downloadInfo.Size)
		}()
	}
	wg.Wait()

	fmt.Fprintf(options.Out, "\n任务结束, 时间: %s, 数据总量: %s\n", time.Since(startTime)/1e6*1e6, converter.ConvertFileSize(totalSize))
	if len(failedList) != 0 {
		fmt.Printf("以下文件下载失败: \n")
		tb := pcstable.NewTable(os.Stdout)
		for k := range failedList {
			tb.Append([]string{strconv.Itoa(k), failedList[k]})
		}
		tb.Render()
	}
}

// RunLocateDownload 执行获取直链
func RunLocateDownload(pcspaths []string, opt *LocateDownloadOption) {
	if opt == nil {
		opt = &LocateDownloadOption{}
	}

	absPaths, err := matchPathByShellPattern(pcspaths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	pcs := GetBaiduPCS()

	if opt.FromPan {
		fds, err := pcs.FilesDirectoriesBatchMeta(absPaths...)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		fidList := make([]int64, 0, len(fds))
		for i := range fds {
			fidList = append(fidList, fds[i].FsID)
		}

		list, err := pcs.LocatePanAPIDownload(fidList...)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		tb := pcstable.NewTable(os.Stdout)
		tb.SetHeader([]string{"#", "fs_id", "路径", "链接"})

		var (
			i          int
			fidStrList = converter.SliceInt64ToString(fidList)
		)
		for k := range fidStrList {
			for i = range list {
				if fidStrList[k] == list[i].FsID {
					tb.Append([]string{strconv.Itoa(k), list[i].FsID, fds[k].Path, list[i].Dlink})
					list = append(list[:i], list[i+1:]...)
					break
				}
			}
		}
		tb.Render()
		fmt.Printf("\n注意: 以上链接不能直接访问, 需要登录百度帐号才可以下载\n")
		return
	}

	for i, pcspath := range absPaths {
		info, err := pcs.LocateDownload(pcspath)
		if err != nil {
			fmt.Printf("[%d] %s, 路径: %s\n", i, err, pcspath)
			continue
		}

		fmt.Printf("[%d] %s: \n", i, pcspath)
		tb := pcstable.NewTable(os.Stdout)
		tb.SetHeader([]string{"#", "链接"})
		for k, u := range info.URLStrings(pcsconfig.Config.EnableHTTPS()) {
			tb.Append([]string{strconv.Itoa(k), u.String()})
		}
		tb.Render()
		fmt.Println()
	}
	fmt.Printf("提示: 访问下载链接, 需将下载器的 User-Agent 设置为: %s\n", pcsconfig.Config.UserAgent())
}

// RunFixMD5 执行修复md5
func RunFixMD5(pcspaths ...string) {
	absPaths, err := matchPathByShellPattern(pcspaths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	pcs := GetBaiduPCS()
	finfoList, err := pcs.FilesDirectoriesBatchMeta(absPaths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	for k, finfo := range finfoList {
		err := pcs.FixMD5ByFileInfo(finfo)
		if err == nil {
			fmt.Printf("[%d] - [%s] 修复md5成功\n", k, finfo.Path)
			continue
		}

		if err.GetError() == baidupcs.ErrFixMD5Failed {
			fmt.Printf("[%d] - [%s] 修复md5失败, 可能是服务器未刷新\n", k, finfo.Path)
			continue
		}
		fmt.Printf("[%d] - [%s] 修复md5失败, 错误信息: %s\n", k, finfo.Path, err)
	}
}

func getLocateDownloadLinks(pcspath string) (dlinks []*url.URL, err error) {
	pcs := GetBaiduPCS()
	dInfo, pcsError := pcs.LocateDownload(pcspath)
	if pcsError != nil {
		return nil, pcsError
	}

	us := dInfo.URLStrings(pcsconfig.Config.EnableHTTPS())
	if len(us) == 0 {
		return nil, ErrDlinkNotFound
	}

	return us, nil
}

func getLocatePanLink(pcs *baidupcs.BaiduPCS, fsID int64) (dlink string, err error) {
	list, err := pcs.LocatePanAPIDownload(fsID)
	if err != nil {
		return
	}

	var link string
	for k := range list {
		if strconv.FormatInt(fsID, 10) == list[k].FsID {
			link = list[k].Dlink
		}
	}

	if link == "" {
		return "", ErrDlinkNotFound
	}

	dc := dlinkclient.NewDlinkClient()
	c := pcsconfig.Config.HTTPClient()
	c.SetResponseHeaderTimeout(30 * time.Second)
	dc.SetClient(c)
	dlink, err = dc.CacheLinkRedirectPr(link)
	return
}

func handleHTTPLinkURL(linkURL *url.URL) {
	if pcsconfig.Config.EnableHTTPS() {
		if linkURL.Scheme == "http" {
			linkURL.Scheme = "https"
		}
	}
}

// fileExist 检查文件是否存在,
// 只有当文件存在, 文件大小不为0或断点续传文件不存在时, 才判断为存在
func fileExist(path string) bool {
	if info, err := os.Stat(path); err == nil {
		if info.Size() == 0 {
			return false
		}
		if _, err = os.Stat(path + DownloadSuffix); err != nil {
			return true
		}
	}

	return false
}
