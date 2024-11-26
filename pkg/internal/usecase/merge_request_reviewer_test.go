package usecase

import (
	"context"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"gitlab-mr-reviewer/pkg/internal/domain"
	"gitlab-mr-reviewer/pkg/internal/repository"
	"gitlab-mr-reviewer/pkg/logging"
	"testing"
)

type mockGitlabRepository struct {
}

func (m *mockGitlabRepository) ListDiffByMergeRequestId(ctx context.Context, projectId, mergeRequestId int32) ([]repository.DiffDto, error) {
	diffs := []repository.DiffDto{
		{
			Diff:        "@@ -0,0 +1,9 @@\\n+package mutation\\n+\\n+import \\\"github.com/pkg/errors\\\"\\n+\\n+var (\\n+\\tErrorFailedCreateConfigmap = errors.New(\\\"Failed to create configmap\\\")\\n+\\tErrorConfigmapExists       = errors.New(\\\"Configmap already exists\\\")\\n+\\tErrorConfigmapNotFound     = errors.New(\\\"Configmap not found\\\")\\n+)\\n",
			NewPath:     "internal/mutation/errors.go",
			OldPath:     "internal/mutation/errors.go",
			NewFile:     true,
			RenameFile:  false,
			DeletedFile: false,
		},
	}
	return diffs, nil
}

func (m *mockGitlabRepository) GetMergeRequest(ctx context.Context, projectId, mergeRequestId int32) (*repository.MergeRequestDto, error) {
	mr := &repository.MergeRequestDto{
		ProjectId:   projectId,
		Id:          mergeRequestId,
		Title:       "This is a test",
		Description: "",
		DiffRefs: repository.DiffRefsDto{
			BaseSha: "63c97907ceb628e4fbbc125ca9ce5bd2a1f9566f",
			HeadSha: "94e7e0bb7144018e544743e1d6f22731f8ddeba1",
		},
	}

	switch mr.Id {
	case 0:
		mr.Description += "@codeReview: ignore"
	}

	return mr, nil
}

func (m *mockGitlabRepository) CreateMergeRequestSummary(ctx context.Context, input repository.CreateMergeRequestSummaryInput) error {
	return nil
}

type mockOpenaiRepository struct {
	relativeChangesSummary string
	releaseNoteSummary     string
}

func (m *mockOpenaiRepository) SummarizeRelativeChanges(ctx context.Context, input repository.SummarizeRelativeChangesInput) (repository.SummarizeRelativeChangesOutput, error) {
	return repository.SummarizeRelativeChangesOutput{
		Messages: []domain.Message{
			{
				Role:    domain.RoleAssistant,
				Content: m.relativeChangesSummary,
			},
		},
	}, nil
}

func (m *mockOpenaiRepository) SummarizeReleaseNote(ctx context.Context, input repository.SummarizeReleaseNoteInput) (repository.SummarizeReleaseNoteOutput, error) {
	return repository.SummarizeReleaseNoteOutput{
		Messages: []domain.Message{
			{
				Role:    domain.RoleAssistant,
				Content: m.releaseNoteSummary,
			},
		},
	}, nil
}

func TestMergeRequestReviewer(t *testing.T) {
	gomega.RegisterTestingT(t)

	var _ = ginkgo.Describe("MergeRequestReviewer", ginkgo.Ordered, func() {
		var logger *logging.ZaprLogger
		var mergerRequestReviewer MergeRequestReviewer

		relativeChangesSummary := "## Summary(Fake Response)\n\nThis merge request introduces a `ConfigMapRepository` interface for Kubernetes API interaction and refactors the sidecar mutator logic, focusing on improved error handling and maintainability. The update enhances performance and reliability by centralizing error management with defined error variables and replacing hardcoded configuration strings with constants.\n\n| Files / Grouped Changes                                           | Summary                                                                                     |\n|-------------------------------------------------------------------|---------------------------------------------------------------------------------------------|\n| `internal/mutation/repository/configmap_repository.go`, `configmap_repository_impl.go` | Implements `ConfigMapRepository` interface for standardized configmap operations.          |\n| `beyla_sidecar_mutator.go`, `pod_webhook_handler.go`, `pod_webhook_handler_test.go` | Refactors mutator logic for better error handling, logging, and utilizes the new repository interface. |\n| `mutation/configmap_mutator.go`                                   | Removes redundant code replaced by the new `ConfigMapRepository` structure.                 |\n| `mutation/constants.go`                                           | Updates configuration path constants for better consistency and manageability.             |\n| `mutation/errors.go`                                              | Introduces custom error variables for precise error management during configmap operations.|\n| `pkg/dependencies_injection/injector.go`                          | Updates dependency injection to utilize the new `ConfigMapRepository`.                     |\n"
		releaseNoteSummary := "### Release Notes(Fake Response)\n\n**New Feature:**\n- Introduced a `ConfigMapRepository` interface to streamline interactions with Kubernetes' API.\n\n**Refactor:**\n- Enhanced error handling and logging in the sidecar mutator by leveraging the new repository.\n- Replaced hardcoded configuration paths with constants for improved maintainability.\n\n**Chore:**\n- Added defined error variables for better error management in configmap operations.\n\n> In Kubernetes' flow, a shift takes place,  \n> ConfigMaps now fit with more grace.  \n> With errors handled and paths aligned,  \n> Sidecars move forward, redefined.  \n> 🎨 A refactor polished, a feature now bright,  \n> Our journey continues with delight! 🚀"

		ginkgo.BeforeAll(func() {
			var err error
			var systemMessage string
			var gitlabRepository repository.GitlabRepository
			var llmRepository repository.LLMRepository
			var pathFilters = []string{}

			logger, err = logging.NewZaprLogger(false, "info")
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			gitlabRepository = &mockGitlabRepository{}
			llmRepository = &mockOpenaiRepository{
				relativeChangesSummary: relativeChangesSummary,
				releaseNoteSummary:     releaseNoteSummary,
			}

			mergerRequestReviewer, err = NewGitlabMergeRequestReviewer(logger, systemMessage, pathFilters, gitlabRepository, llmRepository)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		})

		ginkgo.It("Should success", func() {
			ctx := context.Background()

			output, err := mergerRequestReviewer.Apply(ctx, &MergeRequestReviewInput{
				ProjectId:      1,
				MergeRequestId: 1,
				Model:          "gpt-4o-mini",
			})

			ginkgo.By("output should not error")
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(output).ToNot(gomega.BeNil())

			gomega.Expect(output.SummarizeReleaseNote).To(gomega.Equal(releaseNoteSummary))
			gomega.Expect(output.SummarizeRelativeChanges).To(gomega.Equal(relativeChangesSummary))

		})
		ginkgo.It("Should ignore code review", func() {
			ctx := context.Background()

			_, err := mergerRequestReviewer.Apply(ctx, &MergeRequestReviewInput{
				ProjectId:      0,
				MergeRequestId: 0,
				Model:          "gpt-4o-mini",
			})

			ginkgo.By("should got an error")
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err).To(gomega.Equal(ErrorIgnoreCodeReview))
		})

	})

	ginkgo.RunSpecs(t, "MergeRequestReviewer test")

}
