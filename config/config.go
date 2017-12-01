package pcsconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var (
	// Config 配置信息, 由外部调用
	Config = new(PCSConfig)

	configFileName = "pcs_config.json"
)

// PCSConfig 配置详情
type PCSConfig struct {
	BaiduActiveUID uint64  `json:"baidu_active_uid"`
	BaiduUserList  []Baidu `json:"baidu_user_list"`
	Workdir        string  `json:"workdir"`
	MaxParallel    int     `json:"max_parallel"`
}

func init() {
	// 检查配置
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("错误: %s, 自动初始化配置文件\n", err)
		cfg, err = initConfig()
		if err != nil {
			fmt.Println(err)
		}
	}
	Config = cfg
}

func initConfig() (*PCSConfig, error) {
	cfg := &PCSConfig{
		Workdir:     "/",
		MaxParallel: 100,
	}
	return cfg, cfg.Save()
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
	return conf, nil
}

// Reload 从配置文件重载更新 Config
func Reload() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	Config = cfg
	return nil
}

// Save 保存配置信息到配置文件
func (c *PCSConfig) Save() error {
	data, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configFileName, data, 0666)
}
