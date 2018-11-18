// Package pcsconfig 配置包
package pcsconfig

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/baidupcs/dlinkclient"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/json-iterator/go"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

const (
	// EnvConfigDir 配置路径环境变量
	EnvConfigDir = "BAIDUPCS_GO_CONFIG_DIR"
	// ConfigName 配置文件名
	ConfigName = "pcs_config.json"
)

var (
	pcsConfigVerbose = pcsverbose.New("PCSCONFIG")
	configFilePath   = filepath.Join(GetConfigDir(), ConfigName)

	// Config 配置信息, 由外部调用
	Config = NewConfig(configFilePath)
)

// PCSConfig 配置详情
type PCSConfig struct {
	baiduActiveUID    uint64
	baiduUserList     BaiduUserList
	appID             int    // appid
	cacheSize         int    // 下载缓存
	maxParallel       int    // 最大下载并发量
	maxUploadParallel int    // 最大上传并发量
	maxDownloadLoad   int    // 同时进行下载文件的最大数量
	userAgent         string // 浏览器标识
	saveDir           string // 下载储存路径
	enableHTTPS       bool   // 启用https
	proxy             string // 代理
	localAddrs        string // 本地网卡地址

	configFilePath string
	configFile     *os.File
	fileMu         sync.Mutex
	activeUser     *Baidu
	pcs            *baidupcs.BaiduPCS
	dc             *dlinkclient.DlinkClient
}

// NewConfig 返回 PCSConfig 指针对象
func NewConfig(configFilePath string) *PCSConfig {
	c := &PCSConfig{
		configFilePath: configFilePath,
	}
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

	c.initDefaultConfig()
	err := c.loadConfigFromFile()
	if err != nil {
		return err
	}

	// 载入配置
	// 如果 activeUser 已初始化, 则跳过
	if c.activeUser != nil && c.activeUser.UID == c.baiduActiveUID {
		return nil
	}

	c.activeUser, err = c.GetBaiduUser(&BaiduBase{
		UID: c.baiduActiveUID,
	})
	if err != nil {
		return err
	}
	c.pcs = c.activeUser.BaiduPCS()

	// 设置全局代理
	requester.SetGlobalProxy(c.proxy)
	// 设置本地网卡地址
	requester.SetLocalTCPAddrList(strings.Split(c.localAddrs, ",")...)

	return nil
}

// lazyOpenConfigFile 打开配置文件
func (c *PCSConfig) lazyOpenConfigFile() (err error) {
	if c.configFile != nil {
		return nil
	}

	c.fileMu.Lock()
	os.MkdirAll(filepath.Dir(c.configFilePath), 0700)
	c.configFile, err = os.OpenFile(c.configFilePath, os.O_CREATE|os.O_RDWR, 0600)
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

	// 未初始化
	info, err := c.configFile.Stat()
	if err != nil {
		return err
	}

	if info.Size() == 0 {
		err = c.Save()
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

func (c *PCSConfig) initDefaultConfig() {
	c.appID = 266719
	c.cacheSize = 30000
	c.maxParallel = 100
	c.maxUploadParallel = 10
	c.maxDownloadLoad = 1
	c.userAgent = "netdisk;8.3.1;android-android"

	// 设置默认的下载路径
	switch runtime.GOOS {
	case "windows":
		c.saveDir = pcsutil.ExecutablePathJoin("Downloads")
	case "android":
		// TODO: 获取完整的的下载路径
		c.saveDir = "/sdcard/Download"
	default:
		dataPath, ok := os.LookupEnv("HOME")
		if !ok {
			pcsConfigVerbose.Warn("Environment HOME not set")
			c.saveDir = pcsutil.ExecutablePathJoin("Downloads")
		} else {
			c.saveDir = filepath.Join(dataPath, "Downloads")
		}
	}
}

// GetConfigDir 获取配置路径
func GetConfigDir() string {
	// 从环境变量读取
	configDir, ok := os.LookupEnv(EnvConfigDir)
	if ok {
		if filepath.IsAbs(configDir) {
			return configDir
		}
		// 如果不是绝对路径, 从程序目录寻找
		return pcsutil.ExecutablePathJoin(configDir)
	}

	// 使用旧版
	// 如果旧版的配置文件存在, 则使用旧版
	oldConfigDir := pcsutil.ExecutablePath()
	_, err := os.Stat(filepath.Join(oldConfigDir, ConfigName))
	if err == nil {
		return oldConfigDir
	}

	switch runtime.GOOS {
	case "windows":
		dataPath, ok := os.LookupEnv("APPDATA")
		if !ok {
			pcsConfigVerbose.Warn("Environment APPDATA not set")
			return oldConfigDir
		}
		return filepath.Join(dataPath, "BaiduPCS-Go")
	default:
		dataPath, ok := os.LookupEnv("HOME")
		if !ok {
			pcsConfigVerbose.Warn("Environment HOME not set")
			return oldConfigDir
		}
		configDir = filepath.Join(dataPath, ".config", "BaiduPCS-Go")

		// 检测是否可写
		err = os.MkdirAll(configDir, 0700)
		if err != nil {
			pcsConfigVerbose.Warnf("check config dir error: %s\n", err)
			return oldConfigDir
		}
		return configDir
	}
}

func (c *PCSConfig) fix() {
	if c.cacheSize < 1024 {
		c.cacheSize = 1024
	}
	if c.maxParallel < 1 {
		c.maxParallel = 1
	}
	if c.maxUploadParallel < 1 {
		c.maxUploadParallel = 1
	}
	if c.maxDownloadLoad < 1 {
		c.maxDownloadLoad = 1
	}
}
