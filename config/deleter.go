package pcsconfig

import (
	"fmt"
)

// DeleteBaiduUserByUID 通过uid删除百度帐号
func (c *PCSConfig) DeleteBaiduUserByUID(uid uint64) bool {
	for k := range c.BaiduUserList {
		if c.BaiduUserList[k].UID == uid {
			c.BaiduUserList = append(c.BaiduUserList[:k], c.BaiduUserList[k+1:]...)

			// 修改 正在使用的 百度帐号
			if c.BaiduActiveUID == uid {
				if len(c.BaiduUserList) != 0 {
					c.BaiduActiveUID = c.BaiduUserList[0].UID
				} else {
					c.BaiduActiveUID = 0
				}
			}

			err := c.Save()
			if err != nil {
				fmt.Println(err)
				return false
			}
			return true
		}
	}
	return false
}
