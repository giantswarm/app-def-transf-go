package appdeftransf

import (
	"github.com/giantswarm/user-config"
	"github.com/juju/errgo"
)

// DefType is the app definition type an app can have.
type DefType string

const (
	DefTypeV2GiantSwarm DefType = "V2GiantSwarm"
)

// ParseTypeFromBytes tries to find out what kind of app definition is given by
// b. If no proper type can be detected, it returns an error. Valid definition
// types are: DefTypeV1GiantSwarm.
func ParseTypeFromBytes(b []byte) (DefType, error) {
	dtCheckers := []defTypeChecker{
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
		return "", maskAny(errgo.WithCausef(nil, InvalidDefTypeErr, "expecting %s", DefTypeV2GiantSwarm))
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
	case DefTypeV2GiantSwarm:
		serviceName, err := userconfig.ParseServiceName(b)
		if err != nil {
			return "", maskAny(err)
		}

		return serviceName, nil
	}

	return "", maskAny(errgo.WithCausef(nil, InvalidDefTypeErr, "expecting %s", DefTypeV2GiantSwarm))
}
