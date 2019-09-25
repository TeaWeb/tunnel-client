package main

import (
	"github.com/TeaWeb/tunnel-client/utils"
	"github.com/iwind/TeaGo/Tea"
	"log"
	"runtime"
)

// install service
func main() {
	log.Println("installing ...")
	manager := utils.NewServiceManager("TeaWeb Tunnel", "TeaWeb Tunnel Manager")

	var exePath = Tea.Root + Tea.DS + "bin" + Tea.DS + "teaweb-tunnel"
	if runtime.GOOS == "windows" {
		exePath += ".exe"
	}
	err := manager.Install(exePath, []string{"service"})
	if err != nil {
		log.Println("ERROR: " + err.Error())
		manager.PauseWindow()
		return
	}

	log.Println("install service successfully")

	// start
	log.Println("starting ...")
	err = manager.Start()
	if err != nil {
		log.Println("ERROR: " + err.Error())
	}

	log.Println("started successfully")
	log.Println("done.")

	manager.PauseWindow()
}
