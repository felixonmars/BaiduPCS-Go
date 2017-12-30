package pcsutil

import (
	"net"
)

// ListAddresses 列出本地可用的 IP 地址
func ListAddresses() (addresses []string) {
	ifaces, _ := net.Interfaces()
	addresses = make([]string, 0, len(ifaces))
	for k := range ifaces[:] {
		ifAddrs, _ := ifaces[k].Addrs()
		for l := range ifAddrs[:] {
			switch v := ifAddrs[l].(type) {
			case *net.IPNet:
				addresses = append(addresses, v.IP.String())
			case *net.IPAddr:
				addresses = append(addresses, v.IP.String())
			}
		}
	}
	return
}
