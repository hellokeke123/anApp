package test

import (
	"fmt"
	"syscall"
	"testing"
)

func TestGetRoute(t *testing.T) {
	tab, err := syscall.NetlinkRIB(syscall.RTM_GETROUTE, syscall.AF_INET)
	if err != nil {
		panic(err)
	}
	msgs, err := syscall.ParseNetlinkMessage(tab)
	if err != nil {
		panic(err)
	}
	for _, m := range msgs {
		switch m.Header.Type {
		case syscall.NLMSG_DONE:
			fmt.Println("recv done")
			goto done
		case syscall.RTM_NEWROUTE:
			// 解析数据
		}
	}
done:
	t.Log(msgs)
}
