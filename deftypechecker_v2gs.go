package appdeftransf

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
	"github.com/juju/errgo"
)

// simple app definition

type simpleV2GiantSwarmComponentDef struct {
	Image  string                    `json:"image"`
	Ports  []generictypes.DockerPort `json:"ports"`
	Expose map[string]interface{}    `json:"expose"`
	Scale  map[string]interface{}    `json:"scale"`
}

type simpleV2GiantSwarmAppDef struct {
	ServiceName string                                    `json:"name,omitempty"`
	Components  map[string]simpleV2GiantSwarmComponentDef `json:"components"`
}

// def type checker

type v2GiantSwarmCheck func(simpleV2GiantSwarmAppDef) bool

type v2GiantSwarmDefTypeChecker struct {
	checks []v2GiantSwarmCheck
}

func newV2GiantSwarmDefTypeChecker() defTypeChecker {
	checker := v2GiantSwarmDefTypeChecker{}

	checker.checks = []v2GiantSwarmCheck{
		checker.hasServiceName,
		checker.hasComponents,
		checker.hasComponentName,
		checker.hasComponentImage,
		checker.hasComponentPorts,
		checker.hasComponentExpose,
		checker.hasComponentScale,
	}

	return checker
}

func (dtc v2GiantSwarmDefTypeChecker) Parse(b []byte) (DefType, float64) {
	prob := 0.0

	// On syntax errors we need to check the raw definition. In case we find
	// important keywords we just assume to have a higher probability to deal
	// with a v2 app def.
	match, err := regexp.Match(`"components"(\s+)?:`, b)
	if err != nil {
		return DefTypeV2GiantSwarm, 0.0
	} else if match {
		prob += 10.0
	}

	// In case we have a syntax error which message contains a dollar char ($),
	// we guess the current definition is a swarm.json, and has unparsed
	// variables in it. Because this is only one indicator, we cannot be more
	// sure and just assume the probability that we are right or wrong with our
	// guess is 50%.
	_, err = userconfig.ParseServiceDefinition(b)
	if userconfig.IsSyntax(err) && strings.Contains(errgo.Cause(err).Error(), "$") {
		prob += 50.0
		return DefTypeV2GiantSwarm, prob
	}

	var simpleDef simpleV2GiantSwarmAppDef
	if err := json.Unmarshal(b, &simpleDef); err != nil {
		return DefTypeV2GiantSwarm, prob
	}

	passed := 0
	for _, check := range dtc.checks {
		if check(simpleDef) == true {
			passed++
		}
	}

	return DefTypeV2GiantSwarm, round(float64(passed * 100 / len(dtc.checks)))
}

// private checker

func (dtc v2GiantSwarmDefTypeChecker) hasServiceName(simpleDef simpleV2GiantSwarmAppDef) bool {
	return simpleDef.ServiceName != ""
}

func (dtc v2GiantSwarmDefTypeChecker) hasComponents(simpleDef simpleV2GiantSwarmAppDef) bool {
	return len(simpleDef.Components) > 0
}

func (dtc v2GiantSwarmDefTypeChecker) hasComponentName(simpleDef simpleV2GiantSwarmAppDef) bool {
	return dtc.iterateComponents(simpleDef, func(name string, component simpleV2GiantSwarmComponentDef) bool {
		return name != ""
	})
}

func (dtc v2GiantSwarmDefTypeChecker) hasComponentImage(simpleDef simpleV2GiantSwarmAppDef) bool {
	return dtc.iterateComponents(simpleDef, func(name string, component simpleV2GiantSwarmComponentDef) bool {
		return component.Image != ""
	})
}

func (dtc v2GiantSwarmDefTypeChecker) hasComponentPorts(simpleDef simpleV2GiantSwarmAppDef) bool {
	return dtc.iterateComponents(simpleDef, func(name string, component simpleV2GiantSwarmComponentDef) bool {
		return len(component.Ports) > 0
	})
}

func (dtc v2GiantSwarmDefTypeChecker) hasComponentExpose(simpleDef simpleV2GiantSwarmAppDef) bool {
	return dtc.iterateComponents(simpleDef, func(name string, component simpleV2GiantSwarmComponentDef) bool {
		return len(component.Expose) > 0
	})
}

func (dtc v2GiantSwarmDefTypeChecker) hasComponentScale(simpleDef simpleV2GiantSwarmAppDef) bool {
	return dtc.iterateComponents(simpleDef, func(name string, component simpleV2GiantSwarmComponentDef) bool {
		return len(component.Scale) > 0
	})
}

// private helper

func (dtc v2GiantSwarmDefTypeChecker) iterateComponents(simpleDef simpleV2GiantSwarmAppDef, cb func(name string, component simpleV2GiantSwarmComponentDef) bool) bool {
	if !dtc.hasComponents(simpleDef) {
		return false
	}

	for name, component := range simpleDef.Components {
		return cb(name, component)
	}

	return false
}
