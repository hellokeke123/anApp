package netlink

import (
	netlinkImp "github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	"net"
	"testing"
)

func TestNetlink(t *testing.T) {
	list, err := netlinkImp.RouteList(nil, netlinkImp.FAMILY_V4)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Log(list)
	}

	ifc, err := net.InterfaceByIndex(3)
	//deviceList, err := netlinkImp.G

	//sgs, err := conn.Receive()
	t.Log(ifc)

	addrs, _ := ifc.Addrs()

	for _, addr := range addrs {
		s := addr.String()
		t.Log(s)
	}

	// 监听变化
	lu := make(chan netlinkImp.LinkUpdate)
	done := make(chan struct{})
	defer close(done)
	err = netlinkImp.LinkSubscribe(lu, done)
	if err != nil {
		t.Fatal(err)
	}
	// 监听变化
	for {
		lui := <-lu
		t.Log(lui)
	}

	//_, ipNet, _ := net.ParseCIDR("0.0.0.0/0")
	rule := netlinkImp.NewRule()
	rule.Priority = 200
	rule.Table = 254
	rule.SuppressPrefixlen = 0

	err = netlinkImp.RuleAdd(rule)

	if err != nil {
		t.Fatal(err)
	}

	rule1 := netlinkImp.NewRule()
	rule1.Priority = 201
	rule1.Table = 254
	rule1.IPProto = unix.IPPROTO_ICMP

	err = netlinkImp.RuleAdd(rule1)

	if err != nil {
		t.Fatal(err)
	}

	rule2 := netlinkImp.NewRule()
	rule2.Priority = 202
	rule2.Table = 200

	err = netlinkImp.RuleAdd(rule2)

	if err != nil {
		t.Fatal(err)
	}

	link, err := netlinkImp.LinkByName("nekoray-tun")
	if err != nil {
		t.Fatal(err)
	}
	_, ipNet, _ := net.ParseCIDR("0.0.0.0/0")

	err = netlinkImp.RouteAdd(
		&netlinkImp.Route{
			LinkIndex: link.Attrs().Index,
			Dst:       ipNet,
			Table:     200,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}
