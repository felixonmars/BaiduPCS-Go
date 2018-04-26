// Package pcsconfig 配置包
package pcsconfig

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/json-iterator/go"
	"os"
	"sync"
	"unsafe"
)

var (
	configFilePath = pcsutil.ExecutablePathJoin("pcs_config.json")

	// Config 配置信息, 由外部调用
	Config = NewConfig(configFilePath)

	defaultAppID = 260149
)

// PCSConfig 配置详情
type PCSConfig struct {
	baiduActiveUID uint64
	baiduUserList  BaiduUserList
	appID          int    // appid
	cacheSize      int    // 下载缓存
	maxParallel    int    // 最大下载并发量
	userAgent      string // 浏览器标识
	saveDir        string // 下载储存路径
	enableHTTPS    bool   // 启用https

	configFilePath string
	configFile     *os.File
	fileMu         sync.Mutex
	activeUser     *Baidu
	pcs            *baidupcs.BaiduPCS
}

// NewConfig 返回 PCSConfig 指针对象
func NewConfig(configFilePath string) *PCSConfig {
	c := &PCSConfig{
		configFilePath: configFilePath,
	}
	c.defaultConfig()
	return c
}

// Init 初始化配置
func (c *PCSConfig) Init() error {
	return c.init()
}

// Reload 从文件重载配置
func (c *PCSConfig) Reload() error {
	return c.init()
}

// Close 关闭配置文件
func (c *PCSConfig) Close() error {
	if c.configFile != nil {
		err := c.configFile.Close()
		c.configFile = nil
		return err
	}
	return nil
}

// Save 保存配置信息到配置文件
func (c *PCSConfig) Save() error {
	// 检测配置项是否合法, 不合法则自动修复
	c.fix()

	err := c.lazyOpenConfigFile()
	if err != nil {
		return err
	}

	c.fileMu.Lock()
	defer c.fileMu.Unlock()

	data, err := jsoniter.MarshalIndent((*pcsConfigJSONExport)(unsafe.Pointer(c)), "", " ")
	if err != nil {
		// json数据生成失败
		panic(err)
	}

	// 减掉多余的部分
	err = c.configFile.Truncate(int64(len(data)))
	if err != nil {
		return err
	}

	_, err = c.configFile.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	_, err = c.configFile.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func (c *PCSConfig) init() error {
	if c.configFilePath == "" {
		return ErrConfigFileNotExist
	}
	err := c.loadConfigFromFile()
	if err != nil {
		return err
	}

	// 载入配置
	c.activeUser, err = c.GetBaiduUser(&BaiduBase{
		UID: c.baiduActiveUID,
	})
	if err != nil {
		return err
	}
	c.pcs = c.activeUser.BaiduPCS()

	return nil
}

// lazyOpenConfigFile 打开配置文件
func (c *PCSConfig) lazyOpenConfigFile() (err error) {
	if c.configFile != nil {
		return nil
	}

	c.fileMu.Lock()
	c.configFile, err = os.OpenFile(c.configFilePath, os.O_CREATE|os.O_RDWR, 0640)
	c.fileMu.Unlock()

	if err != nil {
		if os.IsPermission(err) {
			return ErrConfigFileNoPermission
		}
		if os.IsExist(err) {
			return ErrConfigFileNotExist
		}
		return err
	}
	return nil
}

// loadConfigFromFile 载入配置
func (c *PCSConfig) loadConfigFromFile() (err error) {
	err = c.lazyOpenConfigFile()
	if err != nil {
		return err
	}

	c.fileMu.Lock()
	defer c.fileMu.Unlock()

	_, err = c.configFile.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	d := jsoniter.NewDecoder(c.configFile)
	err = d.Decode((*pcsConfigJSONExport)(unsafe.Pointer(c)))
	if err != nil {
		return ErrConfigContentsParseError
	}
	return nil
}

func (c *PCSConfig) defaultConfig() {
	if c.appID == 0 {
		c.appID = defaultAppID
	}
	if c.cacheSize == 0 {
		c.cacheSize = 30000
	}
	if c.maxParallel == 0 {
		c.maxParallel = 100
	}
	if c.saveDir == "" {
		c.saveDir = pcsutil.ExecutablePathJoin("download")
	}
}

func (c *PCSConfig) fix() {
	if c.cacheSize < 1024 {
		c.cacheSize = 1024
	}
	if c.maxParallel < 1 {
		c.maxParallel = 1
	}
}
