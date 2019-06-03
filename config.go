package tunnel_client

type Config struct {
	Tunnels []*TunnelConfig `yaml:"tunnels"`
}
