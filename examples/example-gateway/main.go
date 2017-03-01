package main

import (
	"os"
	"path"
	"path/filepath"

	"github.com/uber-go/zap"
	"github.com/uber/zanzibar/examples/example-gateway/clients"
	"github.com/uber/zanzibar/examples/example-gateway/endpoints"
	"github.com/uber/zanzibar/runtime"
)

func getProjectDir() string {
	goPath := os.Getenv("GOPATH")
	return path.Join(goPath, "src", "github.com", "uber", "zanzibar")
}

func main() {
	config := zanzibar.NewStaticConfigOrDie([]string{
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

	server, err := zanzibar.CreateGateway(config, &zanzibar.Options{
		Clients: clients,
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
		zap.Object("config", config.InspectOrDie()),
	)

	// TODO: handle sigterm gracefully

	server.Wait()

	// TODO: emit metrics about startup.
	// TODO: setup and configure tracing/jeager.
}
