package appdeftransf

import (
	"encoding/json"

	"github.com/giantswarm/user-config"
	"github.com/juju/errgo"
)

// DefType is the app definition type an app can have.
type DefType string

const (
	DefTypeV1GiantSwarm DefType = "V1GiantSwarm"
	DefTypeV2GiantSwarm DefType = "V2GiantSwarm"
)

// ParseTypeFromBytes tries to find out what kind of app definition is given by
// b. If no proper type can be detected, it returns an error. Valid definition
// types are: DefTypeV1GiantSwarm.
func ParseTypeFromBytes(b []byte) (DefType, error) {
	dtCheckers := []defTypeChecker{
		newV1GiantSwarmDefTypeChecker(),
		newV2GiantSwarmDefTypeChecker(),
	}

	var finType DefType = ""
	var finProb float64 = 0.0

	for _, dtChecker := range dtCheckers {
		dtCheckerType, dtCheckerProb := dtChecker.Parse(b)

		if dtCheckerProb > finProb {
			finProb = dtCheckerProb
			finType = dtCheckerType
		}
	}

	if finProb == 0.0 {
		return "", maskAny(errgo.WithCausef(nil, InvalidDefTypeErr, "expecting %s or %s", DefTypeV1GiantSwarm, DefTypeV2GiantSwarm))
	}

	return finType, nil
}

// ParseName tries to find out what name an app given by b might has.
// Internally it calls ParseTypeFromBytes.
func ParseName(b []byte) (string, error) {
	t, err := ParseTypeFromBytes(b)
	if err != nil {
		return "", maskAny(err)
	}

	switch t {
	case DefTypeV1GiantSwarm:
		def, err := userconfig.ParseV1AppDefinition(b)
		if err != nil {
			return "", maskAny(err)
		}

		return def.AppName, nil
	case DefTypeV2GiantSwarm:
		appName, err := userconfig.V2GenerateAppName(b)
		if err != nil {
			return "", maskAny(err)
		}

		return appName, nil
	}

	return "", maskAny(errgo.WithCausef(nil, InvalidDefTypeErr, "expecting %s or %s", DefTypeV1GiantSwarm, DefTypeV2GiantSwarm))
}

func V1GiantSwarmToV2GiantSwarm(v1AppDef userconfig.AppDefinition) (userconfig.V2AppDefinition, error) {
	genericNodes := map[string]map[string]interface{}{
		"nodes": map[string]interface{}{},
	}

	for _, service := range v1AppDef.Services {
		for _, component := range service.Components {
			nodeName := service.ServiceName + "/" + component.ComponentName

			rawComponent, err := json.Marshal(component)
			if err != nil {
				return userconfig.V2AppDefinition{}, maskAny(err)
			}

			var genericComponent map[string]interface{}
			if err := json.Unmarshal(rawComponent, &genericComponent); err != nil {
				return userconfig.V2AppDefinition{}, maskAny(err)
			}

			genericNode := map[string]interface{}{}
			for key, val := range genericComponent {
				if key == "component_name" {
					continue
				}

				if key == "dependencies" {
					if rawDeps, ok := val.([]interface{}); ok {
						for i, rawDep := range rawDeps {
							if m, ok := rawDep.(map[string]interface{}); ok {
								serviceName, componentName := userconfig.ParseDependency(service.ServiceName, m["name"].(string))
								m["name"] = serviceName + "/" + componentName
								rawDeps[i] = m
							}
						}

						val = rawDeps
					}

					genericNode["links"] = val
					continue
				}

				if key == "scaling_policy" {
					genericNode["scale"] = val
					continue
				}

				genericNode[key] = val
			}

			genericNodes["nodes"][nodeName] = genericNode
		}
	}

	raw, err := json.Marshal(genericNodes)
	if err != nil {
		return userconfig.V2AppDefinition{}, maskAny(err)
	}

	var v2AppDef userconfig.V2AppDefinition
	if err := json.Unmarshal(raw, &v2AppDef); err != nil {
		return userconfig.V2AppDefinition{}, maskAny(err)
	}

	return v2AppDef, nil
}
