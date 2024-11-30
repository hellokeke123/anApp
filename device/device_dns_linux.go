package device

import (
	"github.com/hellokeke123/anApp/model"
	"log"
	"os"
	"os/exec"
)

func (d *Device) setDns(dns string) {
	ctlPath, err := exec.LookPath("resolvectl")
	if err != nil {
		log.Println("monitorEnforceDns setdns err", err)
		return
	}
	_ = Exec(ctlPath, "domain", d.Name, "~.").Run()
	_ = Exec(ctlPath, "default-route", d.Name, "true").Run()
	//_ = Exec(ctlPath, append([]string{"dns", d.Name}, common.Map(dns, string)...)...).Run()
	_ = Exec(ctlPath, append([]string{"dns", d.Name}, dns)...).Run()
}

func (d *Device) monitorEnforceDns() {
	model.ContextConfigImp.ContextClient.ReadEnableEnforceDnsChan = make(chan bool, 1)
	model.ContextConfigImp.ContextClient.WriteEnableEnforceDnsChan = make(chan bool, 1)
	log.Println("monitorEnforceDns start")
	for {
		enableEnforceDns := <-model.ContextConfigImp.ContextClient.ReadEnableEnforceDnsChan
		if enableEnforceDns != model.ContextConfigImp.ContextClient.EnableEnforceDns {
			model.ContextConfigImp.ContextClient.EnableEnforceDns = enableEnforceDns
			if enableEnforceDns {
				d.setDns(model.DNS1)
			} else {
				log.Println("monitorEnforceDns cancel start")
				ctlPath, err := exec.LookPath("resolvectl")
				if err != nil {
					log.Println("monitorEnforceDns cancel err", err)
				}
				_ = Exec(ctlPath, "revert", d.Name).Run()
				//_ = Exec(ctlPath, "default-route", d.Name, "false").Run()
			}
		}

		model.ContextConfigImp.ContextClient.WriteEnableEnforceDnsChan <- true
	}
}

func Exec(name string, args ...string) *exec.Cmd {
	command := exec.Command(name, args...)
	command.Env = os.Environ()
	return command
}
