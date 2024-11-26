package repository

import (
	"context"
	"gitlab-mr-reviewer/pkg/internal/domain"
)

type LLMRepository interface {
	SummarizeRelativeChanges(ctx context.Context, input SummarizeRelativeChangesInput) (SummarizeRelativeChangesOutput, error)
	SummarizeReleaseNote(ctx context.Context, input SummarizeReleaseNoteInput) (SummarizeReleaseNoteOutput, error)
}

type SummarizeRelativeChangesInput struct {
	MessageContext []domain.Message
	MaxOutputToken int64
}
type SummarizeRelativeChangesOutput struct {
	Messages []domain.Message
}

type SummarizeReleaseNoteInput struct {
	MessageContext []domain.Message
	MaxOutputToken int64
}
type SummarizeReleaseNoteOutput struct {
	Messages []domain.Message
}
