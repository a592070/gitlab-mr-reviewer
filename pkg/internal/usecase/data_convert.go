package usecase

import (
	"gitlab-mr-reviewer/pkg/internal/domain"
	"gitlab-mr-reviewer/pkg/internal/repository"
)

func toMergeRequestDomain(mergeRequest *repository.MergeRequestDto, diffs []repository.DiffDto) *domain.MergeRequest {
	relativeChanges := make([]domain.RelativeChange, len(diffs))
	for i, diff := range diffs {
		relativeChanges[i] = domain.RelativeChange{
			Diff:        diff.Diff,
			NewPath:     diff.NewPath,
			OldPath:     diff.OldPath,
			NewFile:     diff.NewFile,
			RenameFile:  diff.RenameFile,
			DeletedFile: diff.DeletedFile,
		}
	}

	return domain.NewMergeRequest(
		mergeRequest.Id,
		mergeRequest.ProjectId,
		mergeRequest.Title,
		mergeRequest.Description,
		&domain.DifferentReference{
			BaseSha: mergeRequest.DiffRefs.BaseSha,
			HeadSha: mergeRequest.DiffRefs.HeadSha,
		},
		relativeChanges)
}

func toCreateMergeRequestSummaryInput(mergeRequest *domain.MergeRequest) repository.CreateMergeRequestSummaryInput {
	return repository.CreateMergeRequestSummaryInput{
		ProjectId:          mergeRequest.ProjectID,
		MergeRequestId:     mergeRequest.ID,
		RelativeChangeNote: string(mergeRequest.RelativeChangeNote),
		SummaryNote:        string(mergeRequest.SummaryNote),
	}
}
