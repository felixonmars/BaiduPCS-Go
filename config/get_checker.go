package pcsconfig

import (
	"bytes"
	"fmt"
)

func (c *PCSConfig) GetActiveBDUSS() string {
	for k := range c.BaiduUserList {
		if c.BaiduActiveUID == c.BaiduUserList[k].UID {
			return c.BaiduUserList[k].BDUSS
		}
	}
	return ""
}

func (c *PCSConfig) GetActiveUserName() string {
	for k := range c.BaiduUserList {
		if c.BaiduActiveUID == c.BaiduUserList[k].UID {
			return c.BaiduUserList[k].Name
		}
	}
	return ""
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
