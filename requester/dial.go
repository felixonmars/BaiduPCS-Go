package requester

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"
	"time"
)

var (
	dialer = &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
)

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
		// 检测缓存
		if TCPAddrCache.Existed(address) {
			return net.DialTCP(network, nil, TCPAddrCache.Get(address))
		}

		var (
			ta *net.TCPAddr
		)

		// Resolve TCP address
		ta, err = resolveTCP(ctx, address)

		if err != nil {
			return nil, err
		}

		// 加入缓存
		TCPAddrCache.Set(address, ta)
		return net.DialTCP(network, nil, ta)
	}

	// 非 tcp 请求
	conn, err = dialer.DialContext(ctx, network, address)
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
