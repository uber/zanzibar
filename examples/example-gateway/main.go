package main

import (
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/uber-go/tally/m3"
	"github.com/uber-go/zap"
	"github.com/uber/zanzibar/examples/example-gateway/clients"
	"github.com/uber/zanzibar/examples/example-gateway/endpoints"
	"github.com/uber/zanzibar/runtime"
)

const defaultM3MaxQueueSize = 10000
const defaultM3MaxPacketSize = 1440 // 1440kb in UDP M3MaxPacketSize
const defaultM3FlushInterval = 500 * time.Millisecond

func getProjectDir() string {
	goPath := os.Getenv("GOPATH")
	return path.Join(goPath, "src", "github.com", "uber", "zanzibar")
}

func main() {
	tempLogger := zap.New(
		zap.NewJSONEncoder(),
		zap.Output(os.Stderr),
	)

	config := zanzibar.NewStaticConfig([]string{
		filepath.Join(getProjectDir(), "config", "production.json"),
		filepath.Join(
			getProjectDir(),
			"examples",
			"example-gateway",
			"config",
			"production.json",
		),
		filepath.Join(os.Getenv("CONFIG_DIR"), "production.json"),
	}, nil)

	clients := clients.CreateClients(config)

	m3FlushIntervalConfig := config.GetInt("metrics.m3.flushInterval")

	commonTags := map[string]string{"env": "example"}
	m3Backend, err := metrics.NewM3Backend(
		config.GetString("metrics.m3.hostPort"),
		config.GetString("metrics.tally.service"),
		commonTags, // default tags
		false,      // include host
		defaultM3MaxQueueSize,
		defaultM3MaxPacketSize,
		time.Duration(m3FlushIntervalConfig)*time.Millisecond,
	)
	if err != nil {
		tempLogger.Error("Error initializing m3backend",
			zap.String("error", err.Error()),
		)
		os.Exit(1)
	}

	server, err := zanzibar.CreateGateway(config, &zanzibar.Options{
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
		zap.Object("config", config.Inspect()),
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
