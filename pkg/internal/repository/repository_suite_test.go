package repository

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"testing"
)

func TestRepositorySuite(t *testing.T) {
	gomega.RegisterTestingT(t)
	ginkgo.RunSpecs(t, "Repository Suite")
}
