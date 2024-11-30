package main

import (
	"github.com/BurntSushi/toml"
	"github.com/hellokeke123/anApp/app/v2/client"
	"github.com/hellokeke123/anApp/model"
	logTool "github.com/hellokeke123/anApp/tool/log"
	"github.com/hellokeke123/anApp/tool/procTool"
	"log"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == model.PROC_DAEMON {
		// 创建日志文件
		logTool.CreatLog()

		contextConfig := model.ContextConfig{}
		if _, err := toml.DecodeFile("./cfg/cfg.toml", &contextConfig); err != nil {
			panic(err)
		}
		model.SetContextConfig(contextConfig)
		if contextConfig.App.Context == model.CLIENT {
			client.CreateClient()
		} else {
			log.Println("错误的启动")
		}
		log.Println("anApp is closing...")

	} else {
		// 开启子进程
		procd, err := procTool.CreatChildProcess(model.PROC_DAEMON)

		if err != nil {
			log.Println("CreatChildProcess error", err)
			return
		}
		// 监控退出
		procTool.ManageDaemonMonitor(procd.Process)
		// 线程的退出
		procd.Process.Signal(os.Interrupt)
	}
}
