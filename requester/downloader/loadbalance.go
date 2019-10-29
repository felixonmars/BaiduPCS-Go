package downloader

import (
	"net/http"
	"sync/atomic"
)

// LoadBalancerResponse 负载均衡响应状态
type LoadBalancerResponse struct {
	URL     string
	Referer string
}

// LoadBalancerResponseList 负载均衡列表
type LoadBalancerResponseList struct {
	lbr    []*LoadBalancerResponse
	cursor int32
}

// NewLoadBalancerResponseList 初始化负载均衡列表
func NewLoadBalancerResponseList(lbr []*LoadBalancerResponse) *LoadBalancerResponseList {
	return &LoadBalancerResponseList{
		lbr: lbr,
	}
}

// SequentialGet 顺序获取
func (lbrl *LoadBalancerResponseList) SequentialGet() *LoadBalancerResponse {
	if len(lbrl.lbr) == 0 {
		return nil
	}

	if int(lbrl.cursor) >= len(lbrl.lbr) {
		lbrl.cursor = 0
	}

	lbr := lbrl.lbr[int(lbrl.cursor)]
	atomic.AddInt32(&lbrl.cursor, 1)
	return lbr
}

// RandomGet 随机获取
func (lbrl *LoadBalancerResponseList) RandomGet() *LoadBalancerResponse {
	return lbrl.lbr[RandomNumber(0, len(lbrl.lbr))]
}

// AddLoadBalanceServer 增加负载均衡服务器
func (der *Downloader) AddLoadBalanceServer(urls ...string) {
	der.loadBalansers = append(der.loadBalansers, urls...)
}

// ServerEqual 检测负载均衡的服务器是否一致
func ServerEqual(resp, subResp *http.Response) bool {
	if resp == nil || subResp == nil {
		return false
	}

	header, subHeader := resp.Header, subResp.Header
	if header.Get("Content-MD5") != subHeader.Get("Content-MD5") {
		return false
	}
	if header.Get("Content-Type") != subHeader.Get("Content-Type") {
		return false
	}
	if header.Get("x-bs-meta-crc32") != subHeader.Get("x-bs-meta-crc32") {
		return false
	}
	return true
}
