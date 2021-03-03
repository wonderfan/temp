package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pallinder/go-randomdata"
	"github.com/zhigui-projects/zeus-onestop/starport/internal/version"
	"github.com/zhigui-projects/zeus-onestop/starport/pkg/gacli"
)

const (
	gaid     = "UA-51029217-18" // Google Analytics' tracking id.
	loginAny = "any"
)

var (
	gaclient             *gacli.Client
	starportDir          = ".starport"
	starportAnonIdentity = "anon"
)

// Metric represents an analytics metric.
type Metric struct {
	// IsInstallation sets metrics type as an installation metric.
	IsInstallation bool

	// Err sets metrics type as an error metric.
	Err error

	// Login is the name of anon user.
	Login string
}

func addMetric(m Metric) {
	fullCommand := os.Args
	var rootCommand string
	if len(os.Args) > 1 { // first is starport (binary name).
		rootCommand = os.Args[1]
	}

	var met gacli.Metric
	switch {
	case m.IsInstallation:
		met.Category = "install"
	case m.Err == nil:
		met.Category = "success"
	case m.Err != nil:
		met.Category = "error"
		met.Value = m.Err.Error()
	}
	if m.IsInstallation {
		met.Action = m.Login
	} else {
		met.Action = rootCommand
		met.Label = strings.Join(fullCommand, " ")
	}
	user, _ := prepLoginName()
	met.User = user
	met.Version = version.Version
	gaclient.Send(met)
}

func prepLoginName() (name string, hadLogin bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		return loginAny, false
	}
	if err := os.Mkdir(filepath.Join(home, starportDir), 0700); err != nil {
		return loginAny, false
	}
	anonPath := filepath.Join(home, starportDir, starportAnonIdentity)
	data, err := ioutil.ReadFile(anonPath)
	if err != nil {
		return loginAny, false
	}
	if len(data) != 0 {
		return string(data), true
	}
	name = randomdata.SillyName()
	if err := ioutil.WriteFile(anonPath, []byte(name), 0700); err != nil {
		return loginAny, false
	}
	return name, false
}
