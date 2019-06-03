package tunnel_client

import (
	"errors"
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/logs"
	"io/ioutil"
)

var SharedManager = NewManager()

type Manager struct {
}

func NewManager() *Manager {
	return &Manager{}
}

func (this *Manager) Start() error {
	logs.Println("start manager")
	configFile := Tea.Root + "/configs/config.yml"
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return errors.New("read 'config.yml': " + err.Error())
	}

	configs := &Config{}
	err = yaml.Unmarshal(data, &configs)
	if err != nil {
		return errors.New("read 'config.yml': " + err.Error())
	}

	for _, tunnelConfig := range configs.Tunnels {
		logs.Println("start '" + tunnelConfig.Remote + " -> " + tunnelConfig.Local + "'")

		err = tunnelConfig.Validate()
		if err != nil {
			return errors.New("validate '" + tunnelConfig.Remote + " -> " + tunnelConfig.Local + "' failed: " + err.Error())
		}

		tunnel := NewTunnel(tunnelConfig)
		go func(tunnel *Tunnel) {
			err = tunnel.Start()
			if err != nil {
				logs.Println("start '" + tunnelConfig.Remote + " -> " + tunnelConfig.Local + "' failed: " + err.Error())
			}
		}(tunnel)
	}

	return nil
}
