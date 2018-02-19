package requester

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcstable"
	"net"
	"os"
	"sync"
	"time"
)

var (
	// TCPAddrCache tcp地址缓存
	TCPAddrCache = tcpAddrCache{
		ta:       sync.Map{},
		lifeTime: 1 * time.Minute,
	}
)

type tcpAddrCache struct {
	ta       sync.Map
	lifeTime time.Duration // 生命周期
}

func init() {
	TCPAddrCache.GC()
}

// Set 设置
func (tac *tcpAddrCache) Set(address string, ta *net.TCPAddr) {
	if tac.Existed(address) {
		return
	}

	tac.ta.Store(address, ta)
}

// Existed 检测存在
func (tac *tcpAddrCache) Existed(address string) bool {
	v, existed := tac.ta.Load(address)
	if existed && v == nil {
		return false
	}

	return existed
}

// Get 获取
func (tac *tcpAddrCache) Get(address string) *net.TCPAddr {
	if tac.Existed(address) {
		value, _ := tac.ta.Load(address)
		return value.(*net.TCPAddr)
	}

	return nil
}

// SetLifeTime 设置生命周期
func (tac *tcpAddrCache) SetLifeTime(t time.Duration) {
	tac.lifeTime = t
}

// GC 缓存回收
func (tac *tcpAddrCache) GC() {
	go func() {
		ticker := time.NewTicker(tac.lifeTime)
		for {
			select {
			case <-ticker.C:
				tac.DelAll()
			}
		}
	}()
}

// Del 删除缓存
func (tac *tcpAddrCache) Del(address string) {
	tac.ta.Delete(address)
}

// DelAll 清空缓存
func (tac *tcpAddrCache) DelAll() {
	tac.ta.Range(func(address, value interface{}) bool {
		tac.Del(address.(string))
		return true
	})
}

// PrintAll 输出全部
func (tac *tcpAddrCache) PrintAll() {
	tb := pcstable.NewTable(os.Stdout)
	tb.SetHeader([]string{"address", "tcpaddr"})
	tac.ta.Range(func(address, tcpaddr interface{}) bool {
		tb.Append([]string{address.(string), fmt.Sprint(tcpaddr.(*net.TCPAddr))})
		return true
	})
	tb.Render()
}
