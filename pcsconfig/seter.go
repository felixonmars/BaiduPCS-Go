package pcsconfig

import (
	"github.com/iikira/BaiduPCS-Go/requester"
)

// SetBDUSS 设置百度 bduss, ptoken, stoken 并保存
func (c *PCSConfig) SetBDUSS(bduss, ptoken, stoken string) (username string, err error) {
	b, err := NewUserInfoByBDUSS(bduss)
	if err != nil {
		return "", err
	}
	if c.CheckUIDExist(b.UID) {
		c.DeleteBaiduUserByUID(b.UID) // 删除旧的信息
	}

	b.PTOKEN = ptoken
	b.STOKEN = stoken

	c.BaiduUserList = append(c.BaiduUserList, b)
	c.BaiduActiveUID = b.UID
	return b.Name, c.Save()
}

func setUserAgent(ua string) {
	Config.UserAgent = ua
	requester.UserAgent = ua
}
