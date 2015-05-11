package appdeftransf

import (
	"github.com/giantswarm/user-config"
	"github.com/juju/errgo"
)

type DefType string

func (dt DefType) IsV1GiantSwarm() bool {
	return dt == DefTypeV1GiantSwarm
}

const (
	DefTypeV1GiantSwarm DefType = "V1GiantSwarm"
)

func ParseTypeFromBytes(b []byte) (DefType, error) {
	if _, err := userconfig.ParseV1AppDefinition(b); err == nil {
		return DefTypeV1GiantSwarm, nil
	}

	return "", errgo.Newf("Invalid app definition.")
}

func ParseName(b []byte) (string, error) {
	t, err := ParseTypeFromBytes(b)
	if err != nil {
		return "", Mask(err)
	}

	switch t {
	case DefTypeV1GiantSwarm:
		if def, err := userconfig.ParseV1AppDefinition(b); err == nil {
			return def.AppName, nil
		}
	}

	return "", errgo.Newf("Invalid app definition type '%s'. Expecting %s. Aborting...", t, DefTypeV1GiantSwarm)
}
