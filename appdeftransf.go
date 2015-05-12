package appdeftransf

import (
	"github.com/giantswarm/user-config"
	"github.com/juju/errgo"
)

type DefType string

const (
	DefTypeV1GiantSwarm DefType = "V1GiantSwarm"
)

func ParseTypeFromBytes(b []byte) (DefType, error) {
	dtCheckers := []defTypeChecker{
		newV1GiantSwarmDefTypeChecker(),
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
		return "", invalidDefTypeErr
	}

	return finType, nil
}

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
	}

	return "", errgo.Newf("Invalid app definition type '%s'. Expecting %s. Aborting...", t, DefTypeV1GiantSwarm)
}
