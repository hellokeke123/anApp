package test

import (
	"fmt"
	"github.com/google/gopacket/pcap"
	"log"
	"testing"
)

func TestPcap(t *testing.T) {
	getHandle()
}

func getHandle() {
	ifs, _ := pcap.FindAllDevs()
	fmt.Println(ifs)
	// 打开网卡设备，例如eth0
	handle, err := pcap.OpenLive("\\Device\\NPF_{3905A368-E045-4A55-8452-44F0BF730E45}", 65536, true, pcap.BlockForever)
	if err != nil {
		log.Println(err)
	}
	defer handle.Close()
}
