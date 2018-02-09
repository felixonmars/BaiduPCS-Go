package pcscommand

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/iikira/BaiduPCS-Go/pcscache"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/BaiduPCS-Go/uploader"
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

type upload struct {
	dir   string   // 目录路径
	files []string // 目录下所有子文件
}

// LocalPathInfo 本地文件详情
type LocalPathInfo struct {
	path string   // 本地路径
	file *os.File // 文件

	Length   int64  // 文件大小
	SliceMD5 []byte // 文件前 requiredSliceLen (256KB) 切片的 md5 值
	MD5      []byte // 文件的 md5
	CRC32    uint32 // 文件的 crc32
}

func (lp *LocalPathInfo) check() bool {
	var err error
	lp.file, err = os.Open(lp.path)
	if err != nil {
		return false
	}
	info, _ := lp.file.Stat()
	lp.Length = info.Size()
	return true
}

// md5Sum 获取文件的 md5 值
func (lp *LocalPathInfo) md5Sum() {
	bf := bufio.NewReader(lp.file)
	m := md5.New()
	bf.WriteTo(m)
	lp.MD5 = m.Sum(nil)

	// reset
	lp.file, _ = os.Open(lp.path)
	bf.Reset(nil)
}

// sliceMD5Sum 获取文件前 requiredSliceLen (256KB) 切片的 md5 值
func (lp *LocalPathInfo) sliceMD5Sum() {
	// 获取前 256KB 文件切片的 md5
	buf := make([]byte, requiredSliceLen)
	lp.file.ReadAt(buf[:], requiredSliceLen)
	sliceMD5 := md5.Sum(buf[:])
	lp.SliceMD5 = sliceMD5[:]
}

// crc32Sum 获取文件的 crc32 值
func (lp *LocalPathInfo) crc32Sum() {
	bf := bufio.NewReader(lp.file)

	// 获取 文件 crc32
	c := crc32.NewIEEE()
	bf.WriteTo(c)
	lp.CRC32 = c.Sum32()

	// reset
	lp.file, _ = os.Open(lp.path)
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
func RunUpload(localPaths []string, targetPath string) {
	targetPath, err := getAbsPath(targetPath)
	if err != nil {
		fmt.Printf("警告: 上传文件, 获取网盘路径 %s 错误, %s\n", targetPath, err)
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

	for ftN, uploadInfo := range uploads {

		filesNum := len(uploadInfo.files)

		for fN, file := range uploadInfo.files {
			func() {
				defer fmt.Println()

				fmt.Printf("[%d/%d - %d/%d] - [%s]: 任务开始\n", ftN+1, filesTotalNum, fN+1, filesNum, file)

				localPathInfo := LocalPathInfo{
					path: file,
				}

				if !localPathInfo.check() {
					fmt.Printf("文件不可读, 跳过...\n")
					return
				}

				defer localPathInfo.file.Close() // 关闭文件

				_targetPath := path.Clean(targetPath + "/" + strings.TrimPrefix(localPathInfo.path, uploadInfo.dir))

				panDir, panFile := path.Split(_targetPath)

				// 设置缓存
				if !pcscache.DirCache.Existed(panDir) {
					fdl, err := info.FilesDirectoriesList(panDir, false)
					if err == nil {
						pcscache.DirCache.Set(panDir, &fdl)
					}
				}

				localPathInfo.md5Sum()

				// 检测缓存
				fd := pcscache.DirCache.FindFileDirectory(panDir, panFile)
				if fd != nil {
					if strings.Compare(fd.MD5, hex.EncodeToString(localPathInfo.MD5)) == 0 {
						fmt.Printf("目标文件, %s, 已存在, 跳过...\n", _targetPath)
						return
					}
				}

				// 文件大于256kb, 应该要检测秒传, 反之则不检测秒传
				// 经检验, 秒传文件并非一定要大于256KB
				if localPathInfo.Length >= requiredSliceLen {
					// do nothing.
				}

				if localPathInfo.Length >= 128*pcsutil.MB {
					fmt.Printf("检测秒传中, 请稍候...\n")
				}

				localPathInfo.sliceMD5Sum()

				// 经检验, 文件的 crc32 值并非秒传文件所必需
				// localPathInfo.crc32Sum()

				err := info.RapidUpload(_targetPath, hex.EncodeToString(localPathInfo.MD5), hex.EncodeToString(localPathInfo.SliceMD5), fmt.Sprint(localPathInfo.CRC32), localPathInfo.Length)
				if err == nil {
					fmt.Printf("秒传成功, 保存到网盘路径: %s\n", _targetPath)
					return
				}
				fmt.Printf("秒传失败, 开始上传文件...\n")

				// 秒传失败, 开始上传文件
				err = info.Upload(_targetPath, func(uploadURL string, jar *cookiejar.Jar) (uperr error) {
					h := requester.NewHTTPClient()
					h.SetCookiejar(jar)

					u := uploader.NewUploader(uploadURL, true, uploader.NewFileReaderLen(localPathInfo.file), h)

					exit := make(chan struct{})
					exit2 := make(chan struct{})

					u.OnExecute(func() {
						t := time.Now()
						c := u.GetStatusChan()
						for {
							select {
							case <-exit:
								return
							case v := <-c:
								if v.Length == 0 {
									fmt.Printf("\rPrepareing upload...")
									t = time.Now()
									continue
								}
								fmt.Printf("\r%v/%v %v/s time: %s %v",
									pcsutil.ConvertFileSize(v.Uploaded, 2),
									pcsutil.ConvertFileSize(v.Length, 2),
									pcsutil.ConvertFileSize(v.Speed, 2),
									time.Since(t)/1000000*1000000,
									"[UPLOADING]"+strings.Repeat(" ", 10),
								)
							}
						}
					})

					u.OnFinish(func() {
						exit <- struct{}{}
						exit2 <- struct{}{}
					})

					u.Execute(func(resp *http.Response, err error) {
						if err != nil {
							uperr = err
							return
						}

						json, err := simplejson.NewFromReader(resp.Body)
						if err != nil {
							uperr = fmt.Errorf("json parse error, %s", err)
							return
						}

						pj, ok := json.CheckGet("path")
						if !ok {
							uperr = fmt.Errorf("unknown response data, file saved path not found")
							return
						}

						_targetPath = pj.MustString()
					})

					<-exit2
					return uperr
				})

				fmt.Printf("\n")

				if err != nil {
					fmt.Printf("上传文件失败, %s\n", err)
					return
				}
				fmt.Printf("上传文件成功, 保存到网盘路径: %s\n", _targetPath)
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
		path:   localPath,
		file:   file,
		Length: fileStat.Size(),
	}

	lp.sliceMD5Sum()

	if !sliceMD5Only {
		lp.crc32Sum()
		lp.md5Sum()
	}

	return lp, nil
}
