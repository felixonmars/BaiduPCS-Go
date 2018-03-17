package pcscommand

import (
	"bufio"
	"bytes"
	"container/list"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/pcscache"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/BaiduPCS-Go/requester/multipartreader"
	"github.com/iikira/BaiduPCS-Go/uploader"
	"github.com/json-iterator/go"
	"hash/crc32"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const requiredSliceLen = 256 * pcsutil.KB // 256 KB

type utask struct {
	ListTask
	uploadInfo *LocalPathInfo // 要上传的本地文件详情
	savePath   string
}

// LocalPathInfo 本地文件详情
type LocalPathInfo struct {
	Path string   // 本地路径
	File *os.File // 文件

	Length   int64  // 文件大小
	SliceMD5 []byte // 文件前 requiredSliceLen (256KB) 切片的 md5 值
	MD5      []byte // 文件的 md5
	CRC32    uint32 // 文件的 crc32
}

// check 检查文件状态并获取文件的大小 (Length)
func (lp *LocalPathInfo) check() bool {
	var err error
	lp.File, err = os.Open(lp.Path)
	if err != nil {
		return false
	}
	info, _ := lp.File.Stat()
	lp.Length = info.Size()
	return true
}

// md5Sum 获取文件的 md5 值
func (lp *LocalPathInfo) md5Sum() {
	bf := bufio.NewReader(lp.File)
	m := md5.New()
	bf.WriteTo(m)
	lp.MD5 = m.Sum(nil)

	// reset
	lp.File, _ = os.Open(lp.Path)
	bf.Reset(nil)
}

// sliceMD5Sum 获取文件前 requiredSliceLen (256KB) 切片的 md5 值
func (lp *LocalPathInfo) sliceMD5Sum() {
	// 获取前 256KB 文件切片的 md5
	buf := make([]byte, requiredSliceLen)
	lp.File.ReadAt(buf[:], requiredSliceLen)
	sliceMD5 := md5.Sum(buf[:])
	lp.SliceMD5 = sliceMD5[:]
}

// crc32Sum 获取文件的 crc32 值
func (lp *LocalPathInfo) crc32Sum() {
	bf := bufio.NewReader(lp.File)

	// 获取 文件 crc32
	c := crc32.NewIEEE()
	bf.WriteTo(c)
	lp.CRC32 = c.Sum32()

	// reset
	lp.File, _ = os.Open(lp.Path)
	bf.Reset(nil)
}

// RunRapidUpload 执行秒传文件, 前提是知道文件的大小, md5, 前256KB切片的 md5, crc32
func RunRapidUpload(targetPath, contentMD5, sliceMD5, crc32 string, length int64) {
	targetPath, err := getAbsPath(targetPath)
	if err != nil {
		fmt.Printf("警告: 尝试秒传文件, 获取网盘路径 %s 错误, %s\n", targetPath, err)
	}

	// 检测文件是否存在于网盘路径
	// 很重要, 如果文件存在会直接覆盖!!! 即使是根目录!
	if info.Isdir(targetPath) {
		fmt.Printf("错误: 路径 %s 是一个目录, 不可覆盖\n", targetPath)
		return
	}

	if sliceMD5 == "" {
		sliceMD5 = "ec87a838931d4d5d2e94a04644788a55" // 长度为32
	}

	err = info.RapidUpload(targetPath, contentMD5, sliceMD5, crc32, length)
	if err != nil {
		fmt.Printf("秒传失败, 消息: %s\n", err)
		return
	}

	fmt.Printf("秒传成功, 保存到网盘路径: %s\n", targetPath)
	return
}

// RunUpload 执行文件上传
func RunUpload(localPaths []string, savePath string) {
	absSavePath, err := getAbsPath(savePath)
	if err != nil {
		fmt.Printf("警告: 上传文件, 获取网盘路径 %s 错误, %s\n", savePath, err)
	}

	switch len(localPaths) {
	case 0:
		fmt.Printf("本地路径为空\n")
		return
	}

	var (
		ulist         = list.New()
		lastID        int
		globedPathDir string
		subSavePath   string
	)

	for k := range localPaths {
		globedPaths, err := filepath.Glob(localPaths[k])
		if err != nil {
			fmt.Printf("上传文件, 匹配本地路径失败, %s\n", err)
			continue
		}

		for k2 := range globedPaths {
			walkedFiles, err := pcsutil.WalkDir(globedPaths[k2], "")
			if err != nil {
				fmt.Printf("警告: %s\n", err)
				continue
			}

			for k3 := range walkedFiles {
				// 针对 windows 的目录处理
				if os.PathSeparator == '\\' {
					walkedFiles[k3] = pcsutil.ConvertToUnixPathSeparator(walkedFiles[k3])
					globedPathDir = pcsutil.ConvertToUnixPathSeparator(filepath.Dir(globedPaths[k2]))
				} else {
					globedPathDir = filepath.Dir(globedPaths[k2])
				}

				subSavePath = strings.TrimPrefix(walkedFiles[k3], globedPathDir)

				lastID++
				ulist.PushBack(&utask{
					ListTask: ListTask{
						id:       lastID,
						maxRetry: 3,
					},
					uploadInfo: &LocalPathInfo{
						Path: walkedFiles[k3],
					},
					savePath: path.Clean(absSavePath + "/" + subSavePath),
				})

				fmt.Printf("[%d] 加入上传队列: %s\n", lastID, walkedFiles[k3])
			}
		}
	}

	if lastID == 0 {
		fmt.Printf("未检测到上传的文件, 请检查文件路径或通配符是否正确.\n")
		return
	}

	var (
		e             *list.Element
		task          *utask
		handleTaskErr = func(task *utask, errManifest string, err error) {
			if task == nil {
				panic("task is nil")
			}

			if err == nil {
				return
			}

			// 不重试的情况
			switch {
			case strings.Contains(err.Error(), baidupcs.StrRemoteError):
				fmt.Printf("[%d] %s, %s\n", task.id, errManifest, err)
				return
			}

			fmt.Printf("[%d] %s, %s, 重试 %d/%d\n", task.id, errManifest, err, task.retry, task.maxRetry)

			// 未达到失败重试最大次数, 将任务推送到队列末尾
			if task.retry < task.maxRetry {
				task.retry++
				ulist.PushBack(task)
			}
			time.Sleep(3 * time.Second)
		}
		totalSize int64
	)

	for {
		e = ulist.Front()
		if e == nil { // 结束
			break
		}

		ulist.Remove(e) // 载入任务后, 移除队列

		task = e.Value.(*utask)
		if task == nil {
			continue
		}

		fmt.Printf("[%d] 准备上传: %s\n", task.id, task.uploadInfo.Path)

		if !task.uploadInfo.check() {
			fmt.Printf("[%d] 文件不可读, 跳过...\n", task.id)
			continue
		}

		panDir, panFile := path.Split(task.savePath)

		// 设置缓存
		if !pcscache.DirCache.Existed(panDir) {
			fdl, err := info.FilesDirectoriesList(panDir, false)
			if err == nil {
				pcscache.DirCache.Set(panDir, &fdl)
			}
		}

		if task.uploadInfo.Length >= 128*pcsutil.MB {
			fmt.Printf("[%d] 检测秒传中, 请稍候...\n", task.id)
		}

		task.uploadInfo.md5Sum()

		// 检测缓存, 通过文件的md5值判断本地文件和网盘文件是否一样
		fd := pcscache.DirCache.FindFileDirectory(panDir, panFile)
		if fd != nil {
			decodedMD5, _ := hex.DecodeString(fd.MD5)
			if bytes.Compare(decodedMD5, task.uploadInfo.MD5) == 0 {
				fmt.Printf("[%d] 目标文件, %s, 已存在, 跳过...\n", task.id, task.savePath)
				continue
			}
		}

		// 文件大于256kb, 应该要检测秒传, 反之则不应检测秒传
		// 经测试, 秒传文件并非一定要大于256KB
		if task.uploadInfo.Length >= requiredSliceLen {
			// do nothing
		}

		// 经过测试, 秒传文件并非需要前256kb切片的md5值, 只需格式符合即可
		task.uploadInfo.sliceMD5Sum()

		// 经测试, 文件的 crc32 值并非秒传文件所必需
		// task.uploadInfo.crc32Sum()

		err := info.RapidUpload(task.savePath, hex.EncodeToString(task.uploadInfo.MD5), hex.EncodeToString(task.uploadInfo.SliceMD5), fmt.Sprint(task.uploadInfo.CRC32), task.uploadInfo.Length)
		if err == nil {
			fmt.Printf("[%d] 秒传成功, 保存到网盘路径: %s\n\n", task.id, task.savePath)
			totalSize += task.uploadInfo.Length
			continue
		}

		fmt.Printf("[%d] 秒传失败, 开始上传文件...\n\n", task.id)

		// 秒传失败, 开始上传文件
		err = info.Upload(task.savePath, func(uploadURL string, jar *cookiejar.Jar) (uperr error) {
			h := requester.NewHTTPClient()
			h.SetCookiejar(jar)

			u := uploader.NewUploader(uploadURL, multipartreader.NewFileReadedLen64(task.uploadInfo.File), &uploader.Options{
				IsMultiPart: true,
				Client:      h,
			})

			exit := make(chan struct{})

			u.OnExecute(func() {
				for {
					select {
					case v, ok := <-u.UploadStatus:
						if !ok {
							return
						}

						if v.Length == 0 {
							fmt.Printf("\r[%d] Prepareing upload...", task.id)
							continue
						}

						fmt.Printf("\r[%d] ↑ %s/%s %s/s in %s ............", task.id,
							pcsutil.ConvertFileSize(v.Uploaded, 2),
							pcsutil.ConvertFileSize(v.Length, 2),
							pcsutil.ConvertFileSize(v.Speed, 2),
							v.TimeElapsed,
						)
					}
				}
			})

			u.OnFinish(func() {
				exit <- struct{}{}
			})

			<-u.Execute(func(resp *http.Response, err error) {
				if err != nil {
					uperr = err
					return
				}

				defer resp.Body.Close()

				// http 响应错误处理
				switch resp.StatusCode {
				case 413: // Request Entity Too Large
					// 上传的文件太大了
					uperr = fmt.Errorf(resp.Status)
					return
				}

				// 数据处理
				jsonData := &struct {
					Path string `json:"path"`
					*baidupcs.ErrInfo
				}{
					ErrInfo: baidupcs.NewErrorInfo("上传文件"),
				}

				d := jsoniter.NewDecoder(resp.Body)

				err = d.Decode(jsonData)
				if err != nil {
					uperr = fmt.Errorf("json parse error, %s", err)
					return
				}

				if jsonData.ErrCode != 0 {
					uperr = jsonData.ErrInfo
					return
				}

				if jsonData.Path == "" {
					uperr = fmt.Errorf("unknown response data, file saved path not found")
					return
				}

				task.savePath = jsonData.Path
			})

			<-exit
			close(exit)
			return uperr
		})

		fmt.Printf("\n")

		if err != nil {
			handleTaskErr(task, "上传文件失败", err)
			continue
		}

		fmt.Printf("[%d] 上传文件成功, 保存到网盘路径: %s\n", task.id, task.savePath)
		totalSize += task.uploadInfo.Length
	}

	fmt.Printf("\n")
	fmt.Printf("全部上传完毕, 总大小: %s\n", pcsutil.ConvertFileSize(totalSize))
}

// GetFileSum 获取文件的大小, md5, 前256KB切片的 md5, crc32,
// sliceMD5Only 只获取前256KB切片的 md5
func GetFileSum(localPath string, sliceMD5Only bool) (lp *LocalPathInfo, err error) {
	file, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}

	fileStat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if fileStat.IsDir() {
		return nil, fmt.Errorf("sum %s: is a directory", localPath)
	}

	lp = &LocalPathInfo{
		Path:   localPath,
		File:   file,
		Length: fileStat.Size(),
	}

	lp.sliceMD5Sum()

	if !sliceMD5Only {
		lp.crc32Sum()
		lp.md5Sum()
	}

	return lp, nil
}
