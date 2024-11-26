package cfg

import (
	"gitlab-mr-reviewer/pkg/cli"
	"gitlab-mr-reviewer/pkg/internal/handler"
	"gitlab-mr-reviewer/pkg/internal/repository"
	"gitlab-mr-reviewer/pkg/internal/usecase"
	"gitlab-mr-reviewer/pkg/logging"
)

type CliDependenciesInjector struct {
	MergeRequestCommand cli.Command
}

func NewCliDependenciesInjector(cfg *Config, logger *logging.ZaprLogger) (*CliDependenciesInjector, error) {
	gitlabRepository := repository.NewGitlabRepository(logger, cfg.Gitlab.Url, cfg.Gitlab.Token)
	openaiRepository := repository.NewOpenaiRepository(logger, cfg.OpenAI.Token)

	mergeRequestReviewer, err := usecase.NewGitlabMergeRequestReviewer(logger, cfg.OpenAI.SystemMessage, cfg.Gitlab.PathFilters, gitlabRepository, openaiRepository)
	if err != nil {
		return nil, err
	}

	mergeRequestHandler := handler.NewMergeRequestHandler(logger, mergeRequestReviewer)

	mergeRequestCommand := cli.NewMergeRequestCommand(
		cfg.Gitlab.ProjectId, cfg.Gitlab.MergeRequestId,
		cfg.OpenAI.Model, cfg.OpenAI.MaxInputToken, cfg.OpenAI.MaxOutputToken,
		logger,
		mergeRequestHandler,
	)

	return &CliDependenciesInjector{
		MergeRequestCommand: mergeRequestCommand,
	}, nil
}
