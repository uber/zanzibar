// Copyright (c) 2020 Uber Technologies, Inc.
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

package codegen

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	yaml "github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/uber/zanzibar/parallelize"
)

// moduleType enum defines whether a ModuleClass is a singleton or contains
// multiple directories with multiple configurations
type moduleClassType int

const (
	// SingleModule defines a module class type that has 1 directory
	SingleModule moduleClassType = iota
	// MultiModule defines a module class type with multiple nested directories
	MultiModule moduleClassType = iota
)

const jsonConfigSuffix = "-config.json"
const yamlConfigSuffix = "-config.yaml"

// NewModuleSystem returns a new module system
func NewModuleSystem(
	moduleSearchPaths map[string][]string,
	defaultDependencies map[string][]string,
	selectiveBuilding bool,
	postGenHook ...PostGenHook,
) *ModuleSystem {
	return &ModuleSystem{
		classes:             map[string]*ModuleClass{},
		classOrder:          []string{},
		postGenHook:         postGenHook,
		moduleSearchPaths:   moduleSearchPaths,
		defaultDependencies: defaultDependencies,
		selectiveBuilding:   selectiveBuilding,
	}
}

// ModuleSystem defines the module classes and their type generators
type ModuleSystem struct {
	classes             map[string]*ModuleClass
	classOrder          []string
	postGenHook         []PostGenHook
	moduleSearchPaths   map[string][]string
	defaultDependencies map[string][]string
	selectiveBuilding   bool
}

// PostGenHook provides a way to do work after the build is generated,
// useful to augment the build, e.g. generate mocks after interfaces are generated
type PostGenHook func(map[string][]*ModuleInstance) error

// RegisterClass defines a class of module in the module system
// For example, an "Endpoint" class or a "Client" class
func (system *ModuleSystem) RegisterClass(class ModuleClass) error {
	name := class.Name
	if name == "" {
		return errors.Errorf("A module class name must not be empty")
	}

	if class.NamePlural == "" {
		return errors.Errorf("A module class name plural must not be empty")
	}

	if system.classes[name] != nil {
		return errors.Errorf("Module class %q is already defined", name)
	}

	class.types = map[string]BuildGenerator{}
	system.classes[name] = &class

	return nil
}

// dedup returns a slice containing unique sorted elements of the given slice
func dedup(array []string) []string {
	dict := map[string]bool{}
	for _, elt := range array {
		dict[elt] = true
	}

	i := 0
	unique := make([]string, len(dict))
	for key := range dict {
		unique[i] = key
		i++
	}
	sort.Strings(unique)
	return unique
}

// validateDir checks if the module class can map to the given dir
func (system *ModuleSystem) validateClassDir(class *ModuleClass, dir string) error {
	dir = filepath.Clean(dir)

	if strings.HasPrefix(dir, "..") {
		return errors.Errorf(
			"Module class %q must map to internal directories but found %q",
			class.Name,
			dir,
		)
	}

	return nil
}

// RegisterClassType registers a type generator for a specific module class
// For example, the "http"" type generator for the "Endpoint"" class
func (system *ModuleSystem) RegisterClassType(
	className string,
	classType string,
	generator BuildGenerator,
) error {
	moduleClass := system.classes[className]

	if moduleClass == nil {
		return errors.Errorf(
			"Cannot set class type %q for undefined class %q",
			classType,
			className,
		)
	}

	if moduleClass.types[classType] != nil {
		return errors.Errorf(
			"The class type %q is already defined for class %q",
			classType,
			className,
		)
	}

	moduleClass.types[classType] = generator

	return nil
}

func (system *ModuleSystem) populateResolvedDependencies(
	classInstances map[string]*ModuleInstance,
	resolvedModules map[string]map[string]*ModuleInstance,
) error {
	// Resolve the class dependencies
	for _, classInstance := range classInstances {

		for _, classDependency := range classInstance.Dependencies {
			moduleClassInstances, ok :=
				resolvedModules[classDependency.ClassName]

			if !ok {
				return errors.Errorf(
					"Invalid class name %q in dependencies for %q %q",
					classDependency.ClassName,
					classInstance.ClassName,
					classInstance.InstanceName,
				)
			}

			dependencyInstance := moduleClassInstances[classDependency.InstanceName]

			if dependencyInstance == nil {
				return errors.Errorf(
					"Unknown %q class dependency %q "+
						"in dependencies for %q %q",
					classDependency.ClassName,
					classDependency.InstanceName,
					classInstance.ClassName,
					classInstance.InstanceName,
				)
			}

			resolvedDependencies, ok :=
				classInstance.ResolvedDependencies[classDependency.ClassName]

			if !ok {
				resolvedDependencies = []*ModuleInstance{}
			}

			classInstance.ResolvedDependencies[classDependency.ClassName] =
				appendUniqueModule(resolvedDependencies, dependencyInstance)
		}

		// Sort the dependencies for deterministic code generation
		for className, deps := range classInstance.ResolvedDependencies {
			sortedModuleList, err := sortDependencyList(
				className,
				deps,
			)
			if err != nil {
				return err
			}
			classInstance.ResolvedDependencies[className] = sortedModuleList
		}
	}

	return nil
}

func appendUniqueModule(
	classDeps []*ModuleInstance,
	instance *ModuleInstance,
) []*ModuleInstance {
	for i, classInstance := range classDeps {
		if classInstance.InstanceName == instance.InstanceName {
			classDeps[i] = instance
			return classDeps
		}
	}

	return append(classDeps, instance)
}

func (system *ModuleSystem) populateRecursiveDependencies(
	instances map[string]*ModuleInstance,
) error {
	for _, classInstance := range instances {
		recursiveDeps := map[string]map[string]*ModuleInstance{}
		err := resolveRecursiveDependencies(
			classInstance,
			recursiveDeps,
		)
		if err != nil {
			return err
		}

		for _, className := range system.classOrder {
			moduleMap, ok := recursiveDeps[className]

			if !ok {
				continue
			}

			// If a glob pattern matches the same module class multiple times, make sure that DependencyOrder
			// remains the unique list of classes.

			found := false
			for _, dependencyOrderClassName := range classInstance.DependencyOrder {
				if dependencyOrderClassName == className {
					found = true
					break
				}
			}
			if !found {
				classInstance.DependencyOrder = append(
					classInstance.DependencyOrder,
					className,
				)
			}

			moduleList := make([]*ModuleInstance, len(moduleMap))
			index := 0
			for _, moduleInstance := range moduleMap {
				moduleList[index] = moduleInstance
				index++
			}

			sortedModuleList, err := sortDependencyList(className, moduleList)
			if err != nil {
				return err
			}

			classInstance.RecursiveDependencies[className] = sortedModuleList
		}
	}

	return nil
}

func resolveRecursiveDependencies(
	instance *ModuleInstance,
	resolvedDeps map[string]map[string]*ModuleInstance,
) error {
	for className, depList := range instance.ResolvedDependencies {
		classDeps := resolvedDeps[className]

		if classDeps == nil {
			classDeps = map[string]*ModuleInstance{}
			resolvedDeps[className] = classDeps
		}

		for _, dep := range depList {
			if classDeps[dep.InstanceName] == nil {
				classDeps[dep.InstanceName] = dep
				err := resolveRecursiveDependencies(dep, resolvedDeps)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type sortableDependencyList []*ModuleInstance

func (s sortableDependencyList) Len() int {
	return len(s)
}

func (s sortableDependencyList) Less(i int, j int) bool {
	return s[i].InstanceName < s[j].InstanceName
}

func (s sortableDependencyList) Swap(i int, j int) {
	s[i], s[j] = s[j], s[i]
}

func sortDependencyList(
	className string,
	instances []*ModuleInstance,
) ([]*ModuleInstance, error) {
	instanceList := sortableDependencyList(instances[:])
	sort.Sort(instanceList)
	sorted := make([]*ModuleInstance, len(instances))

	for i, instance := range instances {
		insertIndex := i
		for j := 0; j < i; j++ {
			if insertIndex == i {
				if peerDepends(sorted[j], instance) {
					insertIndex = j
				}
			}

			if insertIndex != i {
				if peerDepends(instance, sorted[j]) {
					return nil, errors.Errorf(
						"Dependency cycle: %s cannot be initialized before %s",
						sorted[j].InstanceName,
						instance.InstanceName,
					)
				}
			}
		}

		for shuffle := i; shuffle > insertIndex; shuffle-- {
			sorted[shuffle] = sorted[shuffle-1]
		}

		sorted[insertIndex] = instance
	}

	return sorted, nil
}

// peerDepends returns true if module a and module b have the same class name
// and a requires b
func peerDepends(a *ModuleInstance, b *ModuleInstance) bool {
	if a.ClassName != b.ClassName {
		return false
	}

	for _, dependency := range a.RecursiveDependencies[a.ClassName] {
		if dependency.InstanceName == b.InstanceName &&
			dependency.ClassName == b.ClassName {
			return true
		}
	}

	return false
}

type byName []*ModuleClass

func (n byName) Len() int           { return len(n) }
func (n byName) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }
func (n byName) Less(i, j int) bool { return n[i].Name < n[j].Name }

// sort the module classes by class height in DAG, lowest node first
func sortModuleClasses(classes []*ModuleClass) ([]*ModuleClass, error) {
	heightMap := map[*ModuleClass]int{}
	sortedGroup := [][]*ModuleClass{}

	sort.Sort(byName(classes))
	for _, class := range classes {
		h, err := height(class, heightMap, []*ModuleClass{})
		if err != nil {
			return nil, err
		}
		l := len(sortedGroup)
		if l < h+1 {
			for i := 0; i < h+1-l; i++ {
				sortedGroup = append(sortedGroup, []*ModuleClass{})
			}
		}
		sortedGroup[h] = append(sortedGroup[h], class)
	}

	sorted := []*ModuleClass{}
	for _, g := range sortedGroup {
		sort.Sort(byName(g))
		sorted = append(sorted, g...)
	}
	return sorted, nil
}

// height marks the height of each module class height in the dependency tree
func height(i *ModuleClass, known map[*ModuleClass]int, seen []*ModuleClass) (int, error) {
	// detect dependency cycle
	for _, class := range seen {
		if i == class {
			var path string
			for _, c := range seen {
				path += c.Name + "->"
			}
			path += i.Name
			return 0, fmt.Errorf("dependency cycle detected for module class %q: %s", i.Name, path)
		}
	}

	if h, ok := known[i]; ok {
		return h, nil
	}
	if len(i.dependentClasses) == 0 {
		known[i] = 0
		return 0, nil
	}

	mh := 0
	seen = append(seen, i)
	for _, instance := range i.dependentClasses {
		ch, err := height(instance, known, seen)
		if err != nil {
			return 0, err
		}
		if ch > mh {
			mh = ch
		}
	}
	known[i] = mh + 1
	return mh + 1, nil
}

func appendUniqueClass(list []*ModuleClass, toAppend *ModuleClass) []*ModuleClass {
	for _, value := range list {
		if toAppend == value {
			return list
		}
	}
	return append(list, toAppend)
}

// populateClassDependencies consolidates the dependencies claimed by module classes
func (system *ModuleSystem) populateClassDependencies() error {
	for _, c := range system.classes {
		for _, d := range c.DependsOn {
			dc, ok := system.classes[d]
			if !ok {
				return fmt.Errorf("module class %q depends on %q which is not defined", c.Name, d)
			}
			c.dependentClasses = appendUniqueClass(c.dependentClasses, dc)
		}
		for _, d := range c.DependedBy {
			dc, ok := system.classes[d]
			if !ok {
				return fmt.Errorf("module class %q is depended by %q which is not defined", c.Name, d)
			}
			dc.dependentClasses = appendUniqueClass(dc.dependentClasses, c)
		}
	}
	return nil
}

// resolveClassOrder sorts the registered classes by dependency and sets
// classOrder field
func (system *ModuleSystem) resolveClassOrder() error {
	err := system.populateClassDependencies()
	if err != nil {
		return errors.Wrap(err, "error resolving module class order")
	}
	classes := make([]*ModuleClass, len(system.classes))
	i := 0
	for _, c := range system.classes {
		classes[i] = c
		i++
	}
	sorted, err := sortModuleClasses(classes)
	if err != nil {
		return errors.Wrap(err, "error resolving module class order")
	}
	system.classOrder = make([]string, len(system.classes))
	for i, c := range sorted {
		system.classOrder[i] = c.Name
	}
	return nil
}

// Given a module instance name like "app/foo/clients/bar", strip the "clients" part, i.e. usually the
// second to last path part.
func stripModuleClassName(classType, moduleDir string) string {
	parts := strings.Split(moduleDir, string(filepath.Separator))

	if len(parts) == 1 {
		return moduleDir
	}

	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == classType {

			// Shift parts left starting with ith element, overwriting parts[i]
			for j := i + 1; j < len(parts); j++ {
				parts[j-1] = parts[j]
			}

			parts = parts[0 : len(parts)-1]

			break
		}
	}
	return strings.Join(parts, string(filepath.Separator))
}

// ResolveModules resolves the module instances from the config on disk
// Using the system class and type definitions, the class directories are
// walked, and a module instance is initialized for each identified module in
// the target directory.
func (system *ModuleSystem) ResolveModules(
	packageRoot string,
	baseDirectory string,
	targetGenDir string,
) (map[string][]*ModuleInstance, error) {
	// resolve module class order before read and resolve each module
	if err := system.resolveClassOrder(); err != nil {
		return nil, err
	}

	resolvedModules := map[string]map[string]*ModuleInstance{}

	// system.classOrder is important. Downstream dependencies **must** be resolved first
	for _, className := range system.classOrder {

		defaultDependencies, err := system.getDefaultDependencies(
			baseDirectory,
			className,
			system.defaultDependencies,
			system.moduleSearchPaths,
		)
		if err != nil {
			return nil, errors.Wrapf(err, "error getting default dependencies for class %s", className)
		}

		for _, moduleDirectoryGlob := range system.moduleSearchPaths[className] {
			moduleDirectoriesAbs, err := filepath.Glob(filepath.Join(baseDirectory, moduleDirectoryGlob))
			if err != nil {
				return nil, errors.Wrapf(err, "error globbing %q", moduleDirectoryGlob)
			}

			runner := parallelize.NewUnboundedRunner(len(moduleDirectoriesAbs))
			for _, moduleDirAbs := range moduleDirectoriesAbs {
				f := system.getInstanceFunc(baseDirectory, className, packageRoot, targetGenDir, defaultDependencies,
					moduleDirectoryGlob)
				wrk := &parallelize.SingleParamWork{Data: moduleDirAbs, Func: f}
				runner.SubmitWork(wrk)
			}

			results, err := runner.GetResult()
			if err != nil {
				return nil, err
			}
			for _, instanceInf := range results {
				if instanceInf == nil {
					continue
				}
				instance := instanceInf.(*ModuleInstance)
				instanceMap := resolvedModules[instance.ClassName]
				if instanceMap == nil {
					instanceMap = map[string]*ModuleInstance{}
					resolvedModules[instance.ClassName] = instanceMap
				}
				instanceMap[instance.InstanceName] = instance
			}

			// Resolve dependencies for all classes
			resolveErr := system.populateResolvedDependencies(
				resolvedModules[className],
				resolvedModules,
			)
			if resolveErr != nil {
				return nil, resolveErr
			}

			// Resolved recursive dependencies for all classes
			recursiveErr := system.populateRecursiveDependencies(resolvedModules[className])
			if recursiveErr != nil {
				return nil, recursiveErr
			}
		}
	}

	classArrayModuleMap := map[string][]*ModuleInstance{}
	for className, instanceMap := range resolvedModules {
		for _, instance := range instanceMap {
			classArrayModuleMap[className] = append(classArrayModuleMap[className], instance)
		}
	}
	return classArrayModuleMap, nil
}

func (system *ModuleSystem) getInstanceFunc(baseDirectory string, className string, packageRoot string, targetGenDir string, defaultDependencies []ModuleDependency, moduleDirectoryGlob string) func(moduleDirAbsInf interface{}) (interface{}, error) {
	f := func(moduleDirAbsInf interface{}) (interface{}, error) {
		moduleDirAbs := moduleDirAbsInf.(string)
		stat, err := os.Stat(moduleDirAbs)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"internal error: cannot stat %q",
				moduleDirAbs,
			)
		}

		if !stat.IsDir() {
			// If a *-config.yaml file, or any other metadata file also matched the glob, skip it, since we are
			// interested only in the containing directories.
			return nil, nil
		}

		moduleDir, err := filepath.Rel(baseDirectory, moduleDirAbs)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"internal error: cannot make %q relative to %q",
				moduleDirAbs,
				baseDirectory,
			)
		}

		classConfigPath, _, _ := getConfigFilePath(moduleDirAbs, className)
		if classConfigPath == "" {
			fmt.Printf("  no class config found in %s directory\n", moduleDirAbs)
			// No class config found in this directory, skip over it
			return nil, nil
		}

		instance, instanceErr := system.readInstance(
			className,
			packageRoot,
			baseDirectory,
			targetGenDir,
			moduleDir,
			moduleDirAbs,
			defaultDependencies,
		)
		if instanceErr != nil {
			return nil, errors.Wrapf(
				instanceErr,
				"Error reading multi instance %q",
				moduleDir,
			)
		}

		if className != instance.ClassName {
			return nil, fmt.Errorf(
				"invariant: all instances in a multi-module directory %q are of type %q (violated by %q)",
				moduleDirectoryGlob,
				className,
				moduleDir,
			)
		}
		return instance, nil
	}
	return f
}

func (system *ModuleSystem) getDefaultDependencies(
	baseDirectory string,
	className string,
	dependencyGlobs map[string][]string,
	moduleSearchPaths map[string][]string,
) ([]ModuleDependency, error) {
	defaultDependencies := make([]ModuleDependency, 0)
	for _, defaultDepDirGlob := range dependencyGlobs[className] {
		defaultDepDirsAbs, err := filepath.Glob(filepath.Join(baseDirectory, defaultDepDirGlob))
		if err != nil {
			return nil, errors.Wrapf(err, "error globbing default dependency %q", defaultDepDirGlob)
		}
		ch := make(chan defaultDependencyRes, len(defaultDepDirsAbs))
		var wg sync.WaitGroup
		wg.Add(len(defaultDepDirsAbs))

		for _, defaultDepDirAbs := range defaultDepDirsAbs {
			go system.getDefaultDependency(baseDirectory, defaultDepDirAbs, moduleSearchPaths,
				className, ch, &wg)
		}
		go func() {
			wg.Wait()
			close(ch)
		}()
		for ele := range ch {
			if ele.err != nil {
				return nil, ele.err
			}
			if ele.dep == nil {
				continue
			}
			defaultDependencies = append(defaultDependencies, *ele.dep)
		}
	}
	return defaultDependencies, nil
}

type defaultDependencyRes struct {
	err error
	dep *ModuleDependency
}

func (system *ModuleSystem) getDefaultDependency(baseDirectory string, defaultDepDirAbs string,
	moduleSearchPaths map[string][]string, className string, ch chan defaultDependencyRes, wg *sync.WaitGroup) {
	defer wg.Done()
	dependencyClassName, err := getClassNameOfDependency(baseDirectory, defaultDepDirAbs, moduleSearchPaths)
	if err != nil {
		ch <- defaultDependencyRes{
			err: errors.Wrapf(
				err,
				"cannot get class name of dependency %s",
				defaultDepDirAbs,
			),
		}
		return
	}

	found := false
	for _, actualClassDependencyName := range system.classes[className].DependsOn {
		if dependencyClassName == actualClassDependencyName {
			found = true
			break
		}
	}
	if !found {
		ch <- defaultDependencyRes{err: errors.Errorf(
			"default dependency class %s is not a dependency of %s",
			dependencyClassName,
			className,
		),
		}
		return
	}

	classConfigPath, _, _ := getConfigFilePath(defaultDepDirAbs, dependencyClassName)
	if classConfigPath == "" {
		// No class config found, skip over this directory
		ch <- defaultDependencyRes{}
		return
	}

	dependencyDir, err := filepath.Rel(baseDirectory, defaultDepDirAbs)
	if err != nil {
		ch <- defaultDependencyRes{err: errors.Wrapf(
			err,
			"cannot make %q relative to %q for default dependency",
			defaultDepDirAbs,
			baseDirectory,
		),
		}
		return
	}

	defaultDependency, err := system.getModuleDependency(classConfigPath, dependencyDir, dependencyClassName)
	if err != nil {
		ch <- defaultDependencyRes{err: errors.Wrapf(
			err,
			"error getting default dependency module %s for class %s",
			defaultDepDirAbs,
			className,
		),
		}
		return
	}
	ch <- defaultDependencyRes{dep: defaultDependency}
	return
}

func (system *ModuleSystem) getModuleDependency(
	classConfigPath string,
	dependencyDir string,
	className string,
) (*ModuleDependency, error) {
	raw, readErr := ioutil.ReadFile(classConfigPath)
	if readErr != nil {
		return nil, errors.Wrapf(
			readErr,
			"error reading ClassConfig %q",
			classConfigPath,
		)
	}

	_, configErr := NewClassConfig(raw)
	if configErr != nil {
		return nil, errors.Wrapf(
			configErr,
			"error unmarshal ClassConfig %q",
			classConfigPath,
		)
	}

	var classPlural string
	for _, moduleClass := range system.classes {
		if moduleClass.Name == className {
			classPlural = moduleClass.NamePlural
		}
	}

	moduleDependency := ModuleDependency{
		ClassName:    className,
		InstanceName: stripModuleClassName(classPlural, dependencyDir),
	}

	return &moduleDependency, nil
}

func getClassNameOfDependency(
	baseDirectory string,
	dependencyDirectory string,
	moduleSearchPaths map[string][]string,
) (string, error) {
	for className, moduleGlobs := range moduleSearchPaths {
		for _, moduleGlob := range moduleGlobs {
			matched, err := filepath.Match(filepath.Join(baseDirectory, moduleGlob), dependencyDirectory)
			if err != nil {
				return "", errors.Wrapf(
					err,
					"error matching module glob %s with dependency directory %s",
					moduleGlob,
					dependencyDirectory,
				)
			}

			if matched {
				return className, nil
			}
		}
	}

	return "", errors.Errorf("could not find class for default dependency %s", dependencyDirectory)
}

func getConfigFilePath(dir, name string) (string, string, string) {
	yamlFileName := name + yamlConfigSuffix
	jsonFileName := ""
	path := filepath.Join(dir, yamlFileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Cannot find yaml file, try json file instead
		jsonFileName = name + jsonConfigSuffix
		yamlFileName = ""
		path = filepath.Join(dir, jsonFileName)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// Cannot find any config file
			path = ""
			jsonFileName = ""
			yamlFileName = ""
		}
	}
	return path, yamlFileName, jsonFileName
}

func (system *ModuleSystem) readInstance(
	className string,
	packageRoot string,
	baseDirectory string,
	targetGenDir string,
	instanceDirectory string,
	classConfigDir string,
	defaultDependencies []ModuleDependency,
) (*ModuleInstance, error) {
	classConfigPath, yamlFileName, jsonFileName := getConfigFilePath(classConfigDir, className)

	raw, readErr := ioutil.ReadFile(classConfigPath)
	if readErr != nil {
		return nil, errors.Wrapf(
			readErr,
			"Error reading ClassConfig %q",
			classConfigPath,
		)
	}

	jsonRaw := []byte{}
	if jsonFileName != "" {
		jsonRaw = raw
	}

	config, configErr := NewClassConfig(raw)
	if configErr != nil {
		return nil, errors.Wrapf(
			configErr,
			"Error unmarshal ClassConfig %q",
			classConfigPath,
		)
	}

	dependencies := readDeps(config.Dependencies)
	dependencies = append(dependencies, defaultDependencies...)

	var classPlural string
	for _, moduleClass := range system.classes {
		if moduleClass.Name == className {
			classPlural = moduleClass.NamePlural
		}
	}
	instanceName := stripModuleClassName(classPlural, instanceDirectory)

	packageInfo, err := readPackageInfo(
		packageRoot,
		baseDirectory,
		targetGenDir,
		className,
		instanceDirectory,
		config,
	)

	if err != nil {
		// TODO: We should accumulate errors and list them all here
		// Expected $class-config.yaml to exist in ...
		return nil, errors.Wrapf(
			err,
			"Error reading class package info for %q %q",
			className,
			instanceName,
		)
	}

	return &ModuleInstance{
		PackageInfo:           packageInfo,
		ClassName:             className,
		ClassType:             config.Type,
		BaseDirectory:         baseDirectory,
		Directory:             instanceDirectory,
		InstanceName:          instanceName,
		Dependencies:          dependencies,
		ResolvedDependencies:  map[string][]*ModuleInstance{},
		RecursiveDependencies: map[string][]*ModuleInstance{},
		DependencyOrder:       []string{},
		JSONFileName:          jsonFileName,
		YAMLFileName:          yamlFileName,
		JSONFileRaw:           jsonRaw,
		YAMLFileRaw:           raw,
		Config:                config.Config,
		SelectiveBuilding:     config.SelectiveBuilding,
	}, nil
}

func readDeps(yamlDeps map[string][]string) []ModuleDependency {
	depCount := 0

	for _, depsList := range yamlDeps {
		depCount += len(depsList)
	}

	deps := make([]ModuleDependency, depCount)
	depIndex := 0

	for className, depsList := range yamlDeps {
		for _, instanceName := range depsList {
			deps[depIndex] = ModuleDependency{
				ClassName:    className,
				InstanceName: instanceName,
			}
			depIndex++
		}
	}

	return deps
}

func readPackageInfo(
	packageRoot string,
	baseDirectory string,
	targetGenDir string,
	className string,
	instanceDirectory string,
	config *ClassConfig,
) (*PackageInfo, error) {
	qualifiedClassName := strings.Title(CamelCase(className))
	qualifiedInstanceName := strings.Title(CamelCase(config.Name))
	defaultAlias := packageName(qualifiedInstanceName + qualifiedClassName)

	relativeGeneratedPath, err := filepath.Rel(baseDirectory, targetGenDir)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Error computing generated import string for %q",
			targetGenDir,
		)
	}

	// The module system presently has special reservations for the "custom"
	// type. We should really extrapolate from the class type info what the
	// default export type is for this instance
	var isExportGenerated bool
	if config.Type == "custom" {
		if config.IsExportGenerated == nil {
			isExportGenerated = false
		} else {
			isExportGenerated = *config.IsExportGenerated
		}
	} else if config.IsExportGenerated == nil {
		isExportGenerated = true
	} else {
		isExportGenerated = *config.IsExportGenerated
	}

	return &PackageInfo{
		// The package name is assumed to be the lower case of the instance
		// Name plus the titular class name, such as fooClient
		PackageName: defaultAlias,
		// The prefixes "Static" and "Generated" are used to ensure global
		// uniqueness of the provided package aliases. Note that the default
		// package is "PackageName".
		PackageAlias:          defaultAlias + "static",
		PackageRoot:           packageRoot,
		GeneratedPackageAlias: defaultAlias + "generated",
		ModulePackageAlias:    defaultAlias + "module",
		PackagePath: path.Join(
			packageRoot,
			instanceDirectory,
		),
		GeneratedPackagePath: filepath.Join(
			packageRoot,
			relativeGeneratedPath,
			instanceDirectory,
		),
		ModulePackagePath: filepath.Join(
			packageRoot,
			relativeGeneratedPath,
			instanceDirectory,
			"module",
		),
		ExportName:            "New" + qualifiedClassName,
		InitializerName:       "Initialize" + qualifiedClassName,
		QualifiedInstanceName: qualifiedInstanceName,
		ExportType:            qualifiedClassName,
		IsExportGenerated:     isExportGenerated,
	}, nil
}

func (system *ModuleSystem) populateSpec(instance *ModuleInstance) error {
	classGenerators := system.classes[instance.ClassName]
	generator := classGenerators.types[instance.ClassType]

	if generator == nil {
		return nil
	}

	specProvider, ok := generator.(SpecProvider)
	if !ok {
		fmt.Fprintf(
			os.Stderr,
			"warning: %q %q generator is not a spec provider, cannot use incremental builds\n",
			instance.ClassName,
			instance.ClassType,
		)

		return fmt.Errorf("%q %q generator does not implement SpecProvider interface", instance.ClassName,
			instance.ClassType)
	}

	spec, err := specProvider.ComputeSpec(instance)
	if err != nil {
		return fmt.Errorf("error when running computespec: %s", err.Error())
	}
	instance.genSpec = spec
	filterNilDeps(instance)
	return nil
}

func filterNilDeps(in *ModuleInstance) {
	filterNilDepsHelper(in.ResolvedDependencies)
	filterNilDepsHelper(in.RecursiveDependencies)
}

func filterNilDepsHelper(deps map[string][]*ModuleInstance) {
	for cls, modInstances := range deps {
		var validClients []*ModuleInstance
		for _, c := range modInstances {
			if c.GeneratedSpec() == nil {
				continue
			}
			validClients = append(validClients, c)
		}
		deps[cls] = validClients
	}
}

// collectTransitiveDependencies will collect every instance in resolvedModules that depends on something in initialInstances.
func (system *ModuleSystem) collectTransitiveDependencies(
	initialInstances []ModuleDependency,
	allModules map[string][]*ModuleInstance,
) (map[string][]*ModuleInstance, error) {

	toBeBuiltModules := make(map[ModuleDependency]*ModuleInstance, 0)

	for _, className := range system.classOrder {

		// Convert every ModuleDependency to its corresponding *ModuleInstance
		for _, initialInstance := range initialInstances {
			for _, instance := range allModules[className] {
				if initialInstance.equal(instance) {
					toBeBuiltModules[instance.AsModuleDependency()] = instance
					break
				}
			}
		}

		// Collect all the ModuleInstances that depend on anything from toBeBuiltModules
		for _, dependentInstance := range toBeBuiltModules {
			for _, instance := range allModules[className] {
				classInstanceTransitives := instance.RecursiveDependencies[dependentInstance.ClassName]
				for _, classInstanceDependency := range classInstanceTransitives {
					if !classInstanceDependency.equal(dependentInstance) {
						continue
					}
					toBeBuiltModules[instance.AsModuleDependency()] = instance
					fmt.Printf(
						"Need to generate %q %q %q because it transitively depends on %q %q %q\n",
						instance.InstanceName,
						instance.ClassName,
						instance.ClassType,
						dependentInstance.InstanceName,
						dependentInstance.ClassName,
						dependentInstance.ClassType,
					)
					break
				}
			}
		}
	}

	toBeBuiltModulesList := make(map[string][]*ModuleInstance)
	for _, instance := range toBeBuiltModules {
		if _, ok := toBeBuiltModulesList[instance.ClassName]; !ok {
			toBeBuiltModulesList[instance.ClassName] = make([]*ModuleInstance, 0)
		}
		toBeBuiltModulesList[instance.ClassName] = append(toBeBuiltModulesList[instance.ClassName], instance)
	}

	system.trimToSelectiveDependencies(toBeBuiltModulesList)

	return toBeBuiltModulesList, nil
}

func instanceFQN(instance *ModuleInstance) string {
	return fmt.Sprintf("%s.%s", instance.ClassName, instance.InstanceName)
}

// IncrementalBuild is like Build but filtered to only the given module instances.
func (system *ModuleSystem) IncrementalBuild(
	packageRoot string,
	baseDirectory string,
	targetGenDir string,
	instances []ModuleDependency,
	resolvedModules map[string][]*ModuleInstance,
	commitChange bool,
) (map[string][]*ModuleInstance, error) {

	errModuleMap := map[*ModuleInstance]struct{}{}
	toBeBuiltModules := make(map[string][]*ModuleInstance)

	for _, className := range system.classOrder {
		var wg sync.WaitGroup
		wg.Add(len(resolvedModules[className]))
		ch := make(chan *ModuleInstance, len(resolvedModules[className]))
		for _, instance := range resolvedModules[className] {
			go func(instance *ModuleInstance) {
				defer wg.Done()
				if err := system.populateSpec(instance); err != nil {
					ch <- instance
				}
			}(instance)
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		for mi := range ch {
			errModuleMap[mi] = struct{}{}
		}
	}

	resolvedModulesCopy := map[string][]*ModuleInstance{}
	for cls, modules := range resolvedModules {
		for _, m := range modules {
			if _, ok := errModuleMap[m]; ok {
				// skipping error modules
				fmt.Println("skipping module gen", m.InstanceName)
				continue
			}
			resolvedModulesCopy[cls] = append(resolvedModulesCopy[cls], m)
		}
	}
	resolvedModules = resolvedModulesCopy

	if instances == nil || len(instances) == 0 {
		for _, modules := range resolvedModules {
			for _, instance := range modules {
				instances = append(instances, instance.AsModuleDependency())
			}
		}
	}

	// If toBeBuiltModules is not empty already, it is likely that one of the modules does not implement
	// the SpecProvider interface, hence incremental build is not possible.
	if len(toBeBuiltModules) == 0 {
		var err error
		toBeBuiltModules, err = system.collectTransitiveDependencies(instances,
			resolvedModules)
		if err != nil {
			// if incrementalBuild fails, perform a full build.
			fmt.Printf("Falling back to full build due to err: %s\n", err.Error())
			toBeBuiltModules = resolvedModules
		}
	}

	moduleCount := 0
	var moduleIndex int32
	for _, moduleList := range toBeBuiltModules {
		moduleCount += len(moduleList)
	}

	for _, class := range system.classOrder {
		var wg sync.WaitGroup
		wg.Add(len(toBeBuiltModules[class]))
		ch := make(chan error, len(toBeBuiltModules[class]))
		for _, moduleInstance := range toBeBuiltModules[class] {
			go func(moduleInstance *ModuleInstance) {
				defer wg.Done()
				physicalGenDir := filepath.Join(targetGenDir, moduleInstance.Directory)
				prettyDir, _ := filepath.Rel(baseDirectory, physicalGenDir)
				PrintGenLine(moduleInstance.ClassType, moduleInstance.ClassName, moduleInstance.InstanceName,
					prettyDir, int(atomic.AddInt32(&moduleIndex, 1)), moduleCount)
				if err := system.Build(packageRoot, baseDirectory, physicalGenDir, moduleInstance, commitChange); err != nil {
					ch <- err
				}
			}(moduleInstance)
		}

		go func() {
			wg.Wait()
			close(ch)
		}()

		for err := range ch {
			if err != nil {
				return nil, err
			}
		}
	}

	var wg sync.WaitGroup
	ch := make(chan error, len(system.postGenHook))
	for i, hook := range system.postGenHook {
		if hook != nil {
			wg.Add(1)
			go func(hook PostGenHook) {
				defer wg.Done()
				if err := hook(toBeBuiltModules); err != nil {
					ch <- errors.Wrapf(err, "error running %dth post generation hook", i)
				}
			}(hook)
		}
	}

	go func() {
		wg.Wait()
		close(ch)
	}()
	for err := range ch {
		if err != nil {
			return toBeBuiltModules, err
		}
	}

	return toBeBuiltModules, nil
}

func (system *ModuleSystem) trimToSelectiveDependencies(toBeBuiltModules map[string][]*ModuleInstance) {
	if system.selectiveBuilding {
		selectiveModuleInstances := system.getSelectiveModules(toBeBuiltModules)
		toBuiltMap := flattenInstances(toBeBuiltModules)
		for _, instance := range selectiveModuleInstances {
			instance.ResolvedDependencies = resolvedSelectiveDependencies(instance, toBuiltMap)
			instance.RecursiveDependencies = recursiveSelectiveDependencies(instance)
		}
	}
}

func (system *ModuleSystem) getSelectiveModules(toBeBuiltModules map[string][]*ModuleInstance) map[string]*ModuleInstance {
	selectiveModuleInstances := map[string]*ModuleInstance{}
	for _, instances := range toBeBuiltModules {
		for _, instance := range instances {
			if instance.SelectiveBuilding {
				selectiveModuleInstances[instanceFQN(instance)] = instance
			}
		}
	}
	return selectiveModuleInstances
}

// recursiveSelectiveDependencies gets a recursive dependencies of resolved (direct)
// dependencies including direct dependencies itself
func recursiveSelectiveDependencies(instance *ModuleInstance) map[string][]*ModuleInstance {
	filteredRecursiveMap := map[string]*ModuleInstance{}
	for _, resolvedDependencies := range instance.ResolvedDependencies {
		for _, resolvedInstance := range resolvedDependencies {
			for _, recursiveDependencies := range resolvedInstance.RecursiveDependencies {
				for _, recursiveInstance := range recursiveDependencies {
					filteredRecursiveMap[instanceFQN(recursiveInstance)] = recursiveInstance
				}
			}
			filteredRecursiveMap[instanceFQN(resolvedInstance)] = resolvedInstance
		}
	}

	filteredRecursiveModules := map[string][]*ModuleInstance{}
	for _, instance := range filteredRecursiveMap {
		filteredRecursiveModules[instance.ClassName] = append(filteredRecursiveModules[instance.ClassName], instance)
	}
	return filteredRecursiveModules
}

// resolvedSelectiveDependencies gets a subset of resolved dependencies (direct) of a instance which needs to be built
func resolvedSelectiveDependencies(instance *ModuleInstance, toBuiltMap map[string]*ModuleInstance) map[string][]*ModuleInstance {
	filteredResolvedModules := map[string][]*ModuleInstance{}
	for _, resolvedDependencies := range instance.ResolvedDependencies {
		for _, resolvedInstance := range resolvedDependencies {
			if _, ok := toBuiltMap[instanceFQN(resolvedInstance)]; ok {
				filteredResolvedModules[resolvedInstance.ClassName] = append(
					filteredResolvedModules[resolvedInstance.ClassName], resolvedInstance)
			}
		}
	}
	// if after filtering there are no dependencies, we will default it to all dependencies.
	// this is done as when a module alone changes toBuilt would not have any dependencies
	if len(filteredResolvedModules) == 0 {
		filteredResolvedModules = instance.ResolvedDependencies
	}
	return filteredResolvedModules
}

func flattenInstances(resolvedModules map[string][]*ModuleInstance) map[string]*ModuleInstance {
	flattenedMap := map[string]*ModuleInstance{}
	for _, instances := range resolvedModules {
		for _, instance := range instances {
			flattenedMap[instanceFQN(instance)] = instance
		}
	}
	return flattenedMap
}

// Build invokes the generator for a module instance and optionally writes the files to disk
func (system *ModuleSystem) Build(packageRoot string, baseDirectory string, physicalGenDir string,
	instance *ModuleInstance, commitChange bool) error {
	classGenerators := system.classes[instance.ClassName]
	generator := classGenerators.types[instance.ClassType]

	if generator == nil {
		fmt.Fprintf(
			os.Stderr,
			"Skipping generation of %q %q class of type %q "+
				"as generator is not defined\n",
			instance.InstanceName,
			instance.ClassName,
			instance.ClassType,
		)
		return nil
	}

	buildResult, err := generator.Generate(instance)

	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"Error generating %q %q of type %q:\n%s\n",
			instance.InstanceName,
			instance.ClassName,
			instance.ClassType,
			err.Error(),
		)
		return err
	}

	if buildResult == nil {
		return nil
	}
	instance.genSpec = buildResult.Spec
	if !commitChange {
		return nil
	}
	runner := parallelize.NewUnboundedRunner(len(buildResult.Files))
	for filePath, content := range buildResult.Files {
		f := func(filePathInf interface{}, contentInf interface{}) (interface{}, error) {

			filePath := filePathInf.(string)
			content := contentInf.([]byte)

			filePath = filepath.Clean(filePath)

			resolvedPath := filepath.Join(
				physicalGenDir,
				filePath,
			)

			if err := writeFile(resolvedPath, content); err != nil {
				return nil, errors.Wrapf(
					err,
					"Error writing to file %q",
					resolvedPath,
				)
			}

			// HACK: The module system writer shouldn't
			// assume that we want to format the files in
			// this way, but we don't have these formatters
			// as a library or a custom post build script
			// for the generators yet.
			if filepath.Ext(filePath) == ".go" {
				if err := FormatGoFile(resolvedPath); err != nil {
					return nil, err
				}
			}
			return nil, nil
		}

		wrk := &parallelize.TwoParamWork{Data1: filePath, Data2: content, Func: f}
		runner.SubmitWork(wrk)
	}

	_, err = runner.GetResult()
	if err != nil {
		return err
	}

	return nil
}

// GenerateBuild will, given a module system configuration directory and a
// target build directory, run the generators assigned to each type of module
// and write the generated output to the module build directory if commitChange
//
// Deprecated: Use Build instead.
func (system *ModuleSystem) GenerateBuild(
	packageRoot string,
	baseDirectory string,
	targetGenDir string,
	commitChange bool,
) (map[string][]*ModuleInstance, error) {
	resolvedModules, err := system.ResolveModules(
		packageRoot,
		baseDirectory,
		targetGenDir,
	)

	if err != nil {
		return nil, err
	}

	moduleCount := 0
	moduleIndex := 0
	for _, moduleList := range resolvedModules {
		moduleCount += len(moduleList)
	}

	for _, class := range system.classOrder {
		for _, moduleInstance := range resolvedModules[class] {
			moduleIndex++

			physicalGenDir := filepath.Join(targetGenDir, moduleInstance.Directory)
			prettyDir, _ := filepath.Rel(baseDirectory, physicalGenDir)
			PrintGenLine(moduleInstance.ClassType, moduleInstance.ClassName, moduleInstance.InstanceName, prettyDir, moduleIndex, moduleCount)
			err := system.Build(packageRoot, baseDirectory, physicalGenDir, moduleInstance, commitChange)
			if err != nil {
				return nil, err
			}
		}
	}

	for i, hook := range system.postGenHook {
		if err := hook(resolvedModules); err != nil {
			return resolvedModules, errors.Wrapf(err, "error running %dth post generation hook", i)
		}
	}

	return resolvedModules, nil
}

// PrintGenLine prints the module generation process to stdout
func PrintGenLine(
	classType, className, instanceName, buildPath string,
	idx, count int,
) {
	fmt.Printf(
		"Generating %12s %12s %-30s in %-50s %d/%d\n",
		classType,
		className,
		instanceName,
		buildPath,
		idx,
		count,
	)
}

// FormatGoFile reformat the go file imports
func FormatGoFile(filePath string) error {
	goimportsCmd := exec.Command("goimports", "-w", "-e", filePath)
	goimportsCmd.Stdout = os.Stdout
	goimportsCmd.Stderr = os.Stderr

	if err := goimportsCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to goimports file: %q", filePath)
	}
	return nil
}

// ModuleClass defines a module class in the build configuration directory.
// THis could be something like an Endpoint class which contains multiple
// endpoint configurations, or a Lib class, that is itself a module instance
type ModuleClass struct {
	Name       string
	NamePlural string
	ClassType  moduleClassType

	DependsOn  []string
	DependedBy []string
	types      map[string]BuildGenerator

	// private field which is populated before module resolving
	dependentClasses []*ModuleClass
}

// BuildResult is the result of running a module generator
type BuildResult struct {
	// Files contains a map of file names to file bytes to be written to the
	// module build directory
	Files map[string][]byte
	// Spec is an arbitrary type that can be used to share computed data
	// between dependencies
	Spec interface{}
}

// BuildGenerator provides a function to generate a module instance build
// artifact from its configuration as part of a build step. For example, an
// Endpoint module instance may generate endpoint handler code
type BuildGenerator interface {
	Generate(
		instance *ModuleInstance,
	) (*BuildResult, error)
}

// SpecProvider is a generator that can provide a specification without
// running the build step.
type SpecProvider interface {
	ComputeSpec(
		instance *ModuleInstance,
	) (interface{}, error)
}

// PackageInfo provides information about the package associated with a module
// instance.
type PackageInfo struct {
	// PackageName is the name of the generated package, and should be the same
	// as the package name used by any custom code in the config directory
	PackageName string
	// PackageAlias is the unique import alias for non-generated packages
	PackageAlias string
	// PackageRoot is the unique import root for non-generated packages
	PackageRoot string
	// GeneratedPackageAlias is the unique import alias for generated packages
	GeneratedPackageAlias string
	// ModulePackageAlias is the unique import alias for the module system's,
	// generated subpackage
	ModulePackageAlias string
	// PackagePath is the full package path for the non-generated code
	PackagePath string
	// GeneratedPackagePath is the full package path for the generated code
	GeneratedPackagePath string
	// ModulePacakgePath is the full package path for the generated dependency
	// structs and initializers
	ModulePackagePath string
	// QualifiedInstanceName for this package. Pascal case name for this module.
	QualifiedInstanceName string
	// ExportName is the name on the module initializer function
	ExportName string
	// ExportType refers to the type returned by the module initializer
	ExportType string
	// InitializerName is the name of function that can fully initialize the
	// module and its dependencies
	InitializerName string
	// IsExportGenerated is true if the export type is provided by the
	// generated package, otherwise it is assumed that the export type resides
	// in the non-generated package
	IsExportGenerated bool
}

// ImportPackagePath returns the correct package path for the module's exported
// type, depending on which package (generated or not) the type lives in
func (info *PackageInfo) ImportPackagePath() string {
	if info.IsExportGenerated {
		return info.GeneratedPackagePath
	}

	return info.PackagePath
}

// ImportPackageAlias returns the correct package alias for referencing the
// module's exported type, depending on whether or not the export is generated
func (info *PackageInfo) ImportPackageAlias() string {
	if info.IsExportGenerated {
		return info.GeneratedPackageAlias
	}

	return info.PackageAlias
}

// ModuleInstance is a configured module inside a module class directory.
// For example, this could be
//     ClassName:    "Endpoint,
//     ClassType:    "http",
//     BaseDirectory "/path/to/service/base/"
//     Directory:    "clients/health/"
//     InstanceName: "health",
type ModuleInstance struct {
	// genSpec is used to share generated specs across dependencies. Generators
	// should not mutate this directly, and should return the spec as a result.
	// Only the module system code should mutate a module instance.
	genSpec interface{}
	// PackageInfo is the name for the generated module instance
	PackageInfo *PackageInfo
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
	// YAML/JSON file
	InstanceName string
	// Config is a reference to the instance "config" key in the instances YAML
	// file.
	Config map[string]interface{}
	// Dependencies is a list of dependent modules as defined in the instances
	// YAML file
	Dependencies []ModuleDependency
	// Resolved dependencies is a list of direct dependent modules after processing
	// (fully resolved)
	ResolvedDependencies map[string][]*ModuleInstance
	// Recursive dependencies is a list of dependent modules and all of their
	// dependencies, i.e. the full tree of dependencies for this module. Each
	// class list is sorted for initialization order
	RecursiveDependencies map[string][]*ModuleInstance
	// DependencyOrder is the bottom to top order in which the recursively
	// resolved dependency class names can depend on each other
	DependencyOrder []string
	// The JSONFileName is file name of the instance JSON file
	JSONFileName string // Deprecated
	// The YAMLFileName is file name of the instance YAML file
	YAMLFileName string
	// JSONFileRaw is the raw JSON file read as bytes used for future parsing
	JSONFileRaw []byte // Deprecated
	// YAMLFileRaw is the raw YAML file read as bytes used for future parsing
	YAMLFileRaw []byte
	// SelectiveBuilding allows the module to be built with subset of dependencies
	SelectiveBuilding bool
}

func (instance *ModuleInstance) String() string {
	return fmt.Sprintf("[instance %q %q]", instance.ClassType, instance.InstanceName)
}

// GeneratedSpec returns the last spec result returned for the module instance
func (instance *ModuleInstance) GeneratedSpec() interface{} {
	return instance.genSpec
}

// Equal checks equality of two ModuleInstances
func (instance *ModuleInstance) equal(other *ModuleInstance) bool {
	return instance.InstanceName == other.InstanceName && instance.ClassName == other.ClassName
}

// AsModuleDependency creates an equivalent ModuleDependency object for this instance
func (instance *ModuleInstance) AsModuleDependency() ModuleDependency {
	return ModuleDependency{
		InstanceName: instance.InstanceName,
		ClassName:    instance.ClassName,
	}
}

// ModuleDependency defines a module instance required by another instance
type ModuleDependency struct {
	// ClassName is the name of the class as defined in the module system
	ClassName string
	// InstanceName is the name of the dependency instance as configu
	InstanceName string
}

func (m ModuleDependency) equal(other *ModuleInstance) bool {
	return m.InstanceName == other.InstanceName && m.ClassName == other.ClassName
}

// ClassConfigBase defines the shared data fields for all class configs.
// Configs of different classes can derive from this base struct by add a
// specific "config" field and "dependencies" field
type ClassConfigBase struct {
	// Name is the class instance name used to identify the module as a
	// dependency. The combination of the class Name and this instance name
	// is unique.
	Name string `yaml:"name" json:"name"`
	// Type refers to the class type used to generate the dependency
	Type string `yaml:"type" json:"type"`
	// IsExportGenerated determines whether or not the export lives in
	// IsExportGenerated defaults to true if not set.
	IsExportGenerated *bool `yaml:"IsExportGenerated,omitempty" json:"IsExportGenerated"`
	// Owner is the Name of the class instance owner
	Owner string `yaml:"owner,omitempty"`
	// SelectiveBuilding allows the module to be built with subset of dependencies
	SelectiveBuilding bool `yaml:"selectiveBuilding,omitempty" json:"selectiveBuilding"`
}

// ClassConfig maps onto a YAML configuration for a class type
type ClassConfig struct {
	ClassConfigBase `yaml:",inline" json:",inline"`
	// Dependencies is a map of class name to a list of instance names. This
	// infers the dependencies struct generated for the initializer
	Dependencies map[string][]string `yaml:"dependencies" json:"dependencies"`
	// The configuration map for this class instance. This depends on the
	// class name and class type, and is interpreted by each module generator.
	Config map[string]interface{} `yaml:"config" json:"config"`
}

// NewClassConfig unmarshals raw bytes into a ClassConfig struct
// or return an error if it cannot be unmarshaled into the struct
func NewClassConfig(raw []byte) (*ClassConfig, error) {
	config := &ClassConfig{}
	parseErr := yaml.Unmarshal(raw, config)

	if parseErr != nil {
		return nil, errors.Wrapf(
			parseErr,
			"Error yaml parsing config",
		)
	}

	if config.Name == "" {
		return nil, errors.Errorf(
			"Error reading instance name",
		)
	}

	if config.Type == "" {
		return nil, errors.Errorf(
			"Error reading instance type",
		)
	}

	if config.Dependencies == nil {
		config.Dependencies = map[string][]string{}
	}

	return config, nil
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
