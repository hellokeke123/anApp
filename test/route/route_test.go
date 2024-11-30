package test

import (
	"fmt"
	"github.com/boss-net/goutils/routing"
	"github.com/hellokeke123/anApp/model"
	"net"
	"testing"
)

func TestRoute(t *testing.T) {
	router, _ := routing.New()
	iface, gateway, src, err := router.Route([]byte{175, 178, 37, 89})
	fmt.Print("合适的路由", iface, gateway, src, err)
}

func TestRoute1(t *testing.T) {
	routes := model.FindRoutes()
	fmt.Println(routes)
	fmt.Println(model.FindContainRoute(net.ParseIP("8.8.8.8"), routes))
}
