package pcsconfig

import (
	"bytes"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/downloader"
	"os"
	"path/filepath"
	"strings"
)

// GetBaiduUserByUID 通过 百度uid 获取 Baidu 指针对象
func (c *PCSConfig) GetBaiduUserByUID(uid uint64) (*Baidu, error) {
	if c.BaiduActiveUID == 0 {
		return nil, fmt.Errorf("初始化状态, 请设置百度帐号")
	}

	for k := range c.BaiduUserList {
		if uid == c.BaiduUserList[k].UID {
			return c.BaiduUserList[k], nil
		}
	}
	return nil, fmt.Errorf("未找到uid 为 %d 的百度帐号", c.BaiduActiveUID)
}

// GetAllBaiduUser 返回所有已登录百度帐号
func (c *PCSConfig) GetAllBaiduUser() string {
	var s bytes.Buffer
	s.WriteString("\nindex\t\tuid\t用户名\n")
	s.WriteString(strings.Repeat("-", 50) + "\n")

	for k := range c.BaiduUserList {
		s.WriteString(fmt.Sprintf("%4d|", k) + "\t" + fmt.Sprintf("%11d|", c.BaiduUserList[k].UID) + "\t" + c.BaiduUserList[k].Name + "\n")
	}
	s.WriteRune('\n')
	return s.String()
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
	dirStr := fmt.Sprintf("%s/%d_%s%s/.",
		Config.SaveDir,
		ActiveBaiduUser.UID,
		ActiveBaiduUser.Name,
		path,
	)

	dir, err := filepath.Abs(dirStr)
	if err != nil {
		dir = filepath.Dir(dirStr + "/.")
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
