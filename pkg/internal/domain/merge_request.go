package domain

import (
	"regexp"
	"strings"
)

const (
	ignore = "@codeReview: ignore"
)

type MergeRequest struct {
	ID        int32
	ProjectID int32

	Title              string
	Description        string
	DifferentReference *DifferentReference
	RelativeChanges    []RelativeChange
	RelativeChangeNote Note
	SummaryNote        Note
}

func NewMergeRequest(
	ID int32, projectID int32,
	title string, description string,
	differentReference *DifferentReference,
	RelativeChanges []RelativeChange) *MergeRequest {
	return &MergeRequest{
		ID: ID, ProjectID: projectID,
		Title: title, Description: description,
		DifferentReference: differentReference, RelativeChanges: RelativeChanges}
}

type DifferentReference struct {
	BaseSha string
	HeadSha string
}

type RelativeChange struct {
	Diff        string
	NewPath     string
	OldPath     string
	NewFile     bool
	RenameFile  bool
	DeletedFile bool
}

type Note string

func (mr *MergeRequest) IgnoreReview() bool {
	return strings.Contains(mr.Description, ignore) || len(mr.RelativeChanges) <= 0
}

// FilterIgnorePaths filter the RelativeChanges which match the given []*regexp.Regexp , and return the filter count
func (mr *MergeRequest) FilterIgnorePaths(pathFilters []*regexp.Regexp) int32 {
	var remainChanges []RelativeChange
	var filterCount int32
	for _, change := range mr.RelativeChanges {
		remain := true
		for _, filter := range pathFilters {
			if filter.MatchString(change.NewPath) {
				remain = false
				filterCount++
				break
			}
		}

		if remain {
			remainChanges = append(remainChanges, change)
		}
	}
	mr.RelativeChanges = remainChanges
	return filterCount
}
