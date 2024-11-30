package model

import (
	"net"
)

// 全局配置
var (
	TunName          string
	TunIp            string
	DNS1             string
	DNS2             string
	OutBoundIpv4     net.IP // 网络出口
	OutBoundIpv6     net.IP // 网络出口
	MTU              int    // Maximum Transmission Unit，最大传输单元  最小46byte，最大一般1500byte
	ContextConfigImp ContextConfig
	Routes           []*CustomRoute
	// 路由更新通知
	ReaRouteUpdateChan chan bool //
)

func init() {
	TunName = "anApp"
	TunIp = "10.0.0.2/24"
	DNS1 = "10.0.0.3"
	DNS2 = "10.0.0.4"
	OutBoundIpv4 = net.IP{1, 1, 1, 1}
	MTU = 1500 //
	ReaRouteUpdateChan = make(chan bool)
}

func GetReleaseUdpPort() (int, error) {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 0,
	})
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.LocalAddr().(*net.UDPAddr)
	return addr.Port, nil
}

func SetContextConfig(c ContextConfig) {
	ContextConfigImp = c
}
