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

package codegen

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
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

const configSuffix = "-config.json"

// NewModuleSystem returns a new module system
func NewModuleSystem() *ModuleSystem {
	return &ModuleSystem{
		classes:    map[string]*ModuleClass{},
		classOrder: []string{},
	}
}

// ModuleSystem defines the module classes and their type generators
type ModuleSystem struct {
	classes    map[string]*ModuleClass
	classOrder []string
}

// RegisterClass defines a class of module in the module system
// For example, an "Endpoint" class or a "Client" class
func (system *ModuleSystem) RegisterClass(class ModuleClass) error {
	name := class.Name
	if name == "" {
		return errors.Errorf("A module class name must not be empty")
	}

	if system.classes[name] != nil {
		return errors.Errorf("Module class %q is already defined", name)
	}

	for i, dir := range class.Directories {
		class.Directories[i] = filepath.Clean(dir)
	}
	class.Directories = dedup(class.Directories)

	for _, dir := range class.Directories {
		if err := system.validateClassDir(&class, dir); err != nil {
			return err
		}
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

// RegisterClassDir adds the given dir to the directories of the module class with given className.
// This method allows projects built on zanzibar to have arbitrary directories to host module class
// configs, therefore is mainly intended for external use.
func (system *ModuleSystem) RegisterClassDir(className string, dir string) error {
	dir = filepath.Clean(dir)
	for _, class := range system.classes {
		if className != class.Name {
			continue
		}

		if err := system.validateClassDir(class, dir); err != nil {
			return err
		}

		for _, registered := range class.Directories {
			if dir == registered {
				return nil
			}
		}
		class.Directories = append(class.Directories, dir)
		return nil
	}
	return errors.Errorf("Module class %q is not found", className)
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
	classInstances []*ModuleInstance,
	resolvedModules map[string][]*ModuleInstance,
) error {
	// Resolve the class dependencies
	for _, classInstance := range classInstances {
		dependencyClassNames := map[string]bool{}

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

			// TODO: We don't want to linear scan here
			var dependencyInstance *ModuleInstance

			for _, instance := range moduleClassInstances {
				if instance.InstanceName == classDependency.InstanceName {
					dependencyInstance = instance
					break
				}
			}

			if dependencyInstance == nil {
				return errors.Errorf(
					"Unknown %q class depdendency %q"+
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

			dependencyClassNames[classDependency.ClassName] = true
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
	instances []*ModuleInstance,
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

			classInstance.DependencyOrder = append(
				classInstance.DependencyOrder,
				className,
			)

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

	resolvedModules := map[string][]*ModuleInstance{}

	for _, className := range system.classOrder {
		class := system.classes[className]
		classInstances := []*ModuleInstance{}

		for _, dir := range class.Directories {

			fullInstanceDirectory := filepath.Join(baseDirectory, dir)

			if class.ClassType == SingleModule {
				instance, instanceErr := system.readInstance(
					packageRoot,
					baseDirectory,
					targetGenDir,
					className,
					dir,
				)
				if instanceErr != nil {
					return nil, errors.Wrapf(
						instanceErr,
						"Error reading single instance %q in %q",
						className,
						dir,
					)
				}
				classInstances = append(classInstances, instance)
			} else {
				instances, err := system.resolveMultiModules(
					packageRoot,
					baseDirectory,
					targetGenDir,
					fullInstanceDirectory,
					className,
					class,
				)

				if err != nil {
					return nil, errors.Wrapf(err,
						"Error reading resolving multi modules of %q",
						className,
					)
				}

				classInstances = append(classInstances, instances...)
			}
		}

		resolvedModules[className] = classInstances

		// Resolve dependencies for all classes
		resolveErr := system.populateResolvedDependencies(
			classInstances,
			resolvedModules,
		)
		if resolveErr != nil {
			return nil, resolveErr
		}

		// Resolved recursive dependencies for all classes
		recursiveErr := system.populateRecursiveDependencies(classInstances)
		if recursiveErr != nil {
			return nil, recursiveErr
		}
	}

	return resolvedModules, nil
}

func (system *ModuleSystem) resolveMultiModules(
	packageRoot string,
	baseDirectory string,
	targetGenDir string,
	classDir string, // full path
	className string,
	class *ModuleClass,
) ([]*ModuleInstance, error) {
	relClassDir, err := filepath.Rel(baseDirectory, classDir)
	if err != nil {
		return nil, errors.Wrapf(err,
			"Error relative class directory for %q",
			className,
		)
	}

	configFile := filepath.Join(classDir, className+configSuffix)
	if _, err := os.Stat(configFile); err == nil {
		instance, instanceErr := system.readInstance(
			packageRoot,
			baseDirectory,
			targetGenDir,
			className,
			relClassDir,
		)
		if instanceErr != nil {
			return nil, errors.Wrapf(
				instanceErr,
				"Error reading multi instance %q in %q",
				className,
				relClassDir,
			)
		}
		return []*ModuleInstance{instance}, nil
	}

	classInstances := []*ModuleInstance{}
	files, err := ioutil.ReadDir(classDir)

	if err != nil {
		// TODO: We should accumulate errors and list them all here
		// Expected $path to be a class directory
		return nil, errors.Wrapf(
			err,
			"Error reading module instance directory %q",
			classDir,
		)
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		instances, err := system.resolveMultiModules(
			packageRoot,
			baseDirectory,
			targetGenDir,
			filepath.Join(classDir, file.Name()),
			className,
			class,
		)
		if err != nil {
			return nil, errors.Wrapf(err,
				"Error reading subdir of multi instance %q in %q",
				className,
				filepath.Join(relClassDir, file.Name()),
			)
		}
		classInstances = append(classInstances, instances...)
	}

	return classInstances, nil
}

func (system *ModuleSystem) readInstance(
	packageRoot string,
	baseDirectory string,
	targetGenDir string,
	className string,
	instanceDirectory string,
) (*ModuleInstance, error) {

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
			"Error reading JSON Config %q",
			classConfigPath,
		)
	}

	dependencies := readDeps(jsonConfig.Dependencies)
	packageInfo, err := readPackageInfo(
		packageRoot,
		baseDirectory,
		targetGenDir,
		className,
		instanceDirectory,
		&jsonConfig,
		dependencies,
	)

	if err != nil {
		// TODO: We should accumulate errors and list them all here
		// Expected $class-config.json to exist in ...
		return nil, errors.Wrapf(
			err,
			"Error reading class package info for %q %q",
			className,
			jsonConfig.Name,
		)
	}

	return &ModuleInstance{
		PackageInfo:           packageInfo,
		ClassName:             className,
		ClassType:             jsonConfig.Type,
		BaseDirectory:         baseDirectory,
		Directory:             instanceDirectory,
		InstanceName:          jsonConfig.Name,
		Dependencies:          dependencies,
		ResolvedDependencies:  map[string][]*ModuleInstance{},
		RecursiveDependencies: map[string][]*ModuleInstance{},
		DependencyOrder:       []string{},
		JSONFileName:          jsonFileName,
		JSONFileRaw:           raw,
	}, nil
}

func readDeps(jsonDeps map[string][]string) []ModuleDependency {
	depCount := 0

	for _, depsList := range jsonDeps {
		depCount += len(depsList)
	}

	deps := make([]ModuleDependency, depCount)
	depIndex := 0

	for className, depsList := range jsonDeps {
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
	jsonConfig *JSONClassConfig,
	dependencies []ModuleDependency,
) (*PackageInfo, error) {
	qualifiedClassName := strings.Title(camelCase(className))
	qualifiedInstanceName := strings.Title(camelCase(jsonConfig.Name))
	defaultAlias := camelCase(strings.ToLower(qualifiedInstanceName)) +
		qualifiedClassName

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
	if jsonConfig.Type == "custom" {
		if jsonConfig.IsExportGenerated == nil {
			isExportGenerated = false
		} else {
			isExportGenerated = *jsonConfig.IsExportGenerated
		}
	} else if jsonConfig.IsExportGenerated == nil {
		isExportGenerated = true
	} else {
		isExportGenerated = *jsonConfig.IsExportGenerated
	}

	return &PackageInfo{
		// The package name is assumed to be the lower case of the instance
		// Name plus the titular class name, such as fooClient
		PackageName: defaultAlias,
		// The prefixes "Static" and "Generated" are used to ensure global
		// uniqueness of the provided package aliases. Note that the default
		// package is "PackageName".
		PackageAlias:          defaultAlias + "Static",
		GeneratedPackageAlias: defaultAlias + "Generated",
		ModulePackageAlias:    defaultAlias + "Module",
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

// GenerateBuild will, given a module system configuration directory and a
// target build directory, run the generators assigned to each type of module
// and write the generated output to the module build directory if commitChange
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
	for _, moduleList := range resolvedModules {
		moduleCount += len(moduleList)
	}

	moduleIndex := 0
	for _, className := range system.classOrder {
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
				"Generating %8s %8s %-20s in %-30s %d/%d\n",
				classInstance.ClassType,
				classInstance.ClassName,
				classInstance.InstanceName,
				prettyBuildPath,
				moduleIndex,
				moduleCount,
			)

			classGenerators := system.classes[classInstance.ClassName]
			generator := classGenerators.types[classInstance.ClassType]

			if generator == nil {
				fmt.Printf(
					"Skipping generation of %q %q class of type %q "+
						"as generator is not defined\n",
					classInstance.InstanceName,
					classInstance.ClassName,
					classInstance.ClassType,
				)
				continue
			}

			buildResult, err := generator.Generate(classInstance)

			if err != nil {
				fmt.Printf(
					"Error generating %q %q of type %q:\n%s\n",
					classInstance.InstanceName,
					classInstance.ClassName,
					classInstance.ClassType,
					err.Error(),
				)
				return nil, err
			}

			if buildResult == nil {
				continue
			}
			classInstance.genSpec = buildResult.Spec
			if !commitChange {
				continue
			}
			for filePath, content := range buildResult.Files {
				filePath = filepath.Clean(filePath)

				resolvedPath := filepath.Join(
					buildPath,
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
			}
		}
	}

	return resolvedModules, nil
}

// FormatGoFile reformat the go file imports
func FormatGoFile(filePath string) error {
	gofmtCmd := exec.Command("gofmt", "-s", "-w", "-e", filePath)
	gofmtCmd.Stdout = os.Stdout
	gofmtCmd.Stderr = os.Stderr

	if err := gofmtCmd.Run(); err != nil {
		return errors.Wrapf(err, "failed to gofmt file: %q", filePath)
	}

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
	Name        string
	ClassType   moduleClassType
	Directories []string
	DependsOn   []string
	DependedBy  []string
	types       map[string]BuildGenerator

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

// PackageInfo provides information about the package associated with a module
// instance.
type PackageInfo struct {
	// PackageName is the name of the generated package, and should be the same
	// as the package name used by any custom code in the config directory
	PackageName string
	// PackageAlias is the unique import alias for non-generated packages
	PackageAlias string
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
	// json file
	InstanceName string
	// Config is a reference to the instance "config" key in the instances json
	//file
	Config interface{}
	// Dependencies is a list of dependent modules as defined in the instances
	// json file
	Dependencies []ModuleDependency
	// Resolved dependencies is a list of dependent modules after processing
	// (fully resolved)
	ResolvedDependencies map[string][]*ModuleInstance
	// Recursive dependencies is a list of dependent modules and all of their
	// dependencies, i.e. the full tree of dependencies for this module. Each
	// class list is sorted for initialization order
	RecursiveDependencies map[string][]*ModuleInstance
	// DependencyOrder is the bottom to top order in which the recursively
	// resolved dependency class names can depend on each other
	DependencyOrder []string
	// The JSONFileName is file name of the instance json file
	JSONFileName string
	// JSONFileRaw is the raw JSON file read as bytes used for future parsing
	JSONFileRaw []byte
}

// GeneratedSpec returns the last spec result returned for the module instance
func (instance *ModuleInstance) GeneratedSpec() interface{} {
	return instance.genSpec
}

// ModuleDependency defines a module instance required by another instance
type ModuleDependency struct {
	// ClassName is the name of the class as defined in the module system
	ClassName string
	// InstanceName is the name of the dependency instance as configu
	InstanceName string
}

// JSONClassConfig maps onto a json configuration for a class type
type JSONClassConfig struct {
	// Name is the class instance name used to identify the module as a
	// dependency. The combination of the class Name and this instance name
	// is unique.
	Name string `json:"name"`
	// The configuration object for this class instance. This depends on the
	// class name and class type, and is interpreted by each module generator.
	Config interface{} `json:"config"`
	// Dependencies is a map of class name to a list of instance names. This
	// infers the dependencies struct generated for the initializer
	Dependencies map[string][]string `json:"dependencies"`
	// Type refers to the class type used to generate the dependency
	Type string `json:"type"`
	// IsExportGenerated determines whether or not the export lives in
	// IsExportGenerated defaults to true if not set.
	IsExportGenerated *bool `json:"IsExportGenerated"`
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
