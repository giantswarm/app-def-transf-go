package appdeftransf

import (
	"github.com/juju/errgo"
)

var (
	InvalidDefTypeErr = errgo.Newf("Invalid definition type.")

	mask = errgo.MaskFunc(errgo.Any)
)

func IsInvalidDefType(err error) bool {
	return errgo.Cause(err) == InvalidDefTypeErr
}
