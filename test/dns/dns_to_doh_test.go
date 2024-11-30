package test

import (
	"bytes"
	"github.com/hellokeke123/anApp/protocol/dns"
	mdns "github.com/miekg/dns"
	"log"
	"testing"
)

func TestDnsToDoh(t *testing.T) {
	query := mdns.Msg{}
	query.SetQuestion("www,baidu.com.", mdns.TypeA)
	msg, _ := query.Pack()
	buff, err := dns.SendDoh("https://1.1.1.1/dns-query", bytes.NewBuffer(msg))

	if buff == nil || err != nil {
		log.Println("dns解析失败:", err)
	}

	// 解析响应报文

	response := mdns.Msg{}
	response.Unpack(buff.Bytes())
	log.Println("获取数据:", response.String())
	log.Println("结束")
}
