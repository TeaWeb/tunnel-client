package tunnel_client

import "net/url"

type TunnelConfig struct {
	Remote string `yaml:"remote"`
	Local  string `yaml:"local"`
	Host   string `yaml:"host"`

	localHost   string
	localScheme string
}

func (this *TunnelConfig) Validate() error {
	u, err := url.Parse(this.Local)
	if err != nil {
		return err
	}
	this.localHost = u.Host
	this.localScheme = u.Scheme

	return nil
}

func (this *TunnelConfig) LocalHost() string {
	return this.localHost
}

func (this *TunnelConfig) LocalScheme() string {
	return this.localScheme
}
