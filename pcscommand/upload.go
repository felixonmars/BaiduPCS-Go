package pcscommand

import (
	"bufio"
	"crypto/md5"
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

type localPathInfo struct {
	path     string
	file     *os.File
	length   int64
	sliceMD5 string
	md5      string
	crc32    string
}

func (lp *localPathInfo) check() bool {
	var err error
	lp.file, err = os.Open(lp.path)
	if err != nil {
		return false
	}
	info, _ := lp.file.Stat()
	lp.length = info.Size()
	return true
}

// md5Sum 获取 文件 md5
func (lp *localPathInfo) md5Sum() {
	bf := bufio.NewReader(lp.file)
	m := md5.New()
	bf.WriteTo(m)
	lp.md5 = fmt.Sprintf("%x", m.Sum(nil))

	// reset
	lp.file, _ = os.Open(lp.path)
}

// otherSum 获取其他的校验值
func (lp *localPathInfo) otherSum() {
	bf := bufio.NewReader(lp.file)

	// 获取前 256KB 文件切片的 md5
	buf := make([]byte, requiredSliceLen)
	lp.file.ReadAt(buf[:], requiredSliceLen)
	lp.sliceMD5 = fmt.Sprintf("%x", md5.Sum(buf[:]))

	bf.Reset(lp.file)

	// 获取 文件 crc32
	c := crc32.NewIEEE()
	bf.WriteTo(c)
	lp.crc32 = fmt.Sprint(c.Sum32())

	// reset
	lp.file, _ = os.Open(lp.path)
	bf.Reset(nil)
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

				localPathInfo := localPathInfo{
					path: file,
				}

				if !localPathInfo.check() {
					fmt.Printf("文件不可读, 跳过...\n")
					return
				}

				defer localPathInfo.file.Close() // 关闭文件

				_targetPath := targetPath + "/" + strings.TrimPrefix(localPathInfo.path, uploadInfo.dir)

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
					if fd.MD5 == localPathInfo.md5 {
						fmt.Printf("目标文件已存在, 跳过...\n")
						return
					}
				}

				// 文件大于256kb, 检测秒传
				if localPathInfo.length >= requiredSliceLen {
					fmt.Printf("检测秒传中, 请稍候...\n")

					localPathInfo.otherSum()

					err := info.RapidUpload(_targetPath, localPathInfo.md5, localPathInfo.sliceMD5, localPathInfo.crc32, localPathInfo.length)
					if err == nil {
						fmt.Printf("秒传成功\n")
						return
					}
					fmt.Printf("秒传失败, 开始上传文件...\n")
				}

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
				fmt.Printf("上传文件成功, 保存位置: %s\n", _targetPath)
			}()
		}
	}
}
