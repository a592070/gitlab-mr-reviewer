package handler

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"gitlab-mr-reviewer/pkg/internal/usecase"
	"gitlab-mr-reviewer/pkg/logging"
)

type MergeRequestHandler interface {
	Review(context.Context, *usecase.MergeRequestReviewInput) error
}

type mergeRequestHandler struct {
	logger               *logging.ZaprLogger
	mergeRequestReviewer usecase.MergeRequestReviewer
}

func NewMergeRequestHandler(logger *logging.ZaprLogger,
	mergeRequestReviewer usecase.MergeRequestReviewer,
) MergeRequestHandler {
	return &mergeRequestHandler{
		logger:               logger,
		mergeRequestReviewer: mergeRequestReviewer,
	}
}

func (h *mergeRequestHandler) Review(ctx context.Context, input *usecase.MergeRequestReviewInput) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(input)
	if err != nil {
		h.logger.Error(err, fmt.Sprintf("Failed to validate input: %#v", input))
		return errors.Wrap(err, "Failed to validate input.")
	}

	_, err = h.mergeRequestReviewer.Apply(ctx, input)
	if err != nil {
		if errors.Is(err, usecase.ErrorIgnoreCodeReview) {
			h.logger.Info("There is nothing to review")
			return nil
		}

		h.logger.Error(err, "Failed to apply input")
		return errors.Wrap(err, "Failed to apply input")
	}

	return nil
}
