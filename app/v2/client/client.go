package client

import (
	"fmt"
	"github.com/hellokeke123/anApp/device"
	"github.com/hellokeke123/anApp/gv"
	"github.com/hellokeke123/anApp/model"
	"github.com/hellokeke123/anApp/tool/procTool"
	"log"
	"runtime"
)

func CreateClient() {
	// 更新路由任务
	model.InitRoute()

	dev, err := device.NewDevice(model.TunName)

	runtime.GOMAXPROCS(runtime.NumCPU() * 10)

	if err != nil {
		log.Println("tun create fail!!!", err)
		return
	} else {
		log.Println("tun success")
	}

	dev.InitGateway()

	err = dev.Config()
	if err != nil {
		log.Println("tun config fail!!!", err)
	}

	defer dev.Close()
	// 做基础配置

	err = gv.Start(dev, model.MTU)
	if err != nil {
		fmt.Println("错误:", err)
	}

	// 退出信号
	procTool.InitDaemonMonitor()

}
