package config

import (
	"errors"
	"net"

	"github.com/BurntSushi/toml"
)

// DNSSettings described general settings of DNS resolver
type DNSSettings struct {
	CustomECS   []net.IP `toml:"custom_ecs"`
	NoECS       bool     `toml:"no_ecs"`
	UserAgent   string   `toml:"user_agent"`
	NoUserAgent bool     `toml:"no_user_agent"`
}

type typeCustomSpecified struct {
	Domain []string `toml:"domain"`
	Suffix []string `toml:"suffix"`
}

type typeGeneralConfig struct {
	Listen     []string  `toml:"listen"`
	Timeout    uint      `toml:"timeout"`     // seconds
	RoundRobin Selectors `toml:"round_robin"` // default: clock
	DNSSettings
}

type typeUpstreamHTTPS struct {
	Host     []string `toml:"host"`
	Port     uint16   `toml:"port"` // default: 443
	Hostname string   `toml:"hostname"`
	Path     string   `toml:"path"` // default: /dns-query
	Google   bool     `toml:"google"`
	Cookie   bool     `toml:"cookie"`
	Weight   int32    `toml:"weight"` // default: 1
	typeCustomSpecified
	DNSSettings
}

type typeUpstreamTLS struct {
	Host     []string `toml:"host"`
	Port     uint16   `toml:"port"` // default: 853
	Hostname string   `toml:"hostname"`
	Weight   int32    `toml:"weight"` // default: 1
	typeCustomSpecified
	DNSSettings
}

type typeTraditional struct {
	Host      []string `toml:"host"`
	Port      uint16   `toml:"port"` // default: 53
	Bootstrap bool     `toml:"bootstrap"`
	Weight    int32    `toml:"weight"` // default: 1
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
	}

	if config.Config.RoundRobin == "" {
		config.Config.RoundRobin = SelectorClock
	}

	for index := range config.HTTPS {
		https := &config.HTTPS[index]
		if https.Port == 0 {
			https.Port = 443
		}
		if https.Path == "" {
			https.Path = "/dns-query"
		}
		if https.Weight < 1 {
			https.Weight = 1
		}
	}

	for index := range config.TLS {
		tls := &config.TLS[index]
		if tls.Port == 0 {
			tls.Port = 853
		}
		if tls.Weight < 1 {
			tls.Weight = 1
		}
	}

	for index := range config.Traditional {
		traditional := &config.Traditional[index]
		if traditional.Port == 0 {
			traditional.Port = 53
		}
		if traditional.Weight < 1 {
			traditional.Weight = 1
		}
	}

	return
}
