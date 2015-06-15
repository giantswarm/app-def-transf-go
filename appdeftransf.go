package appdeftransf

import (
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
		return "", mask(invalidDefTypeErr)
	}

	return finType, nil
}

// ParseName tries to find out what name an app given by b might has.
// Internally it calls ParseTypeFromBytes.
func ParseName(b []byte) (string, error) {
	t, err := ParseTypeFromBytes(b)
	if err != nil {
		return "", mask(err)
	}

	switch t {
	case DefTypeV1GiantSwarm:
		def, err := userconfig.ParseV1AppDefinition(b)
		if err != nil {
			return "", mask(err)
		}

		return def.AppName, nil
	case DefTypeV2GiantSwarm:
		appName, err := userconfig.ParseV2AppName(b)
		if err != nil {
			return "", mask(err)
		}

		return appName, nil
	}

	return "", errgo.Newf("cannot parse app definition, expecting: %s, %s", DefTypeV1GiantSwarm, DefTypeV2GiantSwarm)
}
