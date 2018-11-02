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

// tcpAddrCache tcp地址缓存, 即dns解析后的ip地址
type tcpAddrCache struct {
	ta        sync.Map
	lifeTime  time.Duration // 生命周期
	gcStarted bool
}

// tcpAddrItem tcpAddrCache中缓存的带有超时的TcpAddr
type tcpAddrItem struct {
	ta     *net.TCPAddr
	expire <-chan time.Time // 缓存项是否过期
}

// Set 设置
func (tac *tcpAddrCache) Set(address string, ta *net.TCPAddr) {
	item := &tcpAddrItem{ta, time.After(tac.lifeTime)}
	tac.ta.Store(address, item)
}

// Get 获取
func (tac *tcpAddrCache) Get(address string) *net.TCPAddr {
	value, ok := tac.ta.Load(address)
	if !ok {
		return nil
	}

	return value.(*tcpAddrItem).ta
}

// SetLifeTime 设置生命周期
// 重新设定生命周期将会影响所有的缓存项
func (tac *tcpAddrCache) SetLifeTime(t time.Duration) {
	tac.lifeTime = t
	tac.ta.Range(func(_, v interface{}) bool {
		item := v.(*tcpAddrItem)
		item.expire = time.After(tac.lifeTime)
		return true
	})
}

// GC 缓存回收
func (tac *tcpAddrCache) GC() {
	if tac.gcStarted {
		return
	}

	tac.gcStarted = true
	go func() {
		for {
			time.Sleep(tac.lifeTime) // 这样可以动态修改 lifetime
			tac.DelAll()
		}
	}()
}

// Del 删除缓存
func (tac *tcpAddrCache) Del(address string) {
	tac.ta.Delete(address)
}

// DelAll 清空缓存
func (tac *tcpAddrCache) DelAll() {
	tac.ta.Range(func(address, v interface{}) bool {
		item := v.(*tcpAddrItem)
		select {
		case <-item.expire: // 如果超时再删去缓存项，避免在接近lifeTime前添加的缓存项被删除
			tac.ta.Delete(address)
			return true
		default:
			return true
		}
	})
}

// PrintAll 输出全部 tcp 缓存地址
func (tac *tcpAddrCache) PrintAll() {
	tb := pcstable.NewTable(os.Stdout)
	tb.SetHeader([]string{"address", "tcpaddr"})
	tac.ta.Range(func(address, tcpaddr interface{}) bool {
		tb.Append([]string{address.(string), fmt.Sprint(tcpaddr.(*tcpAddrItem).ta)})
		return true
	})
	tb.Render()
}
