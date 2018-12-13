package zanzibar

import (
	"time"
)

type TChannelClientConfig struct {
	ServiceName       string        `yaml:"serviceName"`
	RoutingKey        string        `yaml:"routingKey`
	IP                string        `yaml:"ip"`
	Port              int           `yaml:"port"`
	Timeout           time.Duration `yaml:"timeout"`
	TimeoutPerAttempt time.Duration `yaml:"timeoutPerAttempt"`

	Staging StagingConfig `yaml:"staging"`
}

type SidecarConfig struct {
	TChannel struct {
		IP   string `yaml:"ip"`
		Port int    `yaml:"port"`
	} `yaml:"tchannel"`
	HTTP struct {
		IP           string `yaml:"ip"`
		Port         int    `yaml:"port"`
		CallerHeader string `yaml:"callerHeader`
		CalleeHeader string `yaml:"calleeHeader`
	} `yaml:"http`
}

type StagingConfig struct {
	ServiceName string `yaml:"serviceName"`
	IP          string `yaml:"ip"`
	Port        int    `yaml:"port"`
}

type HTTPClientConfig struct {
	// ServiceName is the calleee service name
	ServiceName    string            `yaml:"serviceName"`
	Timeout        time.Duration     `yaml:"timeout"`
	DefaultHeaders map[string]string `yaml:"defaultHeaders"`
	IP             string            `yaml:"ip"`
	Port           int               `yaml:"port"`
}

type TChannelConfig struct {
	Port int `yaml:"port"`
	ProcessName string `yaml:"processName"`
	RoutingKey string `yaml:"routingKey`
	ServiceName string `yaml:"serviceName"`
	Deputy struct {
		Timeout time.Duration `yaml:"timeout"`
		TimeoutPerAttempt time.Duration `yaml:"timeoutPerAttempt"`
	} `yaml:"deputy`
}

