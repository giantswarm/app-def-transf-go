package appdeftransf

import (
	"github.com/juju/errgo"
)

var (
	invalidDefTypeErr = errgo.Newf("Invalid definition type.")

	mask = errgo.MaskFunc(errgo.Any)
)
