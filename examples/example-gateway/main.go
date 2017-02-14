package main

import (
	"io/ioutil"
	"os"
	"path"
	"time"

	metrics "github.com/uber-go/tally/m3"
	zap "github.com/uber-go/zap"
	zanzibar "github.com/uber/zanzibar/runtime"
	yaml "gopkg.in/yaml.v2"

	"github.com/uber/zanzibar/examples/example-gateway/clients"
	exampleConfig "github.com/uber/zanzibar/examples/example-gateway/config"
	endpoints "github.com/uber/zanzibar/examples/example-gateway/endpoints"
)

const defaultM3MaxQueueSize = 10000
const defaultM3MaxPacketSize = 1440 // 1440kb in UDP M3MaxPacketSize
const defaultM3FlushInterval = 500 * time.Millisecond

func loadConfig(config *exampleConfig.Config) error {
	configRoot := os.Getenv("UBER_CONFIG_DIR")
	filePath := path.Join(configRoot, "production.yaml")

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	return nil
}

func main() {
	gatewayConfig := &exampleConfig.Config{}
	tempLogger := zap.New(
		zap.NewJSONEncoder(),
		zap.Output(os.Stderr),
	)

	if err := loadConfig(gatewayConfig); err != nil {
		tempLogger.Error("Error initializing configuration",
			zap.String("error", err.Error()),
		)
		os.Exit(1)
	}

	clientOpts := &clients.Options{}
	clientOpts.Contacts.IP = gatewayConfig.Clients.Contacts.IP
	clientOpts.Contacts.Port = gatewayConfig.Clients.Contacts.Port
	clientOpts.GoogleNow.IP = gatewayConfig.Clients.GoogleNow.IP
	clientOpts.GoogleNow.Port = gatewayConfig.Clients.GoogleNow.Port
	clients := clients.CreateClients(clientOpts)

	m3FlushIntervalConfig := gatewayConfig.Metrics.M3.FlushInterval
	var m3FlushInterval time.Duration
	if m3FlushIntervalConfig == 0 {
		m3FlushInterval = defaultM3FlushInterval
	} else {
		m3FlushInterval = m3FlushIntervalConfig
	}

	commonTags := map[string]string{"env": "example"}
	m3Backend, err := metrics.NewM3Backend(
		gatewayConfig.Metrics.M3.HostPort,
		gatewayConfig.Metrics.Tally.Service,
		commonTags, // default tags
		false,      // include host
		defaultM3MaxQueueSize,
		defaultM3MaxPacketSize,
		m3FlushInterval,
	)
	if err != nil {
		tempLogger.Error("Error initializing m3backend",
			zap.String("error", err.Error()),
		)
		os.Exit(1)
	}

	server, err := zanzibar.CreateGateway(&zanzibar.Options{
		IP:   gatewayConfig.IP,
		Port: gatewayConfig.Port,
		Logger: zanzibar.LoggerOptions{
			FileName: gatewayConfig.Logger.FileName,
		},
		Metrics: zanzibar.MetricsOptions{
			FlushInterval: gatewayConfig.Metrics.Tally.FlushInterval,
			Service:       gatewayConfig.Metrics.Tally.Service,
		},
		TChannel: zanzibar.TChannelOptions{
			ServiceName: gatewayConfig.TChannel.ServiceName,
			ProcessName: gatewayConfig.TChannel.ProcessName,
		},

		Clients:        clients,
		MetricsBackend: m3Backend,
	})
	if err != nil {
		panic(err)
	}

	err = server.Bootstrap(endpoints.Register)
	if err != nil {
		panic(err)
	}

	server.Logger.Info("Started EdgeGateway",
		zap.String("realAddr", server.RealAddr),
		zap.Object("config", gatewayConfig),
	)

	// TODO: handle sigterm gracefully

	server.Wait()

	// log.Configure(&cfg.Logging, cfg.Verbose)

	// metrics, err := cfg.Metrics.New()
	// if err != nil {
	// 	log.Fatalf("Could not connect to metrics: %v", err)
	// }
	// metrics.Counter("boot").Inc(1)

	// closer, err := xjaeger.InitGlobalTracer(cfg.Jaeger, cfg.ServiceName, metrics)
	// if err != nil {
	// 	log.Fatalf("Jaeger.InitGlobalTracer failed: %v", err)
	// }
	// defer closer.Close()
}
