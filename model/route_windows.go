package model

import (
	"github.com/hellokeke123/anApp/winipcfg"
	"golang.org/x/sys/windows"
	"log"
	"net"
	"syscall"
	"unsafe"
)

type InterfaceIPAddress struct {
	ifIndex uint32
	ip      net.IP
}

// 找寻路由
func FindRoutes() []*CustomRoute {
	/*	defer func() {
		if r := recover(); r != nil {
			log.Println("update routes err", r)
		}
	}()*/

	table2, err := winipcfg.GetIPForwardTable2(windows.AF_INET)
	if err != nil {
		log.Println("GetIPForwardTable2() returned an error: %v", err)
		return nil
	}

	//unspec := net.IPv4(0, 0, 0, 0)

	mibIfRow2Map := make(map[winipcfg.LUID]winipcfg.MibIfRow2, len(table2))
	mibIPIfRowMap := make(map[winipcfg.LUID]*winipcfg.MibIPInterfaceRow, len(table2))

	routes := make([]*CustomRoute, 0)

	addresses := adapterAddresses()

	for i := range table2 {

		// 过滤回环和 网关为空的路由
		/*		if table2[i].Loopback || table2[i].NextHop.Addr().IsLoopback() || (table2[i].NextHop.Addr().String() == unspec.String()) {
				// Not a default route, so skip
				continue
			}*/

		var row winipcfg.MibIfRow2

		row, rowOk := mibIfRow2Map[table2[i].InterfaceLUID]
		row2, row2Ok := mibIPIfRowMap[table2[i].InterfaceLUID]

		if rowOk && row2Ok {

		} else {
			row.InterfaceIndex = table2[i].InterfaceIndex
			// https://learn.microsoft.com/zh-cn/windows/win32/api/netioapi/ns-netioapi-mib_if_row2
			// 获得网卡信息
			err = winipcfg.GtIfEntry2(&row)
			mibIfRow2Map[table2[i].InterfaceLUID] = row
			// 获得网卡ip信息
			row2, _ = row.InterfaceLUID.IPInterface(windows.AF_INET)
			mibIPIfRowMap[table2[i].InterfaceLUID] = row2

		}

		if row.MediaConnectState == winipcfg.MediaConnectStateConnected && row.Alias() != TunName {
			// 不想写了，直接调方法
			ip := addresses[row.InterfaceIndex]
			_, ipNet, err := net.ParseCIDR(table2[i].DestinationPrefix.Prefix().String())
			if err != nil {
				log.Println(row.Alias(), table2[i].DestinationPrefix.Prefix().String(), "cidr解析错误")
			} else {
				cr := &CustomRoute{
					ifName:  row.Alias(),
					Subnet:  ipNet,
					Ip:      ip,
					Metric:  (int)(table2[i].Metric + row2.Metric),
					IfIndex: table2[i].InterfaceIndex,
				}
				mask, _ := cr.Subnet.Mask.Size()
				if mask > 0 {
					cr.BitCount = countBits(ip, mask)
				}
				routes = append(routes, cr)
			}
		}
	}
	return routes
}

// InterfaceIPAddress
func adapterAddresses() map[uint32]net.IP {
	var b []byte
	l := uint32(15000) // recommended initial size
	for {
		b = make([]byte, l)
		err := windows.GetAdaptersAddresses(syscall.AF_INET, windows.GAA_FLAG_INCLUDE_PREFIX, 0, (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])), &l)
		if err == nil {
			if l == 0 {
				return nil
			}
			break
		}
		if err.(syscall.Errno) != syscall.ERROR_BUFFER_OVERFLOW {
			log.Println("getadaptersaddresses", err)
			return nil
		}
		if l <= uint32(len(b)) {
			log.Println("getadaptersaddresses", err)
			return nil
		}
	}
	result := make(map[uint32]net.IP, l)
	for aa := (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])); aa != nil && aa.FirstUnicastAddress != nil; aa = aa.Next {
		sockaddr, err := aa.FirstUnicastAddress.Address.Sockaddr.Sockaddr()
		if err == nil {
			inet4 := sockaddr.(*syscall.SockaddrInet4)
			ipv4 := net.IPv4(inet4.Addr[0], inet4.Addr[1], inet4.Addr[2], inet4.Addr[3])
			result[aa.IfIndex] = ipv4
			/*		switch sockaddr := sockaddr.(type) {
			  case *syscall.SockaddrInet4:
			}*/

		} else {
			log.Println(aa.IfIndex, "获取地址错误", err)
		}
	}
	return result
}

// 用于网卡改变时候的路由更新
func InitRoute() {
	winipcfg.RegisterInterfaceChangeCallback(
		func(notificationType winipcfg.MibNotificationType, iface *winipcfg.MibIPInterfaceRow) {
			log.Println("RegisterInterfaceChangeCallback", "回调")
			Routes = FindRoutes()
		})
	winipcfg.RegisterRouteChangeCallback(
		func(notificationType winipcfg.MibNotificationType, route *winipcfg.MibIPforwardRow2) {
			log.Println("RegisterRouteChangeCallback", "回调")
			Routes = FindRoutes()
		})
	// 定时任务，在回调失效的情况下
	/*	go func() {
		for {
			Routes = FindRoutes()
			time.Sleep(time.Duration(30) * time.Second)
		}
	}()*/
}

func countBits(ip net.IP, mask int) int {
	if ip == nil {
		return 0
	}
	count := 0
	recJ := 1
	if ip4 := ip.To4(); ip4 != nil {
		for _, b := range ip4 {
			for i := 0; i < 8 && recJ <= mask; i++ {
				recJ++
				if getBit(b, i) == 1 {
					count++
				}
			}
		}
		return count
	}
	return count
}

func getBit(b byte, n int) int {
	if n < 0 || n > 7 {
		return -1 // 错误：无效的位索引
	}
	return int((b >> uint(n)) & 1)
}
