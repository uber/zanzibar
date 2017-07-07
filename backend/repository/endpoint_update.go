package repository

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/codegen"
)

// EndpointConfigJSON is for serializing Endpoint into JSON file.
type EndpointConfigJSON struct {
	EndpointType     string                 `json:"endpointType"`
	EndpointID       string                 `json:"endpointId"`
	HandleID         string                 `json:"handleId"`
	ThriftFile       string                 `json:"thriftFile"`
	ThriftFileSha    string                 `json:"thriftFileSha"`
	ThriftMethodName string                 `json:"thriftMethodName"`
	WorkflowType     string                 `json:"workflowType"`
	ClientID         string                 `json:"clientID"`
	ClientMethod     string                 `json:"clientMethod"`
	TestFixtures     []string               `json:"testFixtures"`
	Middlewares      []MiddlewareConfigJSON `json:"middlewares"`
	ReqHeaderMap     map[string]string      `json:"reqHeaderMap"`
	ResHeaderMap     map[string]string      `json:"resHeaderMap"`
}

// MiddlewareConfigJSON is for serializing Endpoint Middleware configs into JSON file.
type MiddlewareConfigJSON struct {
	Name    string                 `json:"name"`
	Options map[string]interface{} `json:"options"`
}

// NewEndpointConfigJSON converts an EndpointConfig to its JSON format.
func NewEndpointConfigJSON(cfg *EndpointConfig) *EndpointConfigJSON {
	// TODO: Use middleware schemas to validate configs.
	midJsons := make([]MiddlewareConfigJSON, len(cfg.Middlewares))

	for i, mid := range cfg.Middlewares {
		midJsons[i].Name = mid.Name
		midJsons[i].Options = mid.Options
	}
	cfgJSON := &EndpointConfigJSON{
		EndpointType:     string(cfg.Type),
		EndpointID:       strings.TrimSuffix(cfg.ID, "."+cfg.HandleID),
		HandleID:         cfg.HandleID,
		ThriftFile:       cfg.ThriftFile,
		ThriftMethodName: cfg.ThriftServiceName + "::" + cfg.MethodName,
		WorkflowType:     cfg.WorkflowType,
		ClientID:         cfg.ClientID,
		ClientMethod:     cfg.ClientMethod,
		TestFixtures:     []string{},
		Middlewares:      midJsons,
		ReqHeaderMap:     map[string]string{},
		ResHeaderMap:     map[string]string{},
	}
	return cfgJSON
}

// WriteEndpointConfig writes endpoint configs into a runtime repository and
// also updates the meta json file for all endpoints.
func (r *Repository) WriteEndpointConfig(
	endpointCfgDir string,
	config *EndpointConfig,
	thriftFileSha string,
) error {
	if err := r.validateEndpointCfg(config); err != nil {
		return errors.Wrap(err, "invalid endpoint config")
	}
	r.Lock()
	defer r.Unlock()
	endpointDir := codegen.CamelToSnake(strings.TrimSuffix(config.ID, "."+config.HandleID))
	dir := filepath.Join(r.absPath(endpointCfgDir), endpointDir)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "failed to create endpoint config dir")
	}
	fileName := codegen.CamelToSnake(config.HandleID) + ".json"
	cfgJSON := NewEndpointConfigJSON(config)
	cfgJSON.ThriftFileSha = thriftFileSha
	err = writeToJSONFile(filepath.Join(dir, fileName), cfgJSON)
	if err != nil {
		return errors.Wrap(err, "failed to write to config file")
	}
	err = updateEndpointMetaJSON(dir, "endpoint-config.json", fileName, cfgJSON)
	if err != nil {
		return errors.Wrap(err, "failed to write endpoint-config.json")
	}
	return nil
}

func (r *Repository) validateEndpointCfg(req *EndpointConfig) error {
	gatewayConfig, err := r.LatestGatewayConfig()
	if err != nil {
		return errors.Wrap(err, "invalid configuration before updating endpoint")
	}
	clientCfg, ok := gatewayConfig.Clients[req.ClientID]
	if !ok {
		return errors.Errorf("can't find client %s", req.ClientID)
	}
	if clientCfg.Type == HTTP {
		req.WorkflowType = "httpClient"
	} else if clientCfg.Type == TCHANNEL {
		req.WorkflowType = "tchannelClient"
	} else {
		return errors.Errorf("client type %s is not supported", clientCfg.Type)
	}

	// Client method is the second part of <thrift service>::<method name> for tchannel client.
	// For http client, it is only <method name>.
	parts := strings.Split(req.ClientMethod, "::")
	req.ClientMethod = parts[len(parts)-1]
	return nil
}

// updateEndpointMetaJSON adds an endpoint in the meta json file or updates the config for an exsiting endpoint.
func updateEndpointMetaJSON(configDir, metaFile, newFile string, cfgJSON *EndpointConfigJSON) error {
	metaFilePath := filepath.Join(configDir, metaFile)
	fileContent := new(codegen.EndpointClassConfig)
	err := readJSONFile(metaFilePath, fileContent)
	if err != nil {
		return err
	}
	fileContent.Config.Endpoints, err = addToEndpointList(fileContent.Config.Endpoints, newFile, configDir)
	if err != nil {
		return err
	}
	if fileContent.Dependencies == nil {
		fileContent.Dependencies = make(map[string][]string)
	}
	if c := fileContent.Dependencies["client"]; !findString(cfgJSON.ClientID, c) {
		fileContent.Dependencies["client"] = append(c, cfgJSON.ClientID)
	}
	parts := strings.Split(cfgJSON.EndpointID, ".")
	fileContent.Name = parts[0]
	if fileContent.Type == "" {
		fileContent.Type = cfgJSON.EndpointType
	}
	return writeToJSONFile(metaFilePath, fileContent)
}

// addToEndpointList adds 'newFile' to the endpoint list if it doesn't exist.
func addToEndpointList(curEndpoints []string, newFile string, configDir string) ([]string, error) {
	newFilePath, err := filepath.Abs(filepath.Join(configDir, newFile))
	if err != nil {
		return nil, err
	}
	oldFilePaths := make([]string, len(curEndpoints))
	for i, path := range curEndpoints {
		file, err := filepath.Abs(filepath.Join(configDir, path))
		if err != nil {
			return nil, err
		}
		oldFilePaths[i] = file
	}
	if !findString(newFilePath, oldFilePaths) {
		curEndpoints = append(curEndpoints, newFile)
	}
	return curEndpoints, nil
}

func findString(target string, array []string) bool {
	for _, str := range array {
		if str == target {
			return true
		}
	}
	return false
}
