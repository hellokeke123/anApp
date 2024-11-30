package socket

import (
	"fmt"
	mdns "github.com/miekg/dns"
	"golang.org/x/sys/unix"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"
)

/*
*
有效， 高配版
*/
func tcpSocket(t *testing.T) {
	dialer := net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			c.Control(func(fd uintptr) {
				// unix 也有方法绑定
				//或者unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_BINDTOIFINDEX, interfaceIndex)
				//或者unix.BindToDevice(int(fd), interfaceName)
				if err := syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, "wlp0s20f3"); err != nil {
				}
			})
			return nil
		},
		Timeout: 5 * time.Second,
	}

	// 创建一个自定义的 Transport，并指定使用上面创建的 Dialer
	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	client := &http.Client{
		Transport: transport,
	}

	url := "https://sy.tyykj.com/"
	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status code:", resp.Status)
	fmt.Println("Response body:")
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

/*
错误, 这种方式如果不传 ladrr , 那么网卡正确，不能正确绑定网卡ip
*/
func udpSocket1(t *testing.T) {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IP{114, 114, 114, 113},
		Port: 53,
	})
	rawConn, err := conn.SyscallConn()
	if err != nil {

	}
	rawConn.Control(func(fd uintptr) {
		if err := syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, "wlp0s20f3"); err != nil {
			t.Error("bind device fail", err)
		}
	})

	query := mdns.Msg{}
	query.SetQuestion("www.baidu.com.", mdns.TypeA)
	msg, _ := query.Pack()
	conn.Write(msg)

	// 等待DNS服务器回应
	buf := make([]byte, 1024)

	n, err := conn.Read(buf)
	if err != nil {
		log.Println("udp conn read error ", err)
	}

	// 解析响应报文

	response := mdns.Msg{}
	response.Unpack(buf[:n])
	t.Log("获取数据:", response.String())
	t.Log("结束")
}

/*
正确
*/
func udpSocket(t *testing.T) {
	dialer := net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			c.Control(func(fd uintptr) {
				// unix 也有方法绑定
				//或者unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_BINDTOIFINDEX, interfaceIndex)
				//或者unix.BindToDevice(int(fd), interfaceName)
				/*		if err := syscall.SetsockoptString(int(fd), syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, "wlp0s20f3"); err != nil {
						}*/
				if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
					log.Println("reuse socket fail", err)
				}
			})
			return nil
		},
		Timeout: 5 * time.Second,
	}

	conn, err := dialer.Dial("udp", "114.114.114.114:53")

	query := mdns.Msg{}
	query.SetQuestion("www.baidu.com.", mdns.TypeA)
	msg, _ := query.Pack()
	conn.Write(msg)

	// 等待DNS服务器回应
	buf := make([]byte, 1024)

	n, err := conn.Read(buf)
	if err != nil {
		log.Println("udp conn read error ", err)
	}

	// 解析响应报文

	response := mdns.Msg{}
	response.Unpack(buf[:n])
	t.Log("获取数据:", response.String())
	t.Log("结束")
}
