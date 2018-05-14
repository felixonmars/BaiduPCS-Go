// Package pan 百度网盘提取分享文件的下载链接
package pan

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/json-iterator/go"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// SharedInfo 百度网盘文件分享页信息
type SharedInfo struct {
	SharedURL string

	UK            int64  `json:"uk"`            // 百度网盘用户id
	ShareID       int64  `json:"shareid"`       // 分享id
	RootSharePath string `json:"rootSharePath"` // 分享的目录, 基于分享者的网盘根目录

	Timestamp int64  // unix 时间戳
	Sign      []byte // 签名

	Client *requester.HTTPClient
}

// NewSharedInfo 解析百度网盘文件分享页信息,
// sharedURL 分享链接
func NewSharedInfo(sharedURL string) (si *SharedInfo) {
	return &SharedInfo{
		SharedURL: sharedURL,
	}
}

func (si *SharedInfo) inited() bool {
	return si.UK != 0 && si.ShareID != 0 && si.RootSharePath != ""
}

func (si *SharedInfo) lazyInit() {
	if si.Client == nil {
		si.Client = requester.NewHTTPClient()
	}
}

// Auth 验证提取码
// passwd 提取码, 没有则留空
func (si *SharedInfo) Auth(passwd string) error {
	if si.SharedURL == "" {
		return ErrSharedInfoNotSetSharedURL
	}

	si.lazyInit()

	// 不自动跳转
	si.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if strings.Contains(req.URL.String(), "surl=") {
			return http.ErrUseLastResponse
		}

		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}

	resp, err := si.Client.Req("GET", si.SharedURL, nil, nil)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	switch resp.StatusCode / 100 {
	case 3: // 需要输入提取密码
		locURL, err := resp.Location()
		if err != nil {
			return fmt.Errorf("检测提取码, 提取 Location 错误, %s", err)
		}

		// 验证提取密码
		body, err := si.Client.Fetch("POST", "https://pan.baidu.com/share/verify?"+locURL.RawQuery, map[string]string{
			"pwd":       passwd,
			"vcode":     "",
			"vcode_str": "",
		}, map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Referer":      "https://pan.baidu.com/",
		})

		if err != nil {
			return fmt.Errorf("验证提取密码网络错误, %s", err)
		}

		jsonData := &RemoteErrInfo{}

		err = jsoniter.Unmarshal(body, jsonData)
		if err != nil {
			return fmt.Errorf("验证提取密码, json数据解析失败, %s", err)
		}

		switch jsonData.ErrNo {
		case 0: // 密码正确
			break
		default:
			return fmt.Errorf("验证提取密码遇到错误, %s", jsonData)
		}
	case 4, 5:
		return fmt.Errorf(resp.Status)
	}

	return nil
}

// InitInfo 获取 UK, ShareID, RootSharePath 如果有提取码, 先需进行验证
func (si *SharedInfo) InitInfo() error {
	si.lazyInit()

	// 须是手机浏览器的标识, 否则可能抓不到数据
	si.Client.SetUserAgent("Mozilla/5.0 (Linux; Android 7.0; HUAWEI NXT-AL10 Build/HUAWEINXT-AL10) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.137 Mobile Safari/537.36")

	body, err := si.Client.Fetch("GET", si.SharedURL, nil, nil)
	if err != nil {
		return err
	}

	rawYunData := YunDataExp.FindSubmatch(body)
	if len(rawYunData) < 2 {
		// 检测是否需要提取密码
		if bytes.Contains(body, []byte("请输入提取密码")) {
			return fmt.Errorf("需要输入提取密码")
		}
		return fmt.Errorf("分享页数据解析失败")
	}

	err = jsoniter.Unmarshal(rawYunData[1], si)
	if err != nil {
		return fmt.Errorf("分享页, json数据解析失败, %s", err)
	}

	if si.UK == 0 || si.ShareID == 0 {
		return fmt.Errorf("分享页, json数据解析失败, 未找到 shareid 或 uk 值")
	}
	return nil
}

// FileDirectory 文件和目录的信息
type FileDirectory struct {
	FsID     int64  `json:"fs_id"`           // fs_id
	Path     string `json:"path"`            // 路径
	Filename string `json:"server_filename"` // 文件名 或 目录名
	Ctime    int64  `json:"server_ctime"`    // 创建日期
	Mtime    int64  `json:"server_mtime"`    // 修改日期
	MD5      string `json:"md5"`             // md5 值
	Size     int64  `json:"size"`            // 文件大小 (目录为0)
	Isdir    int    `json:"isdir"`           // 是否为目录
	Dlink    string `json:"dlink"`           //下载直链
}

// fileDirectoryString 文件和目录的信息, 字段类型全为 string
type fileDirectoryString struct {
	FsID     string `json:"fs_id"`           // fs_id
	Path     string `json:"path"`            // 路径
	Filename string `json:"server_filename"` // 文件名 或 目录名
	Ctime    string `json:"server_ctime"`    // 创建日期
	Mtime    string `json:"server_mtime"`    // 修改日期
	MD5      string `json:"md5"`             // md5 值
	Size     string `json:"size"`            // 文件大小 (目录为0)
	Isdir    string `json:"isdir"`           // 是否为目录
	Dlink    string `json:"dlink"`           // 下载链接
}

func (fdss *fileDirectoryString) convert() *FileDirectory {
	return &FileDirectory{
		FsID:     MustParseInt64(fdss.FsID),
		Path:     fdss.Path,
		Filename: fdss.Filename,
		Ctime:    MustParseInt64(fdss.Ctime),
		Mtime:    MustParseInt64(fdss.Mtime),
		MD5:      fdss.MD5,
		Size:     MustParseInt64(fdss.Size),
		Isdir:    MustParseInt(fdss.Isdir),
		Dlink:    fdss.Dlink,
	}
}

// List 获取文件列表, subDir 为相对于分享目录的目录
func (si *SharedInfo) List(subDir string) (fds []*FileDirectory, err error) {
	if !si.inited() {
		return nil, ErrSharedInfoNotInit
	}

	si.lazyInit()
	si.signature()
	var (
		isRoot     = 0
		escapedDir string
	)

	cleanedSubDir := path.Clean(subDir)
	if cleanedSubDir == "." || cleanedSubDir == "/" {
		isRoot = 1
	} else {
		dir := path.Clean(si.RootSharePath + "/" + subDir)
		escapedDir = url.PathEscape(dir)
	}

	listURL := fmt.Sprintf(
		"https://pan.baidu.com/share/list?shareid=%d&uk=%d&root=%d&dir=%s&sign=%x&timestamp=%d&devuid=&clienttype=1&channel=android_7.0&version=8.2.0",
		si.ShareID, si.UK,
		isRoot, escapedDir,
		si.Sign, si.Timestamp,
	)

	body, err := si.Client.Fetch("GET", listURL, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("获取文件列表网络错误, %s", err)
	}

	var errInfo = &RemoteErrInfo{}
	if isRoot != 0 { // 根目录
		jsonData := struct {
			*RemoteErrInfo
			List []*fileDirectoryString `json:"list"`
		}{
			RemoteErrInfo: errInfo,
		}

		err = jsoniter.Unmarshal(body, &jsonData)
		if err == nil {
			fds = make([]*FileDirectory, len(jsonData.List))
			for k, info := range jsonData.List {
				fds[k] = info.convert()
			}
		}
	} else {
		jsonData := struct {
			*RemoteErrInfo
			List []*FileDirectory `json:"list"`
		}{
			RemoteErrInfo: errInfo,
		}

		err = jsoniter.Unmarshal(body, &jsonData)
		if err == nil {
			fds = jsonData.List
		}
	}

	if err != nil {
		return nil, fmt.Errorf("获取文件列表, json 数据解析失败, %s", err)
	}

	if errInfo.ErrNo != 0 {
		return nil, errInfo
	}

	return fds, nil
}

// Meta 获取文件/目录元信息, filePath 为相对于分享目录的目录
func (si *SharedInfo) Meta(filePath string) (fd *FileDirectory, err error) {
	cleanedPath := path.Clean(filePath)

	dir, fileName := path.Split(cleanedPath)

	dirInfo, err := si.List(dir)
	if err != nil {
		return nil, err
	}

	for k := range dirInfo {
		if strings.Compare(dirInfo[k].Filename, fileName) == 0 {
			return dirInfo[k], nil
		}
	}

	return nil, fmt.Errorf("未匹配到文件路径")
}
