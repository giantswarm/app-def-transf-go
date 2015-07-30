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
		appName, err := userconfig.V2AppName(b)
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

	nameKey := func(serviceName, componentName string) string {
		return serviceName + "/" + componentName
	}

	// Create node names for each component
	nodeNameMap := make(map[string]string)
	for _, service := range v1AppDef.Services {
		for _, component := range service.Components {
			var nodeName string
			if component.PodName != "" {
				nodeName = service.ServiceName + "/" + component.PodName + "/" + component.ComponentName
			} else {
				nodeName = service.ServiceName + "/" + component.ComponentName
			}
			key := nameKey(service.ServiceName, component.ComponentName)
			nodeNameMap[key] = nodeName
		}
	}

	// Create nodes for all components
	podsCreated := make(map[string]string)
	for _, service := range v1AppDef.Services {
		for _, component := range service.Components {
			key := nameKey(service.ServiceName, component.ComponentName)
			nodeName := nodeNameMap[key]

			// Create pod node if needed
			if component.PodName != "" {
				podNodeName := service.ServiceName + "/" + component.PodName
				if _, ok := podsCreated[podNodeName]; !ok {
					// Need to create pod node
					podNode := map[string]interface{}{}
					podNode["pod"] = "children"

					// Add pod node
					genericNodes["nodes"][podNodeName] = podNode

					// Record that we created this node so we don't duplicate t
					podsCreated[podNodeName] = podNodeName
				}

				// Remove the podname so it will not be part of the v2 def
				component.PodName = ""
			}

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
								// Convert name to node
								serviceName, componentName := userconfig.ParseDependency(service.ServiceName, m["name"].(string))
								m["node"] = nodeNameMap[nameKey(serviceName, componentName)]
								delete(m, "name")
								// Convert port to target_port
								m["target_port"] = m["port"]
								delete(m, "port")
								rawDeps[i] = m
							}
						}

						val = rawDeps
					}

					genericNode["links"] = val
					continue
				}

				if key == "volumes" {
					if rawVols, ok := val.([]interface{}); ok {
						for i, rawVol := range rawVols {
							if m, ok := rawVol.(map[string]interface{}); ok {
								if volumeFromRaw, ok := m["volume-from"]; ok {
									serviceName, componentName := userconfig.ParseDependency(service.ServiceName, volumeFromRaw.(string))
									m["volume-from"] = nodeNameMap[nameKey(serviceName, componentName)]
								}
								if volumesFromRaw, ok := m["volumes-from"]; ok {
									serviceName, componentName := userconfig.ParseDependency(service.ServiceName, volumesFromRaw.(string))
									m["volumes-from"] = nodeNameMap[nameKey(serviceName, componentName)]
								}
								rawVols[i] = m
							}
						}

						val = rawVols
					}

					genericNode["volumes"] = val
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
