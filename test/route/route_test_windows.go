package test

import (
	"fmt"
	"github.com/boss-net/goutils/routing"
	"github.com/hellokeke123/anApp/winipcfg"
	"golang.org/x/sys/windows"
	"net"
	"testing"
)

func TestWindowsRoute(t *testing.T) {
	router, _ := routing.New()
	iface, gateway, src, err := router.Route([]byte{175, 178, 37, 89})
	fmt.Print("合适的路由", iface, gateway, src, err)
}

// 尝试获取有效路由
// https://learn.microsoft.com/zh-cn/windows/win32/api/netioapi/nf-netioapi-getipforwardtable2
func getIPForwardTable2(t *testing.T) {
	table2, err := winipcfg.GetIPForwardTable2(windows.AF_INET)
	if err != nil {
		t.Errorf("GetIPForwardTable2() returned an error: %v", err)
		return
	}

	//validRoutes := make([]winipcfg.MibIPforwardRow2, 100)
	unspec := net.IPv4(0, 0, 0, 0)

	for i := range table2 {

		// 过滤回环和 网关为空的路由
		// 目的是找寻可靠的接口
		if table2[i].Loopback || table2[i].NextHop.Addr().IsLoopback() || (table2[i].NextHop.Addr().String() == unspec.String()) {
			// Not a default route, so skip
			continue
		}
		var row winipcfg.MibIfRow2
		row.InterfaceIndex = table2[i].InterfaceIndex
		// https://learn.microsoft.com/zh-cn/windows/win32/api/netioapi/ns-netioapi-mib_if_row2
		// 获得网卡信息
		err = winipcfg.GtIfEntry2(&row)
		// 获得网卡ip信息
		row2, _ := row.InterfaceLUID.IPInterface(windows.AF_INET)
		// 已连接网卡，似乎判断多余了
		if row.MediaConnectState == winipcfg.MediaConnectStateConnected {
			fmt.Println(row.Description())
			fmt.Println(row.Type)
			fmt.Println(table2[i].DestinationPrefix.Prefix())
			fmt.Println(table2[i].Metric + row2.Metric)
			fmt.Println(row2.DisableDefaultRoutes)
			fmt.Println(table2[i].NextHop.Addr().String())

		}

	}
}
