package appdeftransf

import (
	"encoding/json"
	"strings"

	"github.com/giantswarm/user-config"
	"github.com/juju/errgo"
)

// simple app definition

type simpleV1GiantSwarmComponent struct {
	ComponentName string `json:"component_name"`
}

type simpleV1GiantSwarmService struct {
	ServiceName string                        `json:"service_name"`
	Components  []simpleV1GiantSwarmComponent `json:"components"`
}

type simpleV1GiantSwarmAppDef struct {
	AppName  string                      `json:"app_name"`
	Services []simpleV1GiantSwarmService `json:"services"`
}

// def type checker

type v1GiantSwarmCheck func(simpleV1GiantSwarmAppDef) bool

type v1GiantSwarmDefTypeChecker struct {
	checks []v1GiantSwarmCheck
}

func newV1GiantSwarmDefTypeChecker() defTypeChecker {
	checker := v1GiantSwarmDefTypeChecker{}

	checker.checks = []v1GiantSwarmCheck{
		checker.hasAppName,
		checker.hasServiceName,
		checker.hasComponentName,
		checker.hasServices,
		checker.hasComponents,
	}

	return checker
}

func (dtc v1GiantSwarmDefTypeChecker) Parse(b []byte) (DefType, float64) {
	// In case we have a syntax error which message contains a dollar char ($),
	// we guess the current definition is a swarm.json, and has unoarsed
	// variables in it. Because this is only one indicator, we cannot be more
	// sure and just assume the probability that we are right or wrong with our
	// guess is 50%.
	_, err := userconfig.ParseV1AppDefinition(b)
	if userconfig.IsSyntaxError(err) && strings.Contains(errgo.Cause(err).Error(), "$") {
		return DefTypeV1GiantSwarm, 50.0
	}

	var simpleDef simpleV1GiantSwarmAppDef
	if err := json.Unmarshal(b, &simpleDef); err != nil {
		return DefTypeV1GiantSwarm, 0.0
	}

	passed := 0
	for _, check := range dtc.checks {
		if check(simpleDef) == true {
			passed++
		}
	}

	return DefTypeV1GiantSwarm, round(float64(passed * 100 / len(dtc.checks)))
}

// private checker

func (dtc v1GiantSwarmDefTypeChecker) hasAppName(simpleDef simpleV1GiantSwarmAppDef) bool {
	return simpleDef.AppName != ""
}

func (dtc v1GiantSwarmDefTypeChecker) hasServiceName(simpleDef simpleV1GiantSwarmAppDef) bool {
	return dtc.hasServices(simpleDef) && simpleDef.Services[0].ServiceName != ""
}

func (dtc v1GiantSwarmDefTypeChecker) hasComponentName(simpleDef simpleV1GiantSwarmAppDef) bool {
	return dtc.hasComponents(simpleDef) && simpleDef.Services[0].Components[0].ComponentName != ""
}

func (dtc v1GiantSwarmDefTypeChecker) hasServices(simpleDef simpleV1GiantSwarmAppDef) bool {
	return len(simpleDef.Services) > 0
}

func (dtc v1GiantSwarmDefTypeChecker) hasComponents(simpleDef simpleV1GiantSwarmAppDef) bool {
	return dtc.hasServices(simpleDef) && len(simpleDef.Services[0].Components) > 0
}