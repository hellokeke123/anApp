package device

import (
	"errors"
	"github.com/hellokeke123/anApp/model"
	netlinkImp "github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/tun"
	"io"
	"log"
	"net"
	"os"
	"unsafe"
)

// Device is ...
type Device struct {
	// NativeTun is ...
	*tun.NativeTun
	//
	io.Writer
	// Name is ...
	Name string

	fwpmSession uintptr

	ruleList []*netlinkImp.Rule
}

// 做路由配置
func (device *Device) Config() (err error) {

	device.ruleList = make([]*netlinkImp.Rule, 0)

	//_, ipNet, _ := net.ParseCIDR("0.0.0.0/0")
	// 匹配表中非默认路由的规则
	rule := netlinkImp.NewRule()
	rule.Priority = 200
	rule.Table = 254
	rule.SuppressPrefixlen = 0

	err = netlinkImp.RuleAdd(rule)

	if err != nil && err != unix.EEXIST {
		return err
	} else {
		device.ruleList = append(device.ruleList, rule)
	}

	rule1 := netlinkImp.NewRule()
	rule1.Priority = 201
	rule1.Table = 254
	rule1.IPProto = unix.IPPROTO_ICMP

	err = netlinkImp.RuleAdd(rule1)

	if err != nil && err != unix.EEXIST {
		return err
	} else {
		device.ruleList = append(device.ruleList, rule1)
	}

	rule2 := netlinkImp.NewRule()
	rule2.Priority = 202
	rule2.Table = 200

	err = netlinkImp.RuleAdd(rule2)

	if err != nil && err != unix.EEXIST {
		return err
	} else {
		device.ruleList = append(device.ruleList, rule2)
	}

	link, err := netlinkImp.LinkByName(model.TunName)
	if err != nil {
		return err
	}
	_, ipNet, _ := net.ParseCIDR("0.0.0.0/0")

	err = netlinkImp.RouteAdd(
		&netlinkImp.Route{
			LinkIndex: link.Attrs().Index,
			Dst:       ipNet,
			Table:     200,
		},
	)
	// 保证默认网卡的入口也为出口流量， 部分高版本linux 会自动处理 比如 ubuntu22 ， 大部分多网卡非默认网卡的入口流量网卡会和出口网卡不一致，因为出口还会走路由
	// 这里只保证默认网卡流量正常， 因为如果多网卡本来有问题，为神马要这里解决系统的问题，只处理tun 网卡造成的影响
	route := model.FindContainRoute(net.IPv4(0, 0, 0, 0), model.Routes)
	rule99 := netlinkImp.NewRule()
	rule99.Priority = 150
	rule99.Table = 254
	_, defaultSrcNet, _ := net.ParseCIDR(route.Ip.String() + "/32")
	rule99.Src = defaultSrcNet

	err = netlinkImp.RuleAdd(rule99)
	if err != nil && err != unix.EEXIST {
		return err
	} else {
		device.ruleList = append(device.ruleList, rule99)

		// 这里做动态更新
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("动态更新默认网卡 『ip rule』 入出口流量一致规则")
				}
			}()

			for {
				<-model.ReaRouteUpdateChan
				err = netlinkImp.RuleDel(rule99)
				_, defaultSrcNet, _ := net.ParseCIDR(route.Ip.String() + "/32")
				rule99.Src = defaultSrcNet
				err = netlinkImp.RuleAdd(rule99)
				if err != nil && err != unix.EEXIST {
					log.Println("动态更新默认网卡 『ip rule』 入出口流量一致规则失败", err)
				} else {
					log.Println("动态更新默认网卡 『ip rule』 入出口流量一致规则成功")
				}

			}

		}()

	}

	if err != nil && err != unix.EEXIST {
		return err
	}
	// 判断是否监控dns
	if model.ContextConfigImp.ContextClient.EnableEnforceDns {
		device.setDns(model.DNS1)
		go device.monitorEnforceDns()
	}
	return nil
}

// CreateTUN is ...
func CreateTUN(name string, mtu int) (dev *Device, err error) {
	dev = &Device{}
	//device, err := tun.CreateTUN(name, mtu)
	device, err := CreateTUN0(name, mtu)
	if err != nil {
		return
	}
	dev.NativeTun = device.(*tun.NativeTun)
	dev.Writer = dev
	if dev.Name, err = dev.NativeTun.Name(); err != nil {
		return
	}
	return
}

func (device *Device) InitGateway() {

	ip, ipNet, _ := net.ParseCIDR(model.TunIp)
	device.setInterfaceAddress4([4]byte(ip.To4()), [4]byte(ipNet.Mask), [4]byte{})
	device.Activate()
	/*	err := device.addRouteEntry4([]string{"157.148.69.80/32"})
		if err != nil {
			panic(err)
		}*/
}

// SetInterfaceAddress is ...
// 192.168.1.11/24
// fe80:08ef:ae86:68ef::11/64
func (d *Device) SetInterfaceAddress(address string) error {
	//if _, _, gateway, err := getInterfaceConfig4(address); err == nil {
	//	return d.setInterfaceAddress4("", address, gateway)
	//}
	//if _, _, gateway, err := getInterfaceConfig6(address); err == nil {
	//	return d.setInterfaceAddress6("", address, gateway)
	//}
	return errors.New("tun device address error")
}

// setInterfaceAddress4 is ...
// https://github.com/daaku/go.ip/blob/master/ip.go
func (d *Device) setInterfaceAddress4(addr, mask, gateway [4]byte) (err error) {
	//Addr := parse4(addr)
	//Mask := parse4(mask)
	//Gateway := parse4(gateway)

	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_IP)
	if err != nil {
		return err
	}
	defer unix.Close(fd)

	// ifreq_addr is ...
	type ifreq_addr struct {
		ifr_name [unix.IFNAMSIZ]byte
		ifr_addr unix.RawSockaddrInet4
		_        [8]byte
	}

	ifra := ifreq_addr{
		ifr_addr: unix.RawSockaddrInet4{
			Family: unix.AF_INET,
		},
	}
	copy(ifra.ifr_name[:], d.Name[:])

	ifra.ifr_addr.Addr = addr
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.SIOCSIFADDR, uintptr(unsafe.Pointer(&ifra))); errno != 0 {
		return os.NewSyscallError("ioctl: SIOCSIFADDR", errno)
	}

	ifra.ifr_addr.Addr = mask
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.SIOCSIFNETMASK, uintptr(unsafe.Pointer(&ifra))); errno != 0 {
		return os.NewSyscallError("ioctl: SIOCSIFNETMASK", errno)
	}

	return nil
}

// in6_addr
type in6_addr struct {
	addr [16]byte
}

// setInterfaceAddres6 is ...
func (d *Device) setInterfaceAddress6(addr, mask, gateway string) error {
	Addr := parse6(addr)
	Mask := parse6(mask)
	//Gateway := parse6(gateway)

	fd, err := unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, unix.IPPROTO_IP)
	if err != nil {
		return err
	}
	defer unix.Close(fd)

	// ifreq_ifindex is ...
	type ifreq_ifindex struct {
		ifr_name    [unix.IFNAMSIZ]byte
		ifr_ifindex int32
		_           [20]byte
	}

	ifrf := ifreq_ifindex{}
	copy(ifrf.ifr_name[:], d.Name[:])

	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.SIOCGIFINDEX, uintptr(unsafe.Pointer(&ifrf))); errno != 0 {
		return os.NewSyscallError("ioctl: SIOCGIFINDEX", errno)
	}

	// in6_ifreq_addr is ...
	type in6_ifreq_addr struct {
		ifr6_addr      in6_addr
		ifr6_prefixlen uint32
		ifr6_ifindex   int32
	}

	ones, _ := net.IPMask(Mask[:]).Size()

	ifra := in6_ifreq_addr{
		ifr6_addr: in6_addr{
			addr: Addr,
		},
		ifr6_prefixlen: uint32(ones),
		ifr6_ifindex:   ifrf.ifr_ifindex,
	}

	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.SIOCSIFADDR, uintptr(unsafe.Pointer(&ifra))); errno != 0 {
		return os.NewSyscallError("ioctl: SIOCSIFADDR", errno)
	}

	return nil
}

// addRouteEntry4 is ...
func (d *Device) addRouteEntry4(cidr []string) error {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_IP)
	if err != nil {
		return err
	}
	defer unix.Close(fd)

	nameBytes := [16]byte{}
	copy(nameBytes[:], d.Name[:])

	route := rtentry{
		rt_dst: unix.RawSockaddrInet4{
			Family: unix.AF_INET,
		},
		//rt_gateway: unix.RawSockaddrInet4{
		//	Family: unix.AF_INET,
		//	Addr:   [4]byte{0, 0, 0, 0},
		//},
		rt_genmask: unix.RawSockaddrInet4{
			Family: unix.AF_INET,
		},
		rt_flags: unix.RTF_UP, // | unix.RTF_GATEWAY
		rt_dev:   uintptr(unsafe.Pointer(&nameBytes)),
	}

	for _, item := range cidr {
		_, ipNet, _ := net.ParseCIDR(item)

		ipv4 := ipNet.IP.To4()
		mask := net.IP(ipNet.Mask).To4()

		route.rt_dst.Addr = *(*[4]byte)(unsafe.Pointer(&ipv4[0]))
		route.rt_genmask.Addr = *(*[4]byte)(unsafe.Pointer(&mask[0]))

		if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.SIOCADDRT, uintptr(unsafe.Pointer(&route))); errno != 0 {
			return os.NewSyscallError("ioctl: SIOCADDRT", errno)
		}
	}

	return nil
}

// addRouteEntry6 is ...
func (d *Device) addRouteEntry6(cidr []string) error {
	fd, err := unix.Socket(unix.AF_INET6, unix.SOCK_DGRAM, unix.IPPROTO_IP)
	if err != nil {
		return err
	}
	defer unix.Close(fd)

	// ifreq_ifindex is ...
	type ifreq_ifindex struct {
		ifr_name    [unix.IFNAMSIZ]byte
		ifr_ifindex int32
		_           [20]byte
	}

	ifrf := ifreq_ifindex{}
	copy(ifrf.ifr_name[:], d.Name[:])

	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.SIOCGIFINDEX, uintptr(unsafe.Pointer(&ifrf))); errno != 0 {
		return os.NewSyscallError("ioctl: SIOCGIFINDEX", errno)
	}

	route := in6_rtmsg{
		rtmsg_metric:  1,
		rtmsg_ifindex: ifrf.ifr_ifindex,
	}

	for _, item := range cidr {
		_, ipNet, _ := net.ParseCIDR(item)

		ipv6 := ipNet.IP.To16()
		mask := net.IP(ipNet.Mask).To16()

		ones, _ := net.IPMask(mask).Size()
		route.rtmsg_dst.addr = *(*[16]byte)(unsafe.Pointer(&ipv6[0]))
		route.rtmsg_dst_len = uint16(ones)
		route.rtmsg_flags = unix.RTF_UP
		if ones == 128 {
			route.rtmsg_flags |= unix.RTF_HOST
		}

		if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.SIOCADDRT, uintptr(unsafe.Pointer(&route))); errno != 0 {
			return os.NewSyscallError("ioctl: SIOCADDRT", errno)
		}
	}

	return nil
}

// Activate is ...
func (d *Device) Activate() error {
	fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_IP)
	if err != nil {
		return err
	}
	defer unix.Close(fd)

	// ifreq_flags is ...
	type ifreq_flags struct {
		ifr_name  [unix.IFNAMSIZ]byte
		ifr_flags uint16
		_         [22]byte
	}

	ifrf := ifreq_flags{}
	copy(ifrf.ifr_name[:], d.Name[:])

	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.SIOCGIFFLAGS, uintptr(unsafe.Pointer(&ifrf))); errno != 0 {
		return os.NewSyscallError("ioctl: SIOCGIFFLAGS", errno)
	}

	ifrf.ifr_flags = ifrf.ifr_flags | unix.IFF_UP | unix.IFF_RUNNING
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), unix.SIOCSIFFLAGS, uintptr(unsafe.Pointer(&ifrf))); errno != 0 {
		return os.NewSyscallError("ioctl: SIOCSIFFLAGS", errno)
	}
	return nil
}

func (device *Device) Close() (err error) {

	if device.ruleList != nil && len(device.ruleList) > 0 {
		for _, rule := range device.ruleList {
			netlinkImp.RuleDel(rule)
		}
	}

	err = device.NativeTun.Close()
	return
}
