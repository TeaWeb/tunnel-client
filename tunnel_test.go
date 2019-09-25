package tunnel_client

import (
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaGo/logs"
	"testing"
)

func TestTunnel_Start(t *testing.T) {
	config := &TunnelConfig{
		Remote: "192.168.2.40:8884",
		Local:  "http://127.0.0.1:9991",
		Host:   "www.teaos.cn",
		Secret: "YKCXgsGlDcZv7o5VEjF2iT5K4t3ae5bE",
	}
	err := config.Validate()
	if err != nil {
		t.Fatal(err)
	}

	tunnel := NewTunnel(config)
	err = tunnel.Start()
	if err != nil {
		t.Fatal(err)
	}
}

func TestTunnel_StartRoot(t *testing.T) {
	config := &TunnelConfig{
		Remote: "192.168.2.40:8884",
		Host:   "www.teaos.cn",
		Root:   ".",
		Secret: "YKCXgsGlDcZv7o5VEjF2iT5K4t3ae5bE",
	}
	err := config.Validate()
	if err != nil {
		t.Fatal(err)
	}

	tunnel := NewTunnel(config)
	err = tunnel.Start()
	if err != nil {
		t.Fatal(err)
	}
}

func TestTunnel_Config_Decode(t *testing.T) {
	config := &Config{}
	err := yaml.Unmarshal([]byte(`tunnels:
  - remote: "8.8.8.8:9001"
    root: "D:\\web"
    index: [ "index.html" ]
    host: ""
    secret: ""`), config)
	if err != nil {
		t.Fatal(err)
	}
	logs.PrintAsJSON(config, t)
}
