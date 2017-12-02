package pcsconfig

import (
	"bytes"
	"fmt"
)

func (c *PCSConfig) GetBaiduUserByUID(uid uint64) (Baidu, error) {
	for k := range c.BaiduUserList {
		if uid == c.BaiduUserList[k].UID {
			return c.BaiduUserList[k], nil
		}
	}
	return Baidu{}, fmt.Errorf("未找到百度帐号")
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
