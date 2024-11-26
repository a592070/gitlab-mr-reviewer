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
	logger              *logging.ZaprLogger
	mergeRequestHandler handler.MergeRequestHandler
}

func NewMergeRequestCommand(
	projectId int32,
	mergeRequestId int32,
	logger *logging.ZaprLogger,
	mergeRequestHandler handler.MergeRequestHandler) Command {
	return &MergeRequestCommand{
		projectId:           projectId,
		mergeRequestId:      mergeRequestId,
		logger:              logger,
		mergeRequestHandler: mergeRequestHandler,
	}
}

func (e *MergeRequestCommand) Run() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelFunc()
	err := e.mergeRequestHandler.Review(ctx, &usecase.MergeRequestReviewInput{
		ProjectId:      e.projectId,
		MergeRequestId: e.mergeRequestId,
	})
	if err != nil {
		return err
	}
	e.logger.Info("Finished.")

	return nil
}
