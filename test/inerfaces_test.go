package test

import (
	"net"
	"testing"
)

// 获取有效路由
func TestInterfaces(t *testing.T) {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("无法获取接口列表")
	}
	for _, intf := range interfaces {
		addrs, err := intf.Addrs()
		if err != nil {
			panic("无法获取地址列表")
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}
			//if ipNet.Contains(ip) {
			//	return ipNet
			//}
		}
	}
}
