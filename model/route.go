package model

import (
	"net"
	"sort"
)

const DEFAULT_IF_NAME = "DEFAULT"

type CustomRoute struct {
	// 网卡名称
	ifName string
	//  子网  例如 0.0.0.0/0
	Subnet *net.IPNet
	// ip 地址,流量出口
	Ip net.IP
	// 优先级
	Metric int
	//
	IfIndex uint32
	// 匹配位数
	BitCount int
}

func (r *CustomRoute) GetIfName() string {
	return r.ifName
}

// sort.Interface 接口实现
type SortRoutes []*CustomRoute

func (a SortRoutes) Len() int {
	return len(a)
}

func (a SortRoutes) Less(i, j int) bool {
	if a[i].BitCount == a[j].BitCount {
		return a[i].Metric < a[j].Metric
	} else {
		return a[i].BitCount > a[j].BitCount
	}
}

func (a SortRoutes) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// 选择路由
func FindContainRoute(destIp net.IP, routes []*CustomRoute) *CustomRoute {
	if routes == nil || len(routes) == 0 {
		return getDefaultRoute()
	} else {
		customRoutes := make([]*CustomRoute, 0)
		for i := range routes {
			if routes[i].Subnet.Contains(destIp) {
				customRoutes = append(customRoutes, routes[i])
			}
		}

		if len(customRoutes) == 0 {
			return getDefaultRoute()
		} else {
			sort.Sort(SortRoutes(customRoutes))
			return customRoutes[0]
		}

	}
}

func getDefaultRoute() *CustomRoute {
	_, ipNet, _ := net.ParseCIDR("0.0.0.0/0")
	return &CustomRoute{
		ifName: DEFAULT_IF_NAME,
		Subnet: ipNet,
		Ip:     OutBoundIpv4,
	}
}
