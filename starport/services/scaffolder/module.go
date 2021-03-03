package scaffolder

import (
	"context"
	"errors"
	"fmt"

	module_create "github.com/zhigui-projects/zeus-onestop/starport/templates/module/create"
	module_import "github.com/zhigui-projects/zeus-onestop/starport/templates/module/import"

	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobuffalo/genny"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/cmdrunner"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/cmdrunner/step"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/cosmosver"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/gomodulepath"
)

const (
	wasmImport                 = "github.com/CosmWasm/wasmd"
	apppkg                     = "app"
	moduleDir                  = "x"
	wasmVersionCommitLaunchpad = "b30902fe1fbe5237763775950f729b90bf34d53f"
	wasmVersionCommitStargate  = "f9015cba4793d03cf7a77d7253375b16ad3d3eef"
)

// CreateModule creates a new empty module in the scaffolded app
func (s *Scaffolder) CreateModule(moduleName string) error {
	version, err := s.version()
	if err != nil {
		return err
	}
	majorVersion := version.Major()
	// Check if the module already exist
	ok, err := ModuleExists(s.path, moduleName)
	if err != nil {
		return err
	}
	if ok {
		return fmt.Errorf("the module %v already exists", moduleName)
	}
	path, err := gomodulepath.ParseAt(s.path)
	if err != nil {
		return err
	}

	var (
		g    *genny.Generator
		opts = &module_create.CreateOptions{
			ModuleName: moduleName,
			ModulePath: path.RawPath,
			AppName:    path.Package,
			OwnerName:  owner(path.RawPath),
		}
	)
	if majorVersion == cosmosver.Launchpad {
		g, err = module_create.NewCreateLaunchpad(opts)
	} else {
		g, err = module_create.NewCreateStargate(opts)
	}

	if err != nil {
		return err
	}
	run := genny.WetRunner(context.Background())
	run.With(g)
	if err := run.Run(); err != nil {
		return err
	}
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := s.protoc(pwd, path.RawPath, majorVersion); err != nil {
		return err
	}
	return fmtProject(pwd)
}

// ImportModule imports specified module with name to the scaffolded app.
func (s *Scaffolder) ImportModule(name string) error {
	version, err := s.version()
	if err != nil {
		return err
	}
	majorVersion := version.Major()
	ok, err := isWasmImported(s.path)
	if err != nil {
		return err
	}
	if ok {
		return errors.New("wasm is already imported")
	}

	// import a specific version of ComsWasm
	err = installWasm(version)
	if err != nil {
		return err
	}

	path, err := gomodulepath.ParseAt(s.path)
	if err != nil {
		return err
	}

	// run generator
	var g *genny.Generator
	if majorVersion == cosmosver.Launchpad {
		g, err = module_import.NewImportLaunchpad(&module_import.ImportOptions{
			Feature: name,
			AppName: path.Package,
		})
	} else {
		g, err = module_import.NewImportStargate(&module_import.ImportOptions{
			Feature:          name,
			AppName:          path.Package,
			BinaryNamePrefix: path.Root,
		})
	}

	if err != nil {
		return err
	}
	run := genny.WetRunner(context.Background())
	run.With(g)
	if err := run.Run(); err != nil {
		return err
	}
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return fmtProject(pwd)
}

func ModuleExists(appPath string, moduleName string) (bool, error) {
	abspath, err := filepath.Abs(filepath.Join(appPath, moduleDir, moduleName))
	if err != nil {
		return false, err
	}

	_, err = os.Stat(abspath)
	if err == nil {
		// The module already exists
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	// Error reading the directory
	return false, err
}

func isWasmImported(appPath string) (bool, error) {
	abspath, err := filepath.Abs(filepath.Join(appPath, apppkg))
	if err != nil {
		return false, err
	}
	fset := token.NewFileSet()
	all, err := parser.ParseDir(fset, abspath, func(os.FileInfo) bool { return true }, parser.ImportsOnly)
	if err != nil {
		return false, err
	}
	for _, pkg := range all {
		for _, f := range pkg.Files {
			for _, imp := range f.Imports {
				if strings.Contains(imp.Path.Value, wasmImport) {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func installWasm(version cosmosver.Version) error {
	switch version {
	case cosmosver.LaunchpadAny:
		return cmdrunner.
			New(
				cmdrunner.DefaultStderr(os.Stderr),
			).
			Run(context.Background(),
				step.New(
					step.Exec(
						"go",
						"get",
						wasmImport+"@"+wasmVersionCommitLaunchpad,
					),
				),
			)
	case cosmosver.StargateZeroFourtyAndAbove:
		return cmdrunner.
			New(
				cmdrunner.DefaultStderr(os.Stderr),
			).
			Run(context.Background(),
				step.New(
					step.Exec(
						"go",
						"get",
						wasmImport+"@"+wasmVersionCommitStargate,
					),
				),
			)
	default:
		return errors.New("version not supported")
	}
}
