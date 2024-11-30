package socket

import (
	"github.com/hellokeke123/anApp/model"
	"log"
	"net"
	"syscall"
	"time"
)

func GetDialer(route *model.CustomRoute) net.Dialer {
	return net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			c.Control(func(fd uintptr) {
				// unix 也有方法绑定
				//或者unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_BINDTOIFINDEX, interfaceIndex)
				//或者unix.BindToDevice(int(fd), interfaceName)
				if err := syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, route.GetIfName()); err != nil {
					log.Println("bind ifindex fail ", err)
				}
			})
			return nil
		},
		Timeout: 5 * time.Second,
	}
}
