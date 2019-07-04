package tunnel_client

import "testing"

func TestTunnel_Start(t *testing.T) {
	config := &TunnelConfig{
		Remote: "192.168.2.40:8884",
		Local:  "http://127.0.0.1:9991",
		Host:   "www.teaos.cn",
		Secret: "",
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
