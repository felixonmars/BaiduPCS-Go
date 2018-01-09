package pcsconfig

import (
	"encoding/json"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/util"
	"io/ioutil"
)

var (
	// Config 配置信息, 由外部调用
	Config = NewConfig()

	// ActiveBaiduUser 当前百度帐号
	ActiveBaiduUser = new(Baidu)

	configFileName = pcsutil.ExecutablePathJoin("pcs_config.json")
)

// PCSConfig 配置详情
type PCSConfig struct {
	BaiduActiveUID uint64   `json:"baidu_active_uid"`
	BaiduUserList  []*Baidu `json:"baidu_user_list"`

	MaxParallel int    `json:"max_parallel"` // 最大下载并发量
	SaveDir     string `json:"savedir"`      // 下载储存路径
}

// NewConfig 返回 PCSConfig 指针对象
func NewConfig() *PCSConfig {
	return &PCSConfig{
		BaiduActiveUID: 0,
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
			fmt.Println(err)
		}
	}

	UpdateActiveBaiduUser()
}

func loadConfig() error {
	data, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, Config)
	if err != nil {
		return err
	}

	// 下载目录为空处理, 旧版本兼容
	if Config.SaveDir == "" || Config.SaveDir == "download" {
		Config.SaveDir = pcsutil.ExecutablePathJoin("download")
	}

	return nil
}

// Reload 从配置文件重载更新 Config
func Reload() error {
	err := loadConfig()
	if err != nil {
		return err
	}

	// 更新 当前百度帐号
	return UpdateActiveBaiduUser()
}

// Save 保存配置信息到配置文件, 并重载配置
func (c *PCSConfig) Save() error {
	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(configFileName, data, 0666)
	if err != nil {
		return err
	}

	return Reload()
}

// UpdateActiveBaiduUser 更新 当前百度帐号
func UpdateActiveBaiduUser() error {
	baidu, err := Config.GetBaiduUserByUID(Config.BaiduActiveUID)
	if err != nil {
		return err
	}

	ActiveBaiduUser = baidu
	return nil
}
