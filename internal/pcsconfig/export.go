package pcsconfig

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"github.com/olekukonko/tablewriter"
	"os"
	"strconv"
)

// pcsConfigJSONExport 导出配置详情, 用于生成json数据
type pcsConfigJSONExport struct {
	BaiduActiveUID uint64        `json:"baidu_active_uid"`
	BaiduUserList  BaiduUserList `json:"baidu_user_list"`

	AppID int `json:"appid"` // appid

	CacheSize   int `json:"cache_size"`   // 下载缓存
	MaxParallel int `json:"max_parallel"` // 最大下载并发量

	UserAgent   string `json:"user_agent"`   // 浏览器标识
	SaveDir     string `json:"savedir"`      // 下载储存路径
	EnableHTTPS bool   `json:"enable_https"` // 启用https
}

// ActiveUser 获取当前登录的用户
func (c *PCSConfig) ActiveUser() *Baidu {
	if c.activeUser == nil {
		return &Baidu{}
	}
	return c.activeUser
}

// ActiveUserBaiduPCS 获取当前登录的用户的baidupcs.BaiduPCS
func (c *PCSConfig) ActiveUserBaiduPCS() *baidupcs.BaiduPCS {
	if c.pcs == nil {
		c.pcs = c.ActiveUser().BaiduPCS()
	}
	return c.pcs
}

// BaiduUserList 获取百度用户列表
func (c *PCSConfig) BaiduUserList() BaiduUserList {
	return c.baiduUserList
}

// NumLogins 获取登录的用户数量
func (c *PCSConfig) NumLogins() int {
	return len(c.baiduUserList)
}

// AppID 返回app_id
func (c *PCSConfig) AppID() int {
	return c.appID
}

// CacheSize 返回cache_size, 下载缓存
func (c *PCSConfig) CacheSize() int {
	return c.cacheSize
}

// MaxParallel 返回max_parallel, 下载最大并发量
func (c *PCSConfig) MaxParallel() int {
	return c.maxParallel
}

// UserAgent 返回User-Agent
func (c *PCSConfig) UserAgent() string {
	return c.userAgent
}

// SaveDir 返回下载保存路径
func (c *PCSConfig) SaveDir() string {
	return c.saveDir
}

// EnableHTTPS 返回是否启用https
func (c *PCSConfig) EnableHTTPS() bool {
	return c.enableHTTPS
}

// PrintTable 输出表格
func (c *PCSConfig) PrintTable() {
	tb := pcstable.NewTable(os.Stdout)
	tb.SetHeader([]string{"名称", "值", "建议值", "描述"})
	tb.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_LEFT})
	tb.AppendBulk([][]string{
		[]string{"appid", fmt.Sprint(c.appID), "", "百度 PCS 应用ID"},
		[]string{"enable_https", fmt.Sprint(c.enableHTTPS), "true", "启用 https"},
		[]string{"user_agent", c.userAgent, "", "浏览器标识"},
		[]string{"cache_size", strconv.Itoa(c.cacheSize), "1024 ~ 262144", "下载缓存, 如果硬盘占用高或下载速度慢, 请尝试调大此值"},
		[]string{"max_parallel", strconv.Itoa(c.maxParallel), "50 ~ 500", "下载最大并发量"},
		[]string{"savedir", c.saveDir, "", "下载文件的储存目录"},
	})
	tb.Render()
}
