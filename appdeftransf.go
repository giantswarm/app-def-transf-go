package appdeftransf

import (
	"encoding/json"

	"github.com/giantswarm/user-config"
	"github.com/juju/errgo"
)

const (
	DefTypeV1GiantSwarm = "V1GiantSwarm"
)

func ParseTypeFromBytes(b []byte) (string, error) {
	if _, err := userconfig.ParseV1AppDefinition(b); err == nil {
		return "V1GS", nil
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
