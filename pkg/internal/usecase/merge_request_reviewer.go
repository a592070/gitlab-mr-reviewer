package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"gitlab-mr-reviewer/pkg/internal/domain"
	"gitlab-mr-reviewer/pkg/internal/repository"
	"gitlab-mr-reviewer/pkg/logging"
	"regexp"
	"text/template"
)

var (
	ErrorIgnoreCodeReview = errors.New("Ignore code review.")
)

type MergeRequestReviewer interface {
	Apply(context.Context, *MergeRequestReviewInput) (*MergeRequestReviewOutput, error)
}

type MergeRequestReviewInput struct {
	ProjectId      int32 `json:"project_id,omitempty" validate:"required,gt=0"`
	MergeRequestId int32 `json:"merge_request_id,omitempty" validate:"required,gt=0"`
}
type MergeRequestReviewOutput struct {
	SummarizeRelativeChanges string `json:"summarize_relative_changes"`
	SummarizeReleaseNote     string `json:"summarize_release_note"`
}

type gitlabMergeRequestReviewer struct {
	logger           *logging.ZaprLogger
	gitlabRepository repository.GitlabRepository
	openaiRepository repository.LLMRepository
	systemMessage    string
	pathFilters      []*regexp.Regexp
}

func NewGitlabMergeRequestReviewer(
	logger *logging.ZaprLogger,
	systemMessage string,
	pathFilters []string,
	gitlabRepository repository.GitlabRepository,
	llmRepository repository.LLMRepository) (MergeRequestReviewer, error) {

	filters := make([]*regexp.Regexp, len(pathFilters))
	for i, f := range pathFilters {
		regex, err := regexp.Compile(f)
		if err != nil {
			return nil, errors.Wrap(err, "[NewGitlabMergeRequestReviewer]failed to compile path filter")
		}
		filters[i] = regex
	}

	return &gitlabMergeRequestReviewer{
		logger:           logger,
		gitlabRepository: gitlabRepository,
		openaiRepository: llmRepository,
		systemMessage:    systemMessage,
		pathFilters:      filters,
	}, nil
}

func (r *gitlabMergeRequestReviewer) Apply(ctx context.Context, input *MergeRequestReviewInput) (*MergeRequestReviewOutput, error) {
	mergeRequest, err := r.getMergeRequest(ctx, input.ProjectId, input.MergeRequestId)
	if err != nil {
		return nil, err
	}

	mergeRequest.FilterIgnorePaths(r.pathFilters)
	if mergeRequest.IgnoreReview() {
		return nil, ErrorIgnoreCodeReview
	}

	codeReviewMessageBox := domain.NewCodeReviewMessageBox(r.systemMessage)

	summarizeRelativeChanges, err := r.summarizeRelativeChanges(ctx, mergeRequest, codeReviewMessageBox)
	if err != nil {
		return nil, err
	}
	mergeRequest.RelativeChangeNote = domain.Note(summarizeRelativeChanges)

	summarizeReleaseNote, err := r.summarizeReleaseNote(ctx, codeReviewMessageBox)
	if err != nil {
		return nil, err
	}
	mergeRequest.SummaryNote = domain.Note(summarizeReleaseNote)

	if err := r.gitlabRepository.CreateMergeRequestSummary(ctx, toCreateMergeRequestSummaryInput(mergeRequest)); err != nil {
		return nil, err
	}

	return &MergeRequestReviewOutput{
		SummarizeRelativeChanges: summarizeRelativeChanges,
		SummarizeReleaseNote:     summarizeReleaseNote,
	}, nil
}

func (r *gitlabMergeRequestReviewer) getMergeRequest(ctx context.Context, projectId int32, mergeRequestId int32) (*domain.MergeRequest, error) {
	mergeRequestDto, err := r.gitlabRepository.GetMergeRequest(ctx, projectId, mergeRequestId)
	if err != nil {
		return nil, err
	}
	diffsDto, err := r.gitlabRepository.ListDiffByMergeRequestId(ctx, projectId, mergeRequestId)
	if err != nil {
		return nil, err
	}

	return toMergeRequestDomain(mergeRequestDto, diffsDto), nil
}

func (r *gitlabMergeRequestReviewer) summarizeRelativeChanges(ctx context.Context, mergeRequest *domain.MergeRequest, codeReviewMessageBox *domain.CodeReviewMessagebox) (string, error) {
	relativeChangesPrompt, err := r.generateRelativeChangesPrompt(mergeRequest)
	if err != nil {
		return "", errors.Wrap(err, "Failed to generate relative changes prompt")
	}

	if err := codeReviewMessageBox.AddUserMessage(relativeChangesPrompt); err != nil {
		return "", errors.Wrap(err, "Failed to add relative changes prompt to user message")
	}

	codeReviewData, err := r.openaiRepository.SummarizeRelativeChanges(ctx, repository.SummarizeRelativeChangesInput{
		MessageContext: codeReviewMessageBox.Message,
		MaxOutputToken: codeReviewMessageBox.MaxOutputToken,
	})
	if err != nil {
		return "", errors.Wrap(err, "Failed to create relative changes completion")
	}
	codeReviewMessageBox.AppendMessage(codeReviewData.Messages)

	lastAssistantMessage, err := codeReviewMessageBox.GetLastAssistantMessage()
	if err != nil {
		return "", errors.Wrap(err, "Failed to get last assistant message")
	}
	return lastAssistantMessage.Content, nil
}
func (r *gitlabMergeRequestReviewer) summarizeReleaseNote(ctx context.Context, codeReviewMessageBox *domain.CodeReviewMessagebox) (string, error) {
	releaseNotePrompt := r.generateReleaseNotePrompt()

	if err := codeReviewMessageBox.AddUserMessage(releaseNotePrompt); err != nil {
		return "", errors.Wrap(err, "Failed to add release note prompt to user message")
	}

	codeReviewData, err := r.openaiRepository.SummarizeReleaseNote(ctx, repository.SummarizeReleaseNoteInput{
		MessageContext: codeReviewMessageBox.Message,
		MaxOutputToken: codeReviewMessageBox.MaxOutputToken,
	})
	if err != nil {
		return "", errors.Wrap(err, "Failed to create release note completion")
	}
	codeReviewMessageBox.AppendMessage(codeReviewData.Messages)

	lastAssistantMessage, err := codeReviewMessageBox.GetLastAssistantMessage()
	if err != nil {
		return "", errors.Wrap(err, "Failed to get last assistant message")
	}
	return lastAssistantMessage.Content, nil
}

func (r *gitlabMergeRequestReviewer) generateRelativeChangesPrompt(mr *domain.MergeRequest) (string, error) {
	promptTpl := "Provide your final response in the `markdown` format with the following content:\n" +
		"- Summary (comment on the overall change instead of specific files within 80 words)\n" +
		"- Table of files and their summaries. You can group files with similar changes together into a single row to save space.\n\n" +
		"Avoid additional commentary as this summary will be added as a comment on the GitHub pull request.\n\n" +
		"## Merge Request Title\n" +
		"`{{.Title}}`\n\n" +
		"## Description\n" +
		"```\n" +
		"{{.Description}}\n" +
		"```\n\n" +
		"## Diff\n" +
		"```\n" +
		"{{.FileDiff}}\n" +
		"```"

	marshalDifference, err := json.Marshal(mr.RelativeChanges)
	if err != nil {
		return "", err
	}
	return fillUpTemplate(promptTpl, map[string]string{
		"Title":       mr.Title,
		"Description": mr.Description,
		"FileDiff":    string(marshalDifference),
	})
}

func (r *gitlabMergeRequestReviewer) generateReleaseNotePrompt() string {
	prompt := "Create concise release notes in `markdown` format for this pull request, focusing on its purpose and user story. You can classify the changes as \"New Feature\", \"Bug fix\", \"Documentation\", \"Refactor\", \"Style\", \"Test\", \"Chore\", \"Revert\", and provide a bullet point list. For example: \"New Feature: An integrations page was added to the UI\". Keep your response within 50-100 words. Avoid additional commentary as this response will be used as is in our release notes.\n\n" +
		"Below the release notes, generate a short, celebratory poem about the changes in this PR and add this poem as a quote (> symbol). You can use emojis in the poem, where they are relevant."
	return prompt
}

func fillUpTemplate(tpl string, data interface{}) (string, error) {
	t := template.New("tpl")
	parse, err := t.Parse(tpl)
	if err != nil {
		return "", err
	}
	var wr bytes.Buffer

	if err := parse.Execute(&wr, data); err != nil {
		return "", err
	}
	return wr.String(), nil
}
