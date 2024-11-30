package test

import (
	mdns "github.com/miekg/dns"
	"log"
	"net"
	"testing"
)

func TestDnsHeader(t *testing.T) {
	query := mdns.Msg{}
	query.SetQuestion("www,baidu.com.", mdns.TypeA)
	msg, _ := query.Pack()

	// 发送上诉报文
	udpConn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IP{8, 8, 8, 8},
		Port: 53,
	})
	if err != nil {
		log.Panic("net dial udp error ", err)
	}

	// 发送报文
	udpConn.Write(msg)

	// 等待DNS服务器回应
	buf := make([]byte, 1024)

	n, err := udpConn.Read(buf)
	if err != nil {
		log.Println("udp conn read error ", err)
	}

	// 解析响应报文

	response := mdns.Msg{}
	response.Unpack(buf[:n])
	log.Println("获取数据:", response.String())
	log.Println("结束")
}
