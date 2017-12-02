package pcsconfig

import (
	"fmt"
)

// SetBDUSS 设置百度 bduss 并保存
func (c *PCSConfig) SetBDUSS(bduss string) (username string, err error) {
	b, err := NewWithBDUSS(bduss)
	if err != nil {
		return "", err
	}
	if c.CheckUIDExist(b.UID) {
		return "", fmt.Errorf("登录失败, 用户 %s 已存在", b.Name)
	}
	c.BaiduUserList = append(c.BaiduUserList, b)
	c.BaiduActiveUID = b.UID
	return b.Name, c.Save()
}
