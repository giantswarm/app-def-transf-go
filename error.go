package appdeftransf

import (
	"github.com/juju/errgo"
)

var (
	InvalidDefTypeErr = errgo.Newf("invalid definition type")

	maskAny = errgo.MaskFunc(errgo.Any)
)

func IsInvalidDefType(err error) bool {
	return errgo.Cause(err) == InvalidDefTypeErr
}
