package socket

import (
	"encoding/binary"
	"fmt"
	mdns "github.com/miekg/dns"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

/*
*
有效， 高配版
*/

func tcpSocket(t *testing.T) {
	dialer := net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			c.Control(func(fd uintptr) {
				handle := syscall.Handle(fd)
				if err := bind4(handle, 16); err != nil {
					panic(err)
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
正确
*/
func udpSocket(t *testing.T) {
	dialer := net.Dialer{
		Control: func(network, address string, c syscall.RawConn) error {
			c.Control(func(fd uintptr) {
				handle := syscall.Handle(fd)
				if err := bind4(handle, 16); err != nil {
					panic(err)
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
