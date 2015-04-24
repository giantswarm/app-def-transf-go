package appdeftransf

import (
	"encoding/json"

	"github.com/giantswarm/user-config"
	"github.com/juju/errgo"
	"github.com/op/go-logging"
)

const (
	DefTypeV1GiantSwarm = "V1GiantSwarm"
)

type Conf struct{}

type Deps struct {
	Logger *logging.Logger
}

// AppDefTransf stands for App-Definition-Transformer and is responsible for
// transforming data structures to other data structures.
type AppDefTransf struct {
	Conf
	Deps
}

func NewAppDefTransf(c Conf, d Deps) *AppDefTransf {
	return &AppDefTransf{
		Conf: c,
		Deps: d,
	}
}

func (adt *AppDefTransf) ParseDefAndType(i interface{}) ([]byte, string, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return nil, "", Mask(err)
	}

	t, err := adt.ParseTypeFromBytes(b)
	if err != nil {
		return nil, "", Mask(err)
	}

	return b, t, nil
}

func (adt *AppDefTransf) ParseTypeFromBytes(b []byte) (string, error) {
	if _, err := userconfig.ParseV1AppDefinition(b); err == nil {
		return "V1GS", nil
	}

	return "", errgo.Newf("Invalid app definition.")
}

func (adt *AppDefTransf) ParseName(i interface{}) (string, error) {
	b, t, err := adt.ParseDefAndType(i)
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
