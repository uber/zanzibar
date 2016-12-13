package config

import "time"

// HTTPClientConfig configures a http client.
type HTTPClientConfig struct {
	Port int32  `yaml:"port,omitempty"`
	IP   string `yaml:"ip,omitempty"`
}

// Config type
type Config struct {
	// Logging     log.Configuration
	// Metrics     metrics.Configuration
	// Jaeger      jaeger.Configuration
	// TChannel    xtchannel.Configuration
	Verbose     bool
	ServiceName string `yaml:"serviceName,omitempty"`
	IP          string `yaml:"ip,omitempty"`
	Port        int32  `yaml:"port,omitempty"`
	Logger      struct {
		FileName string `yaml:"fileName,omitempty"`
	}
	Metrics struct {
		Tally struct {
			FlushInterval time.Duration `yaml:"flushInterval,omitempty"`
			Service       string        `yaml:"service,omitempty"`
		}
		M3 struct {
			HostPort      string        `yaml:"hostPort,omitempty"`
			FlushInterval time.Duration `yaml:"flushInterval,omitempty"`
		}
	}
	Clients struct {
		Contacts  HTTPClientConfig
		GoogleNow HTTPClientConfig
	}
}
