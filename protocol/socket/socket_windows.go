package socket

import (
	"encoding/binary"
	"github.com/hellokeke123/anApp/model"
	"net"
	"syscall"
	"time"
	"unsafe"
)

func GetDialer(route *model.CustomRoute) net.Dialer {
	return net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			c.Control(func(fd uintptr) {
				handle := syscall.Handle(fd)
				if err := bind4(handle, int(route.IfIndex)); err != nil {
					panic(err)
				}
			})
			return nil
		},
		Timeout: 5 * time.Second,
	}
}

const (
	IP_UNICAST_IF   = 31
	IPV6_UNICAST_IF = 31
)

func bind4(handle syscall.Handle, ifaceIdx int) error {
	var bytes [4]byte
	binary.BigEndian.PutUint32(bytes[:], uint32(ifaceIdx))
	idx := *(*uint32)(unsafe.Pointer(&bytes[0]))
	//syscall.Setsockopt 其它后缀也行，不一定这种方式
	return syscall.SetsockoptInt(handle, syscall.IPPROTO_IP, IP_UNICAST_IF, int(idx))
}

func bind6(handle syscall.Handle, ifaceIdx int) error {
	//syscall.Setsockopt 其它后缀也行，不一定这种方式
	return syscall.SetsockoptInt(handle, syscall.IPPROTO_IPV6, IPV6_UNICAST_IF, ifaceIdx)
}
