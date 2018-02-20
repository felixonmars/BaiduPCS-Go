package pcscommand

import (
	"bufio"
	"bytes"
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
)

const requiredSliceLen = 256 * pcsutil.KB // 256 KB

type upload struct {
	dir   string   // 目录路径
	files []string // 目录下所有子文件
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
	_, err = info.FilesDirectoriesMeta(targetPath)
	if err == nil {
		fmt.Printf("错误: 路径 %s 已存在\n", targetPath)
		return
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

	var uploads []upload
	for k := range localPaths {
		_paths, err := filepath.Glob(localPaths[k])
		if err != nil {
			fmt.Printf("上传文件, 匹配本地路径失败, %s\n", err)
			continue
		}

		for k2 := range _paths {
			_files, err := pcsutil.WalkDir(_paths[k2], "")
			if err != nil {
				fmt.Println(err)
				continue
			}

			uploads = append(uploads, upload{
				dir:   filepath.Dir(_paths[k2]),
				files: _files,
			})
		}
	}

	filesTotalNum := len(uploads)
	if filesTotalNum == 0 {
		fmt.Printf("未检测到上传的文件, 请检查文件路径或通配符是否正确.\n")
		return
	}

	for ftN, uploadInfo := range uploads {

		filesNum := len(uploadInfo.files)

		for fN, file := range uploadInfo.files {
			func() {
				defer fmt.Println()

				fmt.Printf("[%d/%d - %d/%d] - [%s]: 任务开始\n", ftN+1, filesTotalNum, fN+1, filesNum, file)

				localPathInfo := &LocalPathInfo{
					Path: file,
				}

				if !localPathInfo.check() {
					fmt.Printf("文件不可读, 跳过...\n")
					return
				}

				defer localPathInfo.File.Close() // 关闭文件

				subSavePath := strings.TrimPrefix(localPathInfo.Path, uploadInfo.dir)

				// 针对 windows 的目录处理
				if os.PathSeparator == '\\' {
					subSavePath = strings.Replace(subSavePath, "\\", "/", -1)
				}

				targetPath := path.Clean(absSavePath + "/" + subSavePath)

				panDir, panFile := path.Split(targetPath)

				// 设置缓存
				if !pcscache.DirCache.Existed(panDir) {
					fdl, err := info.FilesDirectoriesList(panDir, false)
					if err == nil {
						pcscache.DirCache.Set(panDir, &fdl)
					}
				}

				if localPathInfo.Length >= 128*pcsutil.MB {
					fmt.Printf("检测秒传中, 请稍候...\n")
				}

				localPathInfo.md5Sum()

				// 检测缓存, 通过文件的md5值判断本地文件和网盘文件是否一样
				fd := pcscache.DirCache.FindFileDirectory(panDir, panFile)
				if fd != nil {
					decodedMD5, _ := hex.DecodeString(fd.MD5)
					if bytes.Compare(decodedMD5, localPathInfo.MD5) == 0 {
						fmt.Printf("目标文件, %s, 已存在, 跳过...\n", targetPath)
						return
					}
				}

				// 文件大于256kb, 应该要检测秒传, 反之则不检测秒传
				// 经检验, 秒传文件并非一定要大于256KB
				if localPathInfo.Length >= requiredSliceLen {
					// do nothing.
				}

				localPathInfo.sliceMD5Sum()

				// 经检验, 文件的 crc32 值并非秒传文件所必需
				// localPathInfo.crc32Sum()

				err := info.RapidUpload(targetPath, hex.EncodeToString(localPathInfo.MD5), hex.EncodeToString(localPathInfo.SliceMD5), fmt.Sprint(localPathInfo.CRC32), localPathInfo.Length)
				if err == nil {
					fmt.Printf("秒传成功, 保存到网盘路径: %s\n", targetPath)
					return
				}
				fmt.Printf("秒传失败, 开始上传文件...\n")

				// 秒传失败, 开始上传文件
				err = info.Upload(targetPath, func(uploadURL string, jar *cookiejar.Jar) (uperr error) {
					h := requester.NewHTTPClient()
					h.SetCookiejar(jar)

					u := uploader.NewUploader(uploadURL, multipartreader.NewFileReadedLen64(localPathInfo.File), &uploader.Options{
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
									fmt.Printf("\rPrepareing upload...")
									continue
								}

								fmt.Printf("\r↑ %s/%s %s/s in %s ............",
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

						targetPath = jsonData.Path
					})

					<-exit
					close(exit)
					return uperr
				})

				fmt.Printf("\n")

				if err != nil {
					fmt.Printf("上传文件失败, %s\n", err)
					return
				}
				fmt.Printf("上传文件成功, 保存到网盘路径: %s\n", targetPath)
			}()
		}
	}
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
