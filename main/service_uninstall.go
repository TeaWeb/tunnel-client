package main

import (
	"github.com/TeaWeb/tunnel-client/utils"
	"log"
)

// 卸载服务
func main() {
	log.Println("uninstalling ...")
	manager := utils.NewServiceManager("TeaWeb Tunnel", "TeaWeb Tunnel Manager")
	err := manager.Uninstall()
	if err != nil {
		log.Println("ERROR: " + err.Error())
		manager.Close()
		manager.PauseWindow()
		return
	}

	log.Println("uninstalled service successfully")
	log.Println("done.")

	manager.PauseWindow()
}
