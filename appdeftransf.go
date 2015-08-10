package appdeftransf

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/giantswarm/generic-types-go"
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
	genericComponents := map[string]map[string]interface{}{
		"components": map[string]interface{}{},
	}

	nameKey := func(serviceName, componentName string) string {
		return serviceName + "/" + componentName
	}

	// Create component names for each component
	componentNameMap := make(map[string]string)
	portSequence := 8000
	portMap := map[string]string{}
	exposeMap := map[string]userconfig.ExposeDefinitions{}

	for _, service := range v1AppDef.Services {
		eds := userconfig.ExposeDefinitions{}

		for _, component := range service.Components {
			var componentName string
			if component.PodName != "" {
				componentName = service.ServiceName + "/" + component.PodName + "/" + component.ComponentName
			} else {
				componentName = service.ServiceName + "/" + component.ComponentName
			}
			key := nameKey(service.ServiceName, component.ComponentName)
			componentNameMap[key] = componentName

			for _, port := range component.Ports {
				portSequence++
				portMap[componentName] = strconv.Itoa(portSequence) + "/tcp"

				ed := userconfig.ExposeDefinition{
					TargetPort: port,
					Port:       generictypes.MustParseDockerPort(strconv.Itoa(portSequence)),
					Component:  userconfig.ComponentName(componentName),
				}

				eds = append(eds, ed)
			}
		}

		exposeMap[service.ServiceName] = eds
	}

	// Create components for all components
	podsCreated := make(map[string]string)
	for _, service := range v1AppDef.Services {
		for _, component := range service.Components {
			key := nameKey(service.ServiceName, component.ComponentName)
			componentName := componentNameMap[key]

			// Create pod component if needed
			if component.PodName != "" {
				podComponentName := service.ServiceName + "/" + component.PodName
				if _, ok := podsCreated[podComponentName]; !ok {
					// Need to create pod component
					podComponent := map[string]interface{}{}
					podComponent["pod"] = "children"

					// Add pod component
					genericComponents["components"][podComponentName] = podComponent

					// Record that we created this component so we don't duplicate t
					podsCreated[podComponentName] = podComponentName
				}

				// Remove the podname so it will not be part of the v2 def
				component.PodName = ""
			}

			rawComponent, err := json.Marshal(component)
			if err != nil {
				return userconfig.V2AppDefinition{}, maskAny(err)
			}

			var v1GenericComponent map[string]interface{}
			if err := json.Unmarshal(rawComponent, &v1GenericComponent); err != nil {
				return userconfig.V2AppDefinition{}, maskAny(err)
			}

			genericComponent := map[string]interface{}{}
			for key, val := range v1GenericComponent {
				if key == "component_name" {
					continue
				}

				if key == "dependencies" {
					if rawDeps, ok := val.([]interface{}); ok {
						for i, rawDep := range rawDeps {
							if m, ok := rawDep.(map[string]interface{}); ok {
								// Convert name to component
								serviceName, componentName := userconfig.ParseDependency(service.ServiceName, m["name"].(string))
								depName := componentNameMap[nameKey(serviceName, componentName)]
								m["component"] = strings.Split(depName, "/")[0]
								delete(m, "name")
								// Convert port to target_port
								m["target_port"] = portMap[depName]
								delete(m, "port")
								rawDeps[i] = m
							}
						}

						val = rawDeps
					}

					genericComponent["links"] = val
					continue
				}

				if key == "volumes" {
					if rawVols, ok := val.([]interface{}); ok {
						for i, rawVol := range rawVols {
							if m, ok := rawVol.(map[string]interface{}); ok {
								if volumeFromRaw, ok := m["volume-from"]; ok {
									serviceName, componentName := userconfig.ParseDependency(service.ServiceName, volumeFromRaw.(string))
									m["volume-from"] = componentNameMap[nameKey(serviceName, componentName)]
								}
								if volumesFromRaw, ok := m["volumes-from"]; ok {
									serviceName, componentName := userconfig.ParseDependency(service.ServiceName, volumesFromRaw.(string))
									m["volumes-from"] = componentNameMap[nameKey(serviceName, componentName)]
								}
								rawVols[i] = m
							}
						}

						val = rawVols
					}

					genericComponent["volumes"] = val
					continue
				}

				if key == "scaling_policy" {
					genericComponent["scale"] = val
					continue
				}

				genericComponent[key] = val
			}

			genericComponents["components"][componentName] = genericComponent
		}

		if len(exposeMap[service.ServiceName]) > 0 {
			genericComponents["components"][service.ServiceName] = map[string]interface{}{
				"expose": exposeMap[service.ServiceName],
			}
		}
	}

	raw, err := json.Marshal(genericComponents)
	if err != nil {
		return userconfig.V2AppDefinition{}, maskAny(err)
	}

	var v2AppDef userconfig.V2AppDefinition
	if err := json.Unmarshal(raw, &v2AppDef); err != nil {
		return userconfig.V2AppDefinition{}, maskAny(err)
	}

	return v2AppDef, nil
}
