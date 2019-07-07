package tunnel_client

import (
	"errors"
	"net/url"
)

type TunnelConfig struct {
	Remote string   `yaml:"remote"`
	Local  string   `yaml:"local"`
	Host   string   `yaml:"host"`
	Secret string   `yaml:"secret"`
	Root   string   `yaml:"root"`  // static files root
	Index  []string `yaml:"index"` // default index page filenames

	localHost   string
	localScheme string
}

func (this *TunnelConfig) Validate() error {
	if len(this.Remote) == 0 {
		return errors.New("'remote' should not be empty")
	}
	if len(this.Local) == 0 && len(this.Root) == 0 {
		return errors.New("'local' or 'root' should not be empty")
	}
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
