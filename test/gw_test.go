package test

import (
	"fmt"
	"net"
	"os/exec"
	"testing"
)

func TestGw(t *testing.T) {
	getGw()
}

func getGw() {
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 遍历每个网络接口
	for _, iface := range ifaces {
		// 获取接口的 IP 地址列表
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// 遍历每个 IP 地址
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			// 如果目标地址和接口在同一子网，则该接口的网关为系统网关
			if ipNet.Contains(net.IPv4(8, 8, 8, 8)) {
				gw := getGateway(iface.Name)
				fmt.Printf("Use interface %s with gateway %s\n", iface.Name, gw)
				return
			}
		}
	}
}

// 根据接口名称获取网关地址
func getGateway(ifaceName string) string {
	routeCmd := fmt.Sprintf("route -n get default | awk '/interface: %s/{getline; print $2}'", ifaceName)
	output, err := exec.Command("sh", "-c", routeCmd).Output()
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	return string(output)
}
