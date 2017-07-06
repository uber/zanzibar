package repository

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	testlib "github.com/uber/zanzibar/test/lib"
)

const (
	exampleGateway                     = "../../examples/example-gateway"
	clientCfgUpdateRequestFile         = "data/client/contacts_client_update.json"
	tchannelClientCfgUpdateRequestFile = "data/enigma_client_config_update.json"
)

func copyDir(src, dest string, ignoredPrefixes []string) error {
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil || src == path {
			return err
		}
		for _, prefix := range ignoredPrefixes {
			if strings.HasPrefix(path, prefix) {
				return nil
			}
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.Mkdir(destPath, info.Mode())
		}
		srcFile, err := os.Open(path)
		defer srcFile.Close()
		if err != nil {
			return err
		}
		destFile, err := os.Create(destPath)
		defer destFile.Close()
		if err != nil {
			return err
		}
		_, err = io.Copy(destFile, srcFile)
		return err
	}
	return filepath.Walk(src, walkFn)
}

func copyExample(t *testing.T) (string, error) {
	tempDir, err := ioutil.TempDir("", "example-gateway")
	if err != nil {
		return "", err
	}
	err = copyDir(exampleGateway, tempDir, []string{
		filepath.Join(exampleGateway, "build"),
		filepath.Join(exampleGateway, "middlewares"),
	})
	if err != nil {
		return "", err
	}
	t.Logf("Temp dir is created at %s\n", tempDir)
	return tempDir, nil
}

func TestUpdateClientConfig(t *testing.T) {
	req := &ClientConfig{}
	err := readJSONFile(clientCfgUpdateRequestFile, req)
	assert.NoError(t, err, "Failed to unmarshal client config.")
	tempDir, err := copyExample(t)
	if !assert.NoError(t, err, "Failed to copy example.") {
		return
	}
	r := &Repository{
		localDir: tempDir,
	}
	clientCfgDir := "clients"
	err = r.UpdateClientConfigs(req, clientCfgDir, "sha1")

	if !assert.NoError(t, err, "Failed to write client config.") {
		return
	}
	clientConfigActFile := filepath.Join(tempDir, clientCfgDir, "contacts", clientConfigFileName)
	actualClientCfg, err := ioutil.ReadFile(clientConfigActFile)
	if !assert.NoError(t, err, "Failed to read client config file.") {
		return
	}
	clientConfigExpFile := filepath.Join(exampleGateway, clientCfgDir, "contacts", clientConfigFileName)
	testlib.CompareGoldenFile(t, clientConfigExpFile, actualClientCfg)

	clientModuleCfg, err := ioutil.ReadFile(filepath.Join(tempDir, clientCfgDir, clientModuleFileName))
	if !assert.NoError(t, err, "Failed to read client module config file.") {
		return
	}
	clientModuleExpFile := filepath.Join(exampleGateway, clientCfgDir, clientModuleFileName)
	testlib.CompareGoldenFile(t, clientModuleExpFile, clientModuleCfg)

	productionJSON, err := ioutil.ReadFile(filepath.Join(tempDir, productionCfgJSONPath))
	if !assert.NoError(t, err, "Failed to read client production JSON config file.") {
		return
	}
	productionJSONExpFile := filepath.Join(exampleGateway, productionCfgJSONPath)
	testlib.CompareGoldenFile(t, productionJSONExpFile, productionJSON)
}
