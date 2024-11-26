package repository

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"gitlab-mr-reviewer/pkg/logging"
)

var _ = ginkgo.Describe("Openai Repository", ginkgo.Ordered, func() {
	var logger *logging.ZaprLogger

	ginkgo.BeforeAll(func() {
		var err error
		logger, err = logging.NewZaprLogger(false, "info")
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("Always success", func() {
		logger.Info("Hello World")
		gomega.Expect(logger).ToNot(gomega.BeNil())
	})

})
