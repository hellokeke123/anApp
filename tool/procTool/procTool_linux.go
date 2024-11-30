package procTool

import (
	"log"
	"os"
	"syscall"
	"time"
)

// 监控父进程 回调
// 等待父进程退出
func monitorParentProcess(eSign chan int) {
	go func() {
		// parent pid
		ppid := syscall.Getppid()
		for {
			log.Println("ppid", ppid)
			p, err := os.FindProcess(ppid)
			if err != nil || p == nil {
				log.Println("parent process monitor not exit", err)
				eSign <- 100
				break
			} else {
				wait, err := p.Wait()
				if wait != nil && wait.Exited() {
					eSign <- 100
					break
				} else if err != nil {
					log.Println("parent process monitor Wait", err)
				}
			}
			time.Sleep(10 * time.Second)
		}
	}()
}

// 监控子进程 回调
// 等待父进程退出
func monitorDaemonProcess(p *os.Process, eSign chan int) {
	go func() {
		if p == nil {
			log.Println("daemon monitor not exit")
			eSign <- 100
		} else {
			wait, err := p.Wait()
			if wait != nil && wait.Exited() {
				eSign <- 100
			} else if err != nil {
				log.Println("daemon monitor Wait", err)
			}
		}
	}()
}
