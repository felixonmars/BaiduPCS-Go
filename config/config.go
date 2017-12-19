package pcsconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var (
	// Config 配置信息, 由外部调用
	Config = new(PCSConfig)

	// ActiveBaiduUser 当前百度帐号
	ActiveBaiduUser *Baidu

	configFileName = "pcs_config.json"
)

// PCSConfig 配置详情
type PCSConfig struct {
	BaiduActiveUID uint64   `json:"baidu_active_uid"`
	BaiduUserList  []*Baidu `json:"baidu_user_list"`

	MaxParallel int    `json:"max_parallel"` // 最大下载并发量
	SaveDir     string `json:"savedir"`      // 下载储存路径
}

// Init 初始化配置
func Init() {
	// 检查配置
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("错误: %s, 自动初始化配置文件\n", err)

		cfg = &PCSConfig{
			BaiduActiveUID: 0,
			MaxParallel:    100,
			SaveDir:        "download",
		}

		err = cfg.Save()
		if err != nil {
			fmt.Println(err)
		}
	}
	Config = cfg

	if UpdateActiveBaiduUser() != nil {
		ActiveBaiduUser = new(Baidu)
	}
}

func loadConfig() (*PCSConfig, error) {
	data, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return nil, err
	}
	conf := new(PCSConfig)
	err = json.Unmarshal(data, conf)
	if err != nil {
		return nil, err
	}

	// 下载目录为空处理
	if conf.SaveDir == "" {
		conf.SaveDir = "download"
	}

	return conf, nil
}

// Reload 从配置文件重载更新 Config
func Reload() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	Config = cfg

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
	if err == nil {
		ActiveBaiduUser = baidu
		return nil
	}
	return err
}
