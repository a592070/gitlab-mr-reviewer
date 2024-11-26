package cli

import (
	"context"
	"gitlab-mr-reviewer/pkg/internal/handler"
	"gitlab-mr-reviewer/pkg/internal/usecase"
	"gitlab-mr-reviewer/pkg/logging"
	"time"
)

type Command interface {
	Run() error
}

type MergeRequestCommand struct {
	projectId           int32
	mergeRequestId      int32
	model               string
	maxInputToken       int64
	maxOutputToken      int64
	logger              *logging.ZaprLogger
	mergeRequestHandler handler.MergeRequestHandler
}

func NewMergeRequestCommand(
	projectId int32,
	mergeRequestId int32,
	model string,
	maxInputToken int64,
	maxOutputToken int64,
	logger *logging.ZaprLogger,
	mergeRequestHandler handler.MergeRequestHandler) Command {
	return &MergeRequestCommand{
		projectId:      projectId,
		mergeRequestId: mergeRequestId,
		model:          model,
		maxInputToken:  maxInputToken,
		maxOutputToken: maxOutputToken,

		logger:              logger,
		mergeRequestHandler: mergeRequestHandler,
	}
}

func (c *MergeRequestCommand) Run() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelFunc()
	err := c.mergeRequestHandler.Review(ctx, &usecase.MergeRequestReviewInput{
		ProjectId:      c.projectId,
		MergeRequestId: c.mergeRequestId,
		Model:          c.model,
		MaxInputToken:  c.maxInputToken,
		MaxOutputToken: c.maxOutputToken,
	})
	if err != nil {
		return err
	}
	c.logger.Info("Finished.")

	return nil
}
