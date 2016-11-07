package vklp

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVkLP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "vklp (private)")
}
