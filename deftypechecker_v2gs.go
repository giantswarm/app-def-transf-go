package appdeftransf

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/giantswarm/generic-types-go"
	"github.com/giantswarm/user-config"
	"github.com/juju/errgo"
)

// simple app definition

type simpleV2GiantSwarmNodeDef struct {
	Image  string                    `json:"image"`
	Ports  []generictypes.DockerPort `json:"ports"`
	Expose map[string]interface{}    `json:"expose"`
	Scale  map[string]interface{}    `json:"scale"`
}

type simpleV2GiantSwarmAppDef struct {
	Nodes map[string]simpleV2GiantSwarmNodeDef `json:"nodes"`
}

// def type checker

type v2GiantSwarmCheck func(simpleV2GiantSwarmAppDef) bool

type v2GiantSwarmDefTypeChecker struct {
	checks []v2GiantSwarmCheck
}

func newV2GiantSwarmDefTypeChecker() defTypeChecker {
	checker := v2GiantSwarmDefTypeChecker{}

	checker.checks = []v2GiantSwarmCheck{
		checker.hasNodes,
		checker.hasNodeName,
		checker.hasServiceImage,
		checker.hasServicePorts,
		checker.hasNodeExpose,
		checker.hasNodeScale,
	}

	return checker
}

func (dtc v2GiantSwarmDefTypeChecker) Parse(b []byte) (DefType, float64) {
	prob := 0.0

	// On syntax errors we need to check the raw definition. In case we find
	// important keywords we just assume to have a higher probability to deal
	// with a v2 app def.
	match, err := regexp.Match(`"nodes"(\s+)?:`, b)
	if err != nil {
		fmt.Printf("%#v\n", errgo.New("cannot parse v2 app definition: regexp.Match failed badly"))
		return DefTypeV2GiantSwarm, 0.0
	} else if match {
		prob += 10.0
	}

	// In case we have a syntax error which message contains a dollar char ($),
	// we guess the current definition is a swarm.json, and has unparsed
	// variables in it. Because this is only one indicator, we cannot be more
	// sure and just assume the probability that we are right or wrong with our
	// guess is 50%.
	_, err = userconfig.ParseV2AppDefinition(b)
	if userconfig.IsSyntaxError(err) && strings.Contains(errgo.Cause(err).Error(), "$") {
		prob += 50.0
		return DefTypeV2GiantSwarm, prob
	}

	var simpleDef simpleV2GiantSwarmAppDef
	if err := json.Unmarshal(b, &simpleDef); err != nil {
		fmt.Printf("%#v\n", errgo.New("cannot parse v2 app definition: json.Unmarshal failed badly"))
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

func (dtc v2GiantSwarmDefTypeChecker) hasNodes(simpleDef simpleV2GiantSwarmAppDef) bool {
	return len(simpleDef.Nodes) > 0
}

func (dtc v2GiantSwarmDefTypeChecker) hasNodeName(simpleDef simpleV2GiantSwarmAppDef) bool {
	return dtc.iterateNodes(simpleDef, func(name string, node simpleV2GiantSwarmNodeDef) bool {
		return name != ""
	})
}

func (dtc v2GiantSwarmDefTypeChecker) hasServiceImage(simpleDef simpleV2GiantSwarmAppDef) bool {
	return dtc.iterateNodes(simpleDef, func(name string, node simpleV2GiantSwarmNodeDef) bool {
		return node.Image != ""
	})
}

func (dtc v2GiantSwarmDefTypeChecker) hasServicePorts(simpleDef simpleV2GiantSwarmAppDef) bool {
	return dtc.iterateNodes(simpleDef, func(name string, node simpleV2GiantSwarmNodeDef) bool {
		return len(node.Ports) > 0
	})
}

func (dtc v2GiantSwarmDefTypeChecker) hasNodeExpose(simpleDef simpleV2GiantSwarmAppDef) bool {
	return dtc.iterateNodes(simpleDef, func(name string, node simpleV2GiantSwarmNodeDef) bool {
		return len(node.Expose) > 0
	})
}

func (dtc v2GiantSwarmDefTypeChecker) hasNodeScale(simpleDef simpleV2GiantSwarmAppDef) bool {
	return dtc.iterateNodes(simpleDef, func(name string, node simpleV2GiantSwarmNodeDef) bool {
		return len(node.Scale) > 0
	})
}

// private helper

func (dtc v2GiantSwarmDefTypeChecker) iterateNodes(simpleDef simpleV2GiantSwarmAppDef, cb func(name string, node simpleV2GiantSwarmNodeDef) bool) bool {
	if !dtc.hasNodes(simpleDef) {
		return false
	}

	for name, node := range simpleDef.Nodes {
		return cb(name, node)
	}

	return false
}
