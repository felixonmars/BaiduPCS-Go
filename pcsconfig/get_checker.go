package pcsconfig

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/downloader"
	"os"
	"path/filepath"
)

// GetBaiduUserByUID 通过 百度uid 获取 Baidu 指针对象
func (c *PCSConfig) GetBaiduUserByUID(uid uint64) (*Baidu, error) {
	// 未设置任何百度帐号
	if c.BaiduActiveUID == 0 {
		return new(Baidu), nil
	}

	for k := range c.BaiduUserList {
		if uid == c.BaiduUserList[k].UID {
			return c.BaiduUserList[k], nil
		}
	}

	return nil, fmt.Errorf("未找到uid 为 %d 的百度帐号", c.BaiduActiveUID)
}

// GetActive 获取当前登录的百度帐号
func (c *PCSConfig) GetActive() (*Baidu, error) {
	return c.GetBaiduUserByUID(c.BaiduActiveUID)
}

// MustGetActive 获取当前登录的百度帐号
func (c *PCSConfig) MustGetActive() *Baidu {
	b, err := c.GetBaiduUserByUID(c.BaiduActiveUID)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return b
}

// CheckUIDExist 检查 百度uid 是否存在于已登录列表
func (c *PCSConfig) CheckUIDExist(uid uint64) bool {
	if uid == 0 {
		return false
	}
	for k := range c.BaiduUserList {
		if uid == c.BaiduUserList[k].UID {
			return true
		}
	}
	return false
}

// GetSavePath 根据提供的网盘文件路径 path, 返回本地储存路径,
// 返回绝对路径, 获取绝对路径出错时才返回相对路径...
func GetSavePath(path string) string {
	au, _ := Config.GetActive()

	dirStr := fmt.Sprintf("%s/%d_%s%s/.",
		Config.SaveDir,
		au.UID,
		au.Name,
		path,
	)

	dir, err := filepath.Abs(dirStr)
	if err != nil {
		dir = filepath.Clean(dirStr)
	}
	return dir
}

// CheckFileExist 检查本地文件是否与网盘的文件重名
func CheckFileExist(path string) bool {
	savePath := GetSavePath(path)
	if _, err := os.Stat(savePath); err == nil {
		if _, err = os.Stat(savePath + downloader.DownloadingFileSuffix); err != nil {
			return true
		}
	}
	return false
}
