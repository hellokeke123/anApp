package model

import (
	netlinkImp "github.com/vishvananda/netlink"
	"log"
	"net"
)

func FindRoutes() []*CustomRoute {
	// 查main 表路由 ，这里不做复杂路由处理
	list, err := netlinkImp.RouteList(nil, netlinkImp.FAMILY_V4)
	if err != nil {
		log.Println("查询如由错误")
	}
	routes := make([]*CustomRoute, 0)

	ifc, _ := net.InterfaceByIndex(list[0].LinkIndex)

	addrs, _ := ifc.Addrs()

	var ip net.IP
	for _, addr := range addrs {
		ip = addr.(*net.IPNet).IP
		break
	}

	routes = append(routes, &CustomRoute{
		ifName: ifc.Name,
		Subnet: &DEFAULT_IPNET,
		Ip:     ip,
		// 优先级,
		Metric: list[0].Priority,
		//
		IfIndex: uint32(list[0].LinkIndex),
		// 匹配位数
		BitCount: 32,
	})

	return routes

	return nil
}

func InitRoute() {
	// 定时任务，在回调失效的情况下
	/*	go func() {
		for {
			Routes = FindRoutes()
			time.Sleep(time.Duration(30) * time.Second)
		}
	}()*/
	Routes = FindRoutes()
	// 监听变化
	go func() {

		lu := make(chan netlinkImp.LinkUpdate)
		done := make(chan struct{})
		defer close(done)
		err := netlinkImp.LinkSubscribe(lu, done)
		if err != nil {
			log.Println("监听 netlink 失败", err)
		}
		// 监听变化
		for {
			lui := <-lu
			ReaRouteUpdateChan <- true
			log.Println("监听 netlink 编号", lui)
			Routes = FindRoutes()
		}
	}()
}
