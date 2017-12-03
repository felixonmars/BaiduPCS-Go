package pcsconfig

import (
	"bytes"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/downloader"
	"os"
	"path/filepath"
)

func (c *PCSConfig) GetBaiduUserByUID(uid uint64) (*Baidu, error) {
	for k := range c.BaiduUserList {
		if uid == c.BaiduUserList[k].UID {
			return c.BaiduUserList[k], nil
		}
	}
	return nil, fmt.Errorf("未找到百度帐号")
}

func (c *PCSConfig) GetAllBaiduUser() string {
	var s bytes.Buffer
	s.WriteString("\nindex\t\tuid\t用户名\n")

	for k := range c.BaiduUserList {
		s.WriteString(fmt.Sprintf("%4d", k) + "\t" + fmt.Sprintf("%11d", c.BaiduUserList[k].UID) + "\t" + c.BaiduUserList[k].Name + "\n")
	}
	s.WriteString("\n")
	return s.String()
}

func (c *PCSConfig) CheckUIDExist(uid uint64) bool {
	for k := range c.BaiduUserList {
		if uid == c.BaiduUserList[k].UID {
			return true
		}
	}
	return false
}

// GetSavePath 根据提供的网盘文件路径 path, 返回本地储存路径
func GetSavePath(path string) string {
	return filepath.Dir(fmt.Sprintf("%s/%d_%s%s/..",
		SaveDir,
		ActiveBaiduUser.UID,
		ActiveBaiduUser.Name,
		path,
	))
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
