package domain

import (
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"regexp"
	"testing"
)

func TestMergeRequest(t *testing.T) {
	gomega.RegisterTestingT(t)

	var _ = ginkgo.Describe("MergeRequest", func() {
		ginkgo.It("FilterIgnorePaths", func() {
			var mergeRequest *MergeRequest
			var pathFilters = []string{
				".*.mod",
				".*.sum",
				".*/gen/.*",
			}
			ginkgo.By(fmt.Sprintf("filter %v", pathFilters))
			filters := make([]*regexp.Regexp, len(pathFilters))
			for i, f := range pathFilters {
				regex, err := regexp.Compile(f)
				if err != nil {
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
				}
				filters[i] = regex
			}

			mergeRequest = NewMergeRequest(1, 1, "", "", &DifferentReference{}, []RelativeChange{
				{
					Diff:    "",
					NewPath: "go.mod",
					OldPath: "",
				},
				{
					Diff:    "",
					NewPath: "/app/go.mod",
					OldPath: "",
				},
				{
					Diff:    "",
					NewPath: "go.sum",
					OldPath: "",
				},
				{
					Diff:    "",
					NewPath: "/app/go.sum",
					OldPath: "",
				},
				{
					Diff:    "",
					NewPath: "/app/gen/generatefile",
					OldPath: "",
				},
				{
					Diff:    "",
					NewPath: "main.go",
					OldPath: "",
				},
			})

			mergeRequest.FilterIgnorePaths(filters)
			gomega.Expect(len(mergeRequest.RelativeChanges)).To(gomega.Equal(1))
		})

		ginkgo.It("IgnoreReview", func() {
			var mergeRequest *MergeRequest
			ginkgo.By("There are no RelativeChanges")
			mergeRequest = NewMergeRequest(1, 1, "", "", &DifferentReference{}, []RelativeChange{})
			gomega.Expect(mergeRequest.IgnoreReview()).To(gomega.BeTrue())

			ginkgo.By("Description contain ")
			mergeRequest = NewMergeRequest(1, 1,
				"",
				`> In Kubernetes' flow, a shift takes place,  
> ConfigMaps now fit with more grace.  
> With errors handled and paths aligned,  
> Sidecars move forward, redefined.  
> 🎨 A refactor polished, a feature now bright,  
> Our journey continues with delight! 🚀
@codeReview: ignore`,
				&DifferentReference{}, []RelativeChange{})
			gomega.Expect(mergeRequest.IgnoreReview()).To(gomega.BeTrue())
		})
	})

	ginkgo.RunSpecs(t, "MergeRequest")
}
