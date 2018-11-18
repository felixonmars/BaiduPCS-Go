package requester

import (
	"context"
	"crypto/tls"
	"errors"
	mathrand "math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var (
	localTCPAddrList = []*net.TCPAddr{}

	// ProxyAddr 代理地址
	ProxyAddr string

	// ErrProxyAddrEmpty 代理地址为空
	ErrProxyAddrEmpty = errors.New("proxy addr is empty")
)

// SetLocalTCPAddrList 设置网卡地址
func SetLocalTCPAddrList(ips ...string) {
	list := make([]*net.TCPAddr, 0, len(ips))
	for k := range ips {
		p := net.ParseIP(ips[k])
		if p == nil {
			continue
		}

		list = append(list, &net.TCPAddr{
			IP: p,
		})
	}
	localTCPAddrList = list
}

func proxyFunc(req *http.Request) (*url.URL, error) {
	u, err := checkProxyAddr(ProxyAddr)
	if err != nil {
		return http.ProxyFromEnvironment(req)
	}

	return u, err
}

func getLocalTCPAddr() *net.TCPAddr {
	if len(localTCPAddrList) == 0 {
		return nil
	}
	i := mathrand.Intn(len(localTCPAddrList))
	return localTCPAddrList[i]
}

func getDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		LocalAddr: getLocalTCPAddr(),
		DualStack: true,
	}
}

func checkProxyAddr(proxyAddr string) (u *url.URL, err error) {
	if proxyAddr == "" {
		return nil, ErrProxyAddrEmpty
	}

	host, port, err := net.SplitHostPort(proxyAddr)
	if err == nil {
		u = &url.URL{
			Host: net.JoinHostPort(host, port),
		}
		return
	}

	u, err = url.Parse(proxyAddr)
	if err == nil {
		return
	}

	return
}

// SetGlobalProxy 设置代理
func SetGlobalProxy(proxyAddr string) {
	ProxyAddr = proxyAddr
}

func getServerName(address string) string {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return address
	}
	return host
}

func resolveTCP(ctx context.Context, address string) (tcpaddr *net.TCPAddr, err error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return
	}

	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		return
	}

	return &net.TCPAddr{
		IP:   addrs[0].IP,
		Port: p,
		Zone: addrs[0].Zone,
	}, nil
}

func dialContext(ctx context.Context, network, address string) (conn net.Conn, err error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
		var (
			ta = TCPAddrCache.Get(address)
		)

		// 检测缓存
		if ta != nil {
			return net.DialTCP(network, getLocalTCPAddr(), ta)
		}

		// Resolve TCP address
		ta, err = resolveTCP(ctx, address)

		if err != nil {
			return nil, err
		}

		// 加入缓存
		TCPAddrCache.Set(address, ta)
		return net.DialTCP(network, getLocalTCPAddr(), ta)
	}

	// 非 tcp 请求
	conn, err = getDialer().DialContext(ctx, network, address)
	return
}

func dial(network, address string) (conn net.Conn, err error) {
	return dialContext(context.Background(), network, address)
}

func (h *HTTPClient) dialTLSFunc() func(network, address string) (tlsConn net.Conn, err error) {
	return func(network, address string) (tlsConn net.Conn, err error) {
		conn, err := dialContext(context.Background(), network, address)
		if err != nil {
			return nil, err
		}

		return tls.Client(conn, &tls.Config{
			ServerName:         getServerName(address),
			InsecureSkipVerify: !h.https,
		}), nil
	}
}
