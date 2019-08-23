// Package pcsupdate 更新包
package pcsupdate

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/pcsliner"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/pcsutil/checkaccess"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"github.com/iikira/BaiduPCS-Go/pcsutil/jsonhelper"
	"github.com/iikira/BaiduPCS-Go/requester/downloader"
	"github.com/iikira/BaiduPCS-Go/requester/rio"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// ReleaseName 分享根目录名称
	ReleaseName = "BaiduPCS-Go-releases"
)

type info struct {
	filename    string
	size        int64
	downloadURL string
}

// CheckUpdate 检测更新
func CheckUpdate(version string, yes bool) {
	if !checkaccess.AccessRDWR(pcsutil.ExecutablePath()) {
		fmt.Printf("程序目录不可写, 无法更新.\n")
		return
	}
	fmt.Println("检测更新中, 稍候...")
	c := pcsconfig.Config.HTTPClient()
	resp, err := c.Req("GET", "https://api.github.com/repos/iikira/BaiduPCS-Go/releases/latest", nil, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		fmt.Printf("获取数据错误: %s\n", err)
		return
	}

	releaseInfo := ReleaseInfo{}
	err = jsonhelper.UnmarshalData(resp.Body, &releaseInfo)
	if err != nil {
		fmt.Printf("json数据解析失败: %s\n", err)
		return
	}

	// 没有更新, 或忽略 Beta 版本, 和版本前缀不符的
	if strings.Contains(releaseInfo.TagName, "Beta") || !strings.HasPrefix(releaseInfo.TagName, "v") || version >= releaseInfo.TagName {
		fmt.Printf("未检测到更新!\n")
		return
	}

	fmt.Printf("检测到新版本: %s\n", releaseInfo.TagName)

	line := pcsliner.NewLiner()
	defer line.Close()

	if !yes {
		y, err := line.State.Prompt("是否进行更新 (y/n): ")
		if err != nil {
			fmt.Printf("输入错误: %s\n", err)
			return
		}

		if y != "y" && y != "Y" {
			fmt.Printf("更新取消.\n")
			return
		}
	}

	builder := &strings.Builder{}
	builder.WriteString("BaiduPCS-Go-" + releaseInfo.TagName + "-" + runtime.GOOS + "-.*?")
	if runtime.GOOS == "darwin" && (runtime.GOARCH == "arm" || runtime.GOARCH == "arm64") {
		builder.WriteString("arm")
	} else {
		switch runtime.GOARCH {
		case "amd64":
			builder.WriteString("(amd64|x86_64|x64)")
		case "386":
			builder.WriteString("(386|x86)")
		case "arm":
			builder.WriteString("(armv5|armv7|arm)")
		case "arm64":
			builder.WriteString("arm64")
		case "mips":
			builder.WriteString("mips")
		case "mips64":
			builder.WriteString("mips64")
		case "mipsle":
			builder.WriteString("(mipsle|mipsel)")
		case "mips64le":
			builder.WriteString("(mips64le|mips64el)")
		default:
			builder.WriteString(runtime.GOARCH)
		}
	}
	builder.WriteString("\\.zip")

	exp := regexp.MustCompile(builder.String())

	var targetList []*info
	for _, asset := range releaseInfo.Assets {
		if asset == nil || asset.State != "uploaded" {
			continue
		}

		if exp.MatchString(asset.Name) {
			targetList = append(targetList, &info{
				filename:    asset.Name,
				size:        asset.Size,
				downloadURL: asset.BrowserDownloadURL,
			})
		}
	}

	var target info
	switch len(targetList) {
	case 0:
		fmt.Printf("未匹配到当前系统的程序更新文件, GOOS: %s, GOARCH: %s\n", runtime.GOOS, runtime.GOARCH)
		return
	case 1:
		target = *targetList[0]
	default:
		fmt.Println()
		for k := range targetList {
			fmt.Printf("%d: %s\n", k, targetList[k].filename)
		}

		fmt.Println()
		t, err := line.State.Prompt("输入序号以下载更新: ")
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}

		i, err := strconv.Atoi(t)
		if err != nil {
			fmt.Printf("输入错误: %s\n", err)
			return
		}

		if i < 0 || i >= len(targetList) {
			fmt.Printf("输入错误: 序号不在范围内\n")
			return
		}

		target = *targetList[i]
	}

	if target.size > 0x7fffffff {
		fmt.Printf("file size too large: %d\n", target.size)
		return
	}

	fmt.Printf("准备下载更新: %s\n", target.filename)

	// 开始下载
	buf := rio.NewBuffer(make([]byte, target.size))
	der := downloader.NewDownloader(target.downloadURL, buf, &downloader.Config{
		MaxParallel: 20,
		CacheSize:   10000,
	})
	der.SetClient(c)

	exitChan := make(chan struct{})
	der.OnExecute(func() {
		defer fmt.Println()
		var (
			ds                            = der.GetDownloadStatusChan()
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

				fmt.Printf("\r ↓ %s/%s %s/s in %s, left %s ............",
					converter.ConvertFileSize(v.Downloaded(), 2),
					converter.ConvertFileSize(v.TotalSize(), 2),
					converter.ConvertFileSize(v.SpeedsPerSecond(), 2),
					v.TimeElapsed()/1e7*1e7, leftStr,
				)
			}
		}
	})

	err = der.Execute()
	close(exitChan)
	if err != nil {
		fmt.Printf("下载发生错误: %s\n", err)
		return
	}

	fmt.Printf("下载完毕\n")

	// 读取文件
	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), target.size)
	if err != nil {
		fmt.Printf("读取更新文件发生错误: %s\n", err)
		return
	}

	execPath := pcsutil.ExecutablePath()

	var fileNum, errTimes int
	for _, zipFile := range reader.File {
		if zipFile == nil {
			continue
		}

		info := zipFile.FileInfo()

		if info.IsDir() {
			continue
		}

		rc, err := zipFile.Open()
		if err != nil {
			fmt.Printf("解析 zip 文件错误: %s\n", err)
			continue
		}

		fileNum++

		name := zipFile.Name[strings.Index(zipFile.Name, "/")+1:]
		if name == "BaiduPCS-Go" {
			err = update(pcsutil.Executable(), rc)
		} else {
			err = update(filepath.Join(execPath, name), rc)
		}

		if err != nil {
			errTimes++
			fmt.Printf("发生错误, zip 路径: %s, 错误: %s\n", zipFile.Name, err)
			continue
		}
	}

	if errTimes == fileNum {
		fmt.Printf("更新失败\n")
		return
	}

	fmt.Printf("更新完毕, 请重启程序\n")
}
