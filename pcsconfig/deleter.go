package pcsconfig

import (
	"fmt"
)

// DeleteBaiduUserByUID 通过uid删除百度帐号
func (c *PCSConfig) DeleteBaiduUserByUID(uid uint64) error {
	for k := range c.BaiduUserList {
		if c.BaiduUserList[k].UID == uid {
			c.BaiduUserList = append(c.BaiduUserList[:k], c.BaiduUserList[k+1:]...)

			// 修改 正在使用的 百度帐号
			// 如果要删除的帐号为当前登录的帐号, 则设置当前登录帐号为列表中第一个帐号
			if c.BaiduActiveUID == uid {
				if len(c.BaiduUserList) != 0 {
					c.BaiduActiveUID = c.BaiduUserList[0].UID
				} else {
					c.BaiduActiveUID = 0
				}
			}

			return c.Save()
		}
	}

	return fmt.Errorf("删除百度帐号失败, uid 不存在")
}
