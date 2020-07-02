package config

import (
	"errors"
	"net"

	"github.com/BurntSushi/toml"
)

// DNSSettings described general settings of DNS resolver
type DNSSettings struct {
	CustomECS []net.IP `toml:"custom_ecs"`
	NoECS     bool     `toml:"no_ecs"`
}

type typeCustomSpecified struct {
	Domain []string `toml:"domain"`
	Suffix []string `toml:"suffix"`
}

type typeGeneralConfig struct {
	Listen  []string `toml:"listen"`
	Timeout uint     `toml:"timeout"` // seconds
	DNSSettings
}

type typeUpstreamHTTPS struct {
	Host     []string `toml:"host"`
	Port     uint16   `toml:"port"` // default: 443
	Hostname string   `toml:"hostname"`
	Path     string   `toml:"path"` // default: /dns-query
	Google   bool     `toml:"google"`
	Cookie   bool     `toml:"cookie"`
	typeCustomSpecified
	DNSSettings
}

type typeUpstreamTLS struct {
	Host     []string `toml:"host"`
	Port     uint16   `toml:"port"` // default: 853
	Hostname string   `toml:"hostname"`
	typeCustomSpecified
	DNSSettings
}

type typeTraditional struct {
	Host      []string `toml:"host"`
	Port      uint16   `toml:"port"` // default: 53
	Bootstrap bool     `toml:"bootstrap"`
	typeCustomSpecified
	DNSSettings
}

// Config described user configuration
type Config struct {
	Config      typeGeneralConfig              `toml:"config"`
	HTTPS       []typeUpstreamHTTPS            `toml:"https"`
	TLS         []typeUpstreamTLS              `toml:"tls"`
	Traditional []typeTraditional              `toml:"traditional"`
	Hosts       map[string]map[string][]string `toml:"hosts"`
}

// LoadConfig from configuration file
func LoadConfig(configPath string) (config *Config, err error) {
	config = &Config{}
	_, err = toml.DecodeFile(configPath, &config)
	if err != nil {
		return
	}

	if len(config.Config.Listen) == 0 {
		err = errors.New("no listen address")
		return
	} else if len(config.HTTPS)+len(config.TLS) == 0 {
		err = errors.New("no available upstream")
		return
	}

	for index := range config.HTTPS {
		https := &config.HTTPS[index]
		if https.Port == 0 {
			https.Port = 443
		}
		if https.Path == "" {
			https.Path = "/dns-query"
		}
	}

	for index := range config.TLS {
		tls := &config.TLS[index]
		if tls.Port == 0 {
			tls.Port = 853
		}
	}

	for index := range config.Traditional {
		traditional := &config.Traditional[index]
		if traditional.Port == 0 {
			traditional.Port = 53
		}
	}

	return
}
