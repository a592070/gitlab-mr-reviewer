package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"gitlab-mr-reviewer/pkg/cfg"
	"gitlab-mr-reviewer/pkg/logging"
	"log"
	"os"
)

func main() {
	command := cfg.NewCommand()
	config, err := cfg.NewCliConfig(command)
	if err != nil {
		handleError(command, err)
	}
	logger, err := logging.NewZaprLogger(config.IsReleaseMode, config.LogLevel)
	if err != nil {
		handleError(command, err)
	}

	injector, err := cfg.NewCliDependenciesInjector(config, logger)
	if err != nil {
		handleError(command, err)
	}

	if err := injector.MergeRequestCommand.Run(); err != nil {
		handleError(command, err)
	}
}

// handleError log error message to std and exit
func handleError(command *cobra.Command, err error) {
	log.Output(2, fmt.Sprintf("[ERROR] %s", err.Error()))
	command.Help()
	os.Exit(0)
}
