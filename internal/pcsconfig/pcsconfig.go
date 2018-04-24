// Package pcsconfig 配置包
package pcsconfig

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/json-iterator/go"
	"io/ioutil"
)

var (
	// Config 配置信息, 由外部调用
	Config = NewConfig()

	configFileName = pcsutil.ExecutablePathJoin("pcs_config.json")

	defaultAppID = 260149
)

// PCSConfig 配置详情
type PCSConfig struct {
	BaiduActiveUID uint64        `json:"baidu_active_uid"`
	BaiduUserList  BaiduUserList `json:"baidu_user_list"`

	AppID int `json:"appid"` // appid

	CacheSize   int `json:"cache_size"`   // 下载缓存
	MaxParallel int `json:"max_parallel"` // 最大下载并发量

	UserAgent   string `json:"user_agent"`   // 浏览器标识
	SaveDir     string `json:"savedir"`      // 下载储存路径
	EnableHTTPS bool   `json:"enable_https"` // 启用https
}

// NewConfig 返回 PCSConfig 指针对象
func NewConfig() *PCSConfig {
	return &PCSConfig{
		BaiduActiveUID: 0,
		AppID:          defaultAppID,
		CacheSize:      30000,
		MaxParallel:    100,
		SaveDir:        pcsutil.ExecutablePathJoin("download"),
	}
}

// Init 初始化配置
func Init() {
	// 检查配置
	err := loadConfig()
	if err != nil {
		fmt.Printf("错误: %s, 自动初始化配置文件\n", err)

		err = Config.Save()
		if err != nil {
			panic(err)
		}
	}
}

// Reload 从配置文件重载更新 Config
func Reload() error {
	err := loadConfig()
	if err != nil {
		return err
	}

	return nil
}

func loadConfig() error {
	data, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return err
	}

	err = jsoniter.Unmarshal(data, Config)
	if err != nil {
		return err
	}

	// 下载目录为空处理, 旧版本兼容
	if Config.SaveDir == "" || Config.SaveDir == "download" {
		Config.SaveDir = pcsutil.ExecutablePathJoin("download")
	}

	// 设置浏览器标识
	if Config.UserAgent != "" {
		setUserAgent(Config.UserAgent)
	}

	return nil
}

// Save 保存配置信息到配置文件
func (c *PCSConfig) Save() error {
	// 检测配置项是否合法, 不合法则无法保存
	err := c.CheckValid()
	if err != nil {
		return err
	}

	data, err := jsoniter.MarshalIndent(c, "", " ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(configFileName, data, 0666)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %s", err)
	}

	return nil
}
