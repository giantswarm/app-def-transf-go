package appdeftransf

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAppDefTransf(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AppDefTransf")
}
