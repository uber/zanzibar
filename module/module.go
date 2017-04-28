// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package module

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// moduleType enum defines whether a ModuleClass is a singleton or contains
// multiple directories with multiple configurations
type classType int

const (
	// SingleModule defines a module class type that has 1 directory
	SingleModule classType = iota
	// MultiModule defines a module class type with multiple nested directories
	MultiModule classType = iota
)

const configSuffix = "-config.json"

// NewSystem returns a new module system
func NewSystem() *System {
	return &System{
		classes:    map[string]*Class{},
		classOrder: []string{},
	}
}

// System defines the module classes and their type generators
type System struct {
	classes    map[string]*Class
	classOrder []string
}

// RegisterClass defines a class of module in the module system
// For example, an "Endpoint" class or a "Client" class
func (moduleSystem *System) RegisterClass(name string, class Class) error {
	if name == "" {
		return errors.Errorf("A module class name must not be empty")
	}

	if moduleSystem.classes[name] != nil {
		return errors.Errorf(
			"The module class \"%s\" is already defined",
			name,
		)
	}

	// Validate the module class dependencies
	// (this validation ensures that circular deps cannot exist)
	for _, moduleType := range class.ClassDependencies {
		if moduleSystem.classes[moduleType] == nil {
			return errors.Errorf(
				"The module class \"%s\" depends on class type \"%s\", "+
					"which is not yet defined",
				name,
				moduleType,
			)
		}
	}

	class.Directory = filepath.Clean(class.Directory)

	if strings.HasPrefix(class.Directory, "..") {
		return errors.Errorf(
			"The module class \"%s\" must map to an internal directory but was \"%s\"",
			name,
			class.Directory,
		)
	}

	// Validate the module class directory name is unique
	for moduleClassName, moduleClass := range moduleSystem.classes {
		if class.Directory == moduleClass.Directory {
			return errors.Errorf(
				"The module class \"%s\" conflicts with directory \"%s\" from class \"%s\"",
				name,
				class.Directory,
				moduleClassName,
			)
		}
	}

	class.types = map[string]BuildGenerator{}
	moduleSystem.classes[name] = &class
	moduleSystem.classOrder = append(moduleSystem.classOrder, name)

	return nil
}

// RegisterClassType registers a type generator for a specific module class
// For example, the "http"" type generator for the "Endpoint"" class
func (moduleSystem *System) RegisterClassType(
	className string,
	classType string,
	generator BuildGenerator,
) error {
	moduleClass := moduleSystem.classes[className]

	if moduleClass == nil {
		return errors.Errorf(
			"Cannot set class type \"%s\" for undefined class \"%s\"",
			classType,
			className,
		)
	}

	if moduleClass.types[classType] != nil {
		return errors.Errorf(
			"The class type \"%s\" is already defined for class \"%s\"",
			classType,
			className,
		)
	}

	moduleClass.types[classType] = generator

	return nil
}

// ResolveModules resolves the module instances from the config on disk
// Using the system class and type definitions, the class directories are
// walked, and a module instance is initialized for each identified module in
// the target directory.
func (moduleSystem *System) ResolveModules(
	baseDirectory string,
) (map[string][]*Instance, error) {

	resolvedModules := map[string][]*Instance{}

	for _, className := range moduleSystem.classOrder {
		class := moduleSystem.classes[className]
		fullInstanceDirectory := filepath.Join(baseDirectory, class.Directory)

		classInstances := []*Instance{}

		if class.ClassType == SingleModule {
			instance, instanceErr := readInstance(
				className,
				baseDirectory,
				class.Directory,
			)
			if instanceErr != nil {
				return nil, errors.Wrapf(
					instanceErr,
					"Error reading single instance \"%s\" in \"%s\"",
					className,
					class.Directory,
				)
			}
			classInstances = append(classInstances, instance)
		} else {

			files, err := ioutil.ReadDir(fullInstanceDirectory)

			if err != nil {
				// TODO: We should accumulate errors and list them all here
				// Expected $path to be a class directory
				return nil, errors.Wrapf(
					err,
					"Error reading module instance directory \"%s\"",
					fullInstanceDirectory,
				)
			}

			for _, file := range files {
				if file.IsDir() {
					instance, instanceErr := readInstance(
						className,
						baseDirectory,
						filepath.Join(class.Directory, file.Name()),
					)
					if instanceErr != nil {
						return nil, errors.Wrapf(
							instanceErr,
							"Error reading multi instance \"%s\" in \"%s\"",
							className,
							filepath.Join(class.Directory, file.Name()),
						)
					}
					classInstances = append(classInstances, instance)
				}
			}
		}

		resolvedModules[className] = classInstances
	}

	return resolvedModules, nil
}

func readInstance(
	className string,
	baseDirectory string,
	instanceDirectory string,
) (*Instance, error) {

	jsonFileName := className + configSuffix
	classConfigPath := filepath.Join(
		baseDirectory,
		instanceDirectory,
		jsonFileName,
	)

	jsonConfig := JSONClassConfig{}
	raw, err := jsonConfig.Read(classConfigPath)

	if err != nil {
		// TODO: We should accumulate errors and list them all here
		// Expected $class-config.json to exist in ...
		return nil, errors.Wrapf(
			err,
			"Error reading JSON Config \"%s\"",
			classConfigPath,
		)
	}

	return &Instance{
		ClassName:     className,
		ClassType:     jsonConfig.Type,
		BaseDirectory: baseDirectory,
		Directory:     instanceDirectory,
		InstanceName:  jsonConfig.Name,
		Dependencies:  readDeps(jsonConfig.Dependencies),
		JSONFileName:  jsonFileName,
		JSONFileRaw:   raw,
	}, nil
}

func readDeps(jsonDeps map[string][]string) []Dependency {
	depCount := 0

	for _, depsList := range jsonDeps {
		depCount += len(depsList)
	}

	deps := make([]Dependency, depCount)
	depIndex := 0

	for className, depsList := range jsonDeps {
		for _, instanceName := range depsList {
			deps[depIndex] = Dependency{
				ClassName:    className,
				InstanceName: instanceName,
			}
			depIndex++
		}
	}

	return deps
}

// GenerateBuild will, given a module system configuration directory and a
// target build directory, run the generators assigned to each type of module
// and write the generated output to the module build directory
func (moduleSystem *System) GenerateBuild(
	baseDirectory string,
	targetGenDir string,
) error {
	resolvedModules, err := moduleSystem.ResolveModules(baseDirectory)

	if err != nil {
		return err
	}

	moduleCount := 0
	for _, moduleList := range resolvedModules {
		moduleCount += len(moduleList)
	}

	moduleIndex := 0
	for _, className := range moduleSystem.classOrder {
		classInstances := resolvedModules[className]

		for _, classInstance := range classInstances {
			moduleIndex++
			buildPath := filepath.Join(
				targetGenDir,
				classInstance.Directory,
			)
			prettyBuildPath := filepath.Join(
				".",
				filepath.Base(targetGenDir),
				classInstance.Directory,
			)
			fmt.Printf(
				"Generating %8s %8s %-10s in %-30s %d/%d\n",
				classInstance.ClassType,
				classInstance.ClassName,
				classInstance.InstanceName,
				prettyBuildPath,
				moduleIndex,
				moduleCount,
			)

			classGenerators := moduleSystem.classes[classInstance.ClassName]
			generator := classGenerators.types[classInstance.ClassType]

			if generator == nil {
				fmt.Printf(
					"Skipping generation of \"%s\" \"%s\" class of type \"%s\" "+
						"as generator is not defined\n",
					classInstance.InstanceName,
					classInstance.ClassName,
					classInstance.ClassType,
				)
				continue
			}

			files, err := generator.Generate(classInstance)

			if err != nil {
				fmt.Printf(
					"Error generating \"%s\" \"%s\" class of type \"%s\"\n%s\n",
					classInstance.InstanceName,
					classInstance.ClassName,
					classInstance.ClassType,
					err.Error(),
				)
				return err
			}

			for filePath, content := range files {
				filePath = filepath.Clean(filePath)

				if strings.HasPrefix(filePath, "..") {
					return errors.Errorf(
						"Module \"%s\" generated a file outside the build dir \"%s\"",
						classInstance.Directory,
						filePath,
					)
				}

				resolvedPath := filepath.Join(
					buildPath,
					filePath,
				)

				if err := writeFile(resolvedPath, content); err != nil {
					return errors.Wrapf(
						err,
						"Error writing to file \"%s\"",
						resolvedPath,
					)
				}

				// HACK: The module system writer shouldn't
				// assume that we want to format the files in
				// this way, but we don't have these formatters
				// as a library or a custom post build script
				// for the generators yet.
				if filepath.Ext(filePath) == ".go" {
					if err := formatGoFile(resolvedPath); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func formatGoFile(filePath string) error {
	gofmtCmd := exec.Command("gofmt", "-s", "-w", "-e", filePath)
	gofmtCmd.Stdout = os.Stdout
	gofmtCmd.Stderr = os.Stderr

	if err := gofmtCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to gofmt file: \"%s\"", filePath)
	}

	goimportsCmd := exec.Command("goimports", "-w", "-e", filePath)
	goimportsCmd.Stdout = os.Stdout
	goimportsCmd.Stderr = os.Stderr

	if err := goimportsCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to goimports file: %q", filePath)
	}

	return nil
}

// Class defines a module class in the build configuration directory. This
// coud be something like an Endpoint class which contains multiple endpoint
// configurations, or a Lib class, that is itself a module instance
type Class struct {
	ClassType         classType
	Directory         string
	ClassDependencies []string
	types             map[string]BuildGenerator
}

// BuildGenerator provides a function to generate a module instance build
// artifact from its configuration as part of a build step. For example, an
// Endpoint module instance may generate endpoint handler code
type BuildGenerator interface {
	Generate(
		instance *Instance,
	) (map[string][]byte, error)
}

// Instance is a configured module on disk inside a module class directory.
// For example, this could be
//     ClassName:    "Endpoint,
//     ClassType:    "http",
//     BaseDirectory "/path/to/service/base/"
//     Directory:    "clients/health/"
//     InstanceName: "health",
type Instance struct {
	// ClassName is the name of the class as defined in the module system
	ClassName string
	// ClassType is the type of the class as defined in the module system
	ClassType string
	// BaseDirectory is the absolute path to module system system top level
	// directory
	BaseDirectory string
	// Directory is the relative instance directory
	Directory string
	// InstanceName is the name of the instance as configured in the instance's
	// json file
	InstanceName string
	// Config is a reference to the instance "config" key in the instances json
	//file
	Config interface{}
	// Dependency is a list of dependent modules as defined in the instances
	// json file
	Dependencies []Dependency
	// Resolved dependencies is a list of dependent modules after processing
	// (fully resolved)
	ResolvedDependencies []Dependency
	// The JSONFileName is file name of the instance json file
	JSONFileName string
	// JSONFileRaw is the raw JSON file read as bytes used for future parsing
	JSONFileRaw []byte
}

// Dependency defines a module instance required by another module instance
type Dependency struct {
	ClassName    string
	InstanceName string
}

// JSONClassConfig maps onto a json configuration for a class type
type JSONClassConfig struct {
	Name         string              `json:"name"`
	Config       interface{}         `json:"config"`
	Dependencies map[string][]string `json:"dependencies"`
	Type         string              `json:"type"`
}

// Read will read a class configuration json file into a jsonClassConfig struct
// or return an error if it cannot be unmarshaled into the struct
func (jsonConfig *JSONClassConfig) Read(
	classConfigPath string,
) ([]byte, error) {
	configFile, readErr := ioutil.ReadFile(classConfigPath)
	if readErr != nil {
		return nil, errors.Wrapf(
			readErr,
			"Error reading class config %q",
			classConfigPath,
		)
	}

	parseErr := json.Unmarshal(configFile, &jsonConfig)

	if parseErr != nil {
		return nil, errors.Wrapf(
			parseErr,
			"Error JSON parsing clss config %q",
			configFile,
		)
	}

	if jsonConfig.Name == "" {
		return nil, errors.Errorf(
			"Error reading instance name from %q",
			classConfigPath,
		)
	}

	if jsonConfig.Type == "" {
		return nil, errors.Errorf(
			"Error reading instance type from %q",
			classConfigPath,
		)
	}

	if jsonConfig.Dependencies == nil {
		jsonConfig.Dependencies = map[string][]string{}
	}

	return configFile, nil
}

// writeFile is like ioutil.WriteFile with a mkdirp step
func writeFile(filePath string, bytes []byte) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return errors.Wrapf(
				err, "could not make directory: %q", filePath,
			)
		}
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	defer closeFile(file)
	if err != nil {
		return errors.Wrapf(
			err, "Could not open file for writing: %q", filePath,
		)
	}

	n, err := file.Write(bytes)

	if err != nil {
		return errors.Wrapf(err, "Error writing to file %q", filePath)
	}

	if n != len(bytes) {
		return errors.Wrapf(
			err,
			"Error writing full contents to file: %q",
			filePath,
		)
	}

	return nil
}

func closeFile(file *os.File) {
	_ = file.Close()
}
