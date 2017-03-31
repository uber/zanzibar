// Code generated by zanzibar
// @generated

package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/uber-go/zap"
	tchannel "github.com/uber/tchannel-go"
	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints"
	"github.com/uber/zanzibar/runtime"
)

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)
	return zanzibar.GetDirnameFromRuntimeCaller(file)
}

func getConfigDirName() string {
	return filepath.Join(
		getDirName(),
		"..",
		"config",
	)
}

func getConfig(seedConfig map[string]interface{}) *zanzibar.StaticConfig {
	return zanzibar.NewStaticConfigOrDie([]string{
		filepath.Join(getDirName(), "zanzibar-defaults.json"),
		filepath.Join(getConfigDirName(), "production.json"),
		filepath.Join(os.Getenv("CONFIG_DIR"), "production.json"),
	}, seedConfig)
}

func createGateway() (*zanzibar.Gateway, error) {
	listenIP, err := tchannel.ListenIP()
	if err != nil {
		return nil, err
	}
	config := getConfig(map[string]interface{}{
		"ip": listenIP.String(),
	})
	clients := clients.CreateClients(config)
	return zanzibar.CreateGateway(config, &zanzibar.Options{
		Clients: clients,
	})
}

func logAndWait(server *zanzibar.Gateway) {
	server.Logger.Info("Started ExampleGateway",
		zap.String("realAddr", server.RealAddr),
		zap.Object("config", server.InspectOrDie()),
	)

	// TODO: handle sigterm gracefully
	server.Wait()
	// TODO: emit metrics about startup.
	// TODO: setup and configure tracing/jeager.
}

func main() {
	server, err := createGateway()
	if err != nil {
		panic(err)
	}

	err = server.Bootstrap(endpoints.Register)
	if err != nil {
		panic(err)
	}
	logAndWait(server)
}
