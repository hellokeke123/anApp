package procTool

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

/*
做退出程序监控
*/
func InitDaemonMonitor() {
	//系统退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL) // , syscall.SIGHUP  ssh断开也会
	// 自定义退出
	eSign := make(chan int, 1)
	// 创建父进程监控 回调
	monitorParentProcess(eSign)
	// 收到信号回调
	for {
		select {
		case sign := <-sigCh:
			log.Println("进程信号，退出", sign)
			return
		case <-eSign:
			log.Println("父进程状态变化，退出")
			return
		}
	}
}

// 根据当前线程创建子进程
func CreatChildProcess(arg ...string) (*exec.Cmd, error) {
	filePath, _ := filepath.Abs(os.Args[0])
	procd := exec.Command(filePath, arg...)

	// 将其他命令传入生成出的进程 共享进程流
	/*		cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr*/

	handleStdout := func(r io.Reader) {
		log.Println("获取到daemon输出流")
		s := bufio.NewScanner(r)
		for s.Scan() {
			line := s.Text()
			log.Println(line)
		}
		if err := s.Err(); err != nil {
			log.Println("daemon输出流异常", err)
		}
	}

	stdout, err := procd.StdoutPipe()
	if err != nil {
		log.Println("获取daemon输出流失败", err)
	} else {
		go handleStdout(stdout)
	}

	err = procd.Start()
	return procd, err
}

/*
做退出程序监控
*/
func ManageDaemonMonitor(prod *os.Process) {
	//系统退出
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGHUP)
	// 自定义退出
	eSign := make(chan int, 1)
	// 创建父进程监控 回调
	monitorDaemonProcess(prod, eSign)
	// 收到信号回调
	for {
		select {
		case sign := <-sigCh:
			log.Println("进程信号，退出", sign)
			return
		case <-eSign:
			log.Println("子进程状态变化，退出")
			return
		}
	}
}
