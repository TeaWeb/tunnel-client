package tunnel_client

import "testing"

func TestTunnel_Start(t *testing.T) {
	config := &TunnelConfig{
		Remote: "127.0.0.1:9001",
		Local:  "http://127.0.0.1",
		Host:   "www.teaos.cn",
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
