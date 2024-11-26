package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gitlab-mr-reviewer/pkg/logging"
	"io"
	"net/http"
	"strings"
)

type GitlabRepository interface {
	ListDiffByMergeRequestId(ctx context.Context, projectId, mergeRequestId int32) ([]DiffDto, error)
	GetMergeRequest(ctx context.Context, projectId, mergeRequestId int32) (*MergeRequestDto, error)
	CreateMergeRequestSummary(context.Context, CreateMergeRequestSummaryInput) error
}

type CommitDto struct {
	Id      string `json:"id"`
	ShortId string `json:"short_id"`
	Title   string `json:"title"`
	Message string `json:"Message"`
}
type DiffDto struct {
	Diff        string `json:"diff"`
	NewPath     string `json:"new_path"`
	OldPath     string `json:"old_path"`
	NewFile     bool   `json:"new_file"`
	RenameFile  bool   `json:"renamed_file"`
	DeletedFile bool   `json:"deleted_file"`
}
type DiffRefsDto struct {
	BaseSha string `json:"base_sha"`
	HeadSha string `json:"head_sha"`
}
type MergeRequestDto struct {
	ProjectId   int32       `json:"project_id"`
	Id          int32       `json:"iid"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	DiffRefs    DiffRefsDto `json:"diff_refs"`
}

type CreateMergeRequestSummaryInput struct {
	ProjectId, MergeRequestId       int32
	RelativeChangeNote, SummaryNote string
}

type gitlabRepository struct {
	logger        *logging.ZaprLogger
	httpClient    *http.Client
	baseUrl       string
	authorization string
}

func NewGitlabRepository(logger *logging.ZaprLogger, baseUrl string, authorization string) GitlabRepository {
	return &gitlabRepository{
		httpClient:    &http.Client{},
		logger:        logger,
		baseUrl:       baseUrl,
		authorization: authorization,
	}
}

func (r *gitlabRepository) ListDiffByMergeRequestId(ctx context.Context, projectId, mergeRequestId int32) ([]DiffDto, error) {
	var diffs []DiffDto

	url := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/diffs", r.baseUrl, projectId, mergeRequestId)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("PRIVATE-TOKEN", r.authorization)

	response, err := r.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			r.logger.Info(fmt.Sprintf("Failed to read response body: %s", err))
		}
		r.logger.Info(fmt.Sprintf("response: %s", string(bodyBytes)))
		return nil, errors.New(fmt.Sprintf("response status code %d", response.StatusCode))
	}

	if err := json.NewDecoder(response.Body).Decode(&diffs); err != nil {
		return nil, err
	}

	return diffs, nil
}

func (r *gitlabRepository) GetMergeRequest(ctx context.Context, projectId, mergeRequestId int32) (*MergeRequestDto, error) {
	var mr MergeRequestDto

	url := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d", r.baseUrl, projectId, mergeRequestId)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		r.logger.Error(err, "Failed to create request")
		return nil, err
	}
	request.Header.Add("PRIVATE-TOKEN", r.authorization)

	response, err := r.httpClient.Do(request)
	if err != nil {
		r.logger.Error(err, "Failed to send request")
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			r.logger.Info(fmt.Sprintf("Failed to read response body: %s", err.Error()))
		}
		r.logger.Info(fmt.Sprintf("response: %s", string(bodyBytes)))
		return nil, errors.New(fmt.Sprintf("response status code %d", response.StatusCode))
	}

	if err := json.NewDecoder(response.Body).Decode(&mr); err != nil {
		r.logger.Error(err, "Failed to decode response")
		return nil, err
	}

	return &mr, nil
}

func (r *gitlabRepository) CreateMergeRequestSummary(ctx context.Context, input CreateMergeRequestSummaryInput) error {
	builder := strings.Builder{}
	builder.WriteString(":robot: CodeReviewerBot\n\n")
	builder.WriteString(input.RelativeChangeNote)
	builder.WriteString("\n---\n")
	builder.WriteString(input.SummaryNote)
	builder.WriteString("\n---\n")
	builder.WriteString("### Ignoring further reviews\n- Type `@codeReview: ignore` anywhere in the MR description to ignore further reviews from the bot.")

	note := builder.String()
	r.logger.Debug(fmt.Sprintf("generated note: %s", note))

	return r.createMergeRequestNote(ctx, input.ProjectId, input.MergeRequestId, note)
}

func (r *gitlabRepository) createMergeRequestNote(ctx context.Context, projectId, mergeRequestId int32, note string) error {
	url := fmt.Sprintf("%s/api/v4/projects/%d/merge_requests/%d/notes", r.baseUrl, projectId, mergeRequestId)

	requestBody := map[string]string{
		"body": note,
	}
	requestBodyByte, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(requestBodyByte))
	if err != nil {
		return err
	}
	request.Header.Add("PRIVATE-TOKEN", r.authorization)
	request.Header.Set("Content-Type", "application/json")

	response, err := r.httpClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusCreated {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			r.logger.Info(fmt.Sprintf("Failed to read response body: %s", err))
		}
		r.logger.Info(fmt.Sprintf("response: %s", string(bodyBytes)))
		return errors.New(fmt.Sprintf("Failed to create note. status: %d, response: %s", response.StatusCode, string(bodyBytes)))
	}
	return nil
}

func (r *gitlabRepository) ListCommitFromSource() {

}
