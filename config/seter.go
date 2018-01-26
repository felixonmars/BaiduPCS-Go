package pcsconfig

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/requester"
	"strconv"
)

// SetBDUSS 设置百度 bduss, ptoken, stoken 并保存
func (c *PCSConfig) SetBDUSS(bduss, ptoken, stoken string) (username string, err error) {
	b, err := NewWithBDUSS(bduss)
	if err != nil {
		return "", err
	}
	if c.CheckUIDExist(b.UID) {
		return "", fmt.Errorf("登录失败, 用户 %s 已存在", b.Name)
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

func (c *PCSConfig) Set(key, value string) (err error) {
	switch key {
	case "appid", "cache_size", "max_parallel":
		intVal, err := strconv.Atoi(value)
		if err != nil {
			fmt.Printf("%s 不合法, 错误: %s\n", key, err)
			return err
		}

		if intVal <= 0 {
			fmt.Printf("%s 不合法, 值应为一个正整数\n", key)
			return nil
		}

		switch key {
		case "appid":
			c.AppID = intVal
		case "cache_size":
			c.CacheSize = intVal
		case "max_parallel":
			c.MaxParallel = intVal
		}

		err = c.Save()
		if err != nil {
			fmt.Println("设置失败, 错误:", err)
			return nil
		}
		fmt.Printf("设置成功, %s -> %v\n", key, value)

	case "user_agent", "savedir":
		switch key {
		case "user_agent":
			setUserAgent(value)
		case "savedir":
			c.SaveDir = value
		}

		c.Save()
		fmt.Printf("设置成功, %s -> %v\n", key, value)

	default:
		return fmt.Errorf("未知设定值: %s\n\n", key)
	}
	return nil
}
