package test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
)

func TestHttp(t *testing.T) {
	sendHttpReq()
}

func sendHttpReq() {
	// 创建一个 Dialer，并设置本地 IP 和端口
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP:   net.ParseIP("192.168.3.120"), // 设置本地 IP 地址
			Port: 0,                            // 0 表示随机选择一个空闲端口
		},
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
