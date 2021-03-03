package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pallinder/go-randomdata"
	"github.com/zhigui-projects/zeus-onestop/starport/internal/version"
)

const (
	loginAny = "any"
)

var (
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
