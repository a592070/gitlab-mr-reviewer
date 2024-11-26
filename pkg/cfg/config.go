package cfg

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
)

const (
	name                 = "gitlab-mr-reviewer"
	defaultConfigFileDir = "config/config.yaml"
)

type Config struct {
	LogLevel      string `validate:"required,oneof=debug info warn error"`
	IsReleaseMode bool
	Gitlab        struct {
		Url            string `validate:"required"`
		Token          string `validate:"required"`
		ProjectId      int32  `validate:"required_with_all=Gitlab.ProjectId Gitlab.MergeRequestId,omitempty,gte=1"`
		MergeRequestId int32  `validate:"required_with_all=Gitlab.ProjectId Gitlab.MergeRequestId,omitempty,gte=1"`
		PathFilters    []string
	}
	OpenAI struct {
		Token          string `validate:"required"`
		SystemMessage  string `validate:"required"`
		Model          string `validate:"required"`
		MaxInputToken  int64  `validate:"required"`
		MaxOutputToken int64  `validate:"required"`
	}
}

func NewCommand() *cobra.Command {
	var rootCmd = &cobra.Command{
		Short: fmt.Sprintf("%s is able to generate merge-request review summaries using OpenAI", name),
		Long:  fmt.Sprintf("%s is able to generate merge-request review summaries using OpenAI", name),
		Use:   name,
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
			for i := range args {
				log.Printf("Running with arg: %s", args[i])
			}
		},
	}
	helpFunc := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		helpFunc(cmd, args)
		os.Exit(0)
	})

	return rootCmd
}

func NewCliConfig(rootCmd *cobra.Command) (*Config, error) {
	v := viper.New()
	var configFilePath string
	rootCmd.PersistentFlags().StringVar(&configFilePath, "config", defaultConfigFileDir, "Config file directory")
	rootCmd.PersistentFlags().Int32("project", 0, "Gitlab Project ID, or use GITLAB_PROJECTID environment variable.")
	rootCmd.PersistentFlags().Int32("merge-request", 0, "Gitlab MergeRequest ID, or use GITLAB_MERGEREQUESTID environment variable.")
	rootCmd.PersistentFlags().String("gitlab-url", "", "Gitlab URL, or use GITLAB_URL environment variable.")
	rootCmd.PersistentFlags().String("gitlab-token", "", "Gitlab authorization token, or use GITLAB_TOKEN environment variable.")
	rootCmd.PersistentFlags().String("openai-token", "", "OpenAI authorization token, or use OPENAI_TOKEN environment variable.")
	rootCmd.PersistentFlags().String("log", "info", "Log level, or use LOGLEVEL environment variable.")

	err := rootCmd.Execute()
	if err != nil {
		return nil, errors.Wrap(err, "[NewCliConfig]failed to execute cmd")
	}

	v.SetConfigFile(configFilePath)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.BindPFlag("Gitlab.ProjectId", rootCmd.PersistentFlags().Lookup("project")); err != nil {
		return nil, errors.Wrap(err, "[NewCliConfig]failed to bind flag")
	}
	if err := v.BindPFlag("Gitlab.MergeRequestId", rootCmd.PersistentFlags().Lookup("merge-request")); err != nil {
		return nil, errors.Wrap(err, "[NewCliConfig]failed to bind flag")
	}
	if err := v.BindPFlag("Gitlab.URL", rootCmd.PersistentFlags().Lookup("gitlab-url")); err != nil {
		return nil, errors.Wrap(err, "[NewCliConfig]failed to bind flag")
	}
	if err := v.BindPFlag("Gitlab.Token", rootCmd.PersistentFlags().Lookup("gitlab-token")); err != nil {
		return nil, errors.Wrap(err, "[NewCliConfig]failed to bind flag")
	}
	if err := v.BindPFlag("OpenAI.Token", rootCmd.PersistentFlags().Lookup("openai-token")); err != nil {
		return nil, errors.Wrap(err, "[NewCliConfig]failed to bind flag")
	}
	if err := v.BindPFlag("LogLevel", rootCmd.PersistentFlags().Lookup("log")); err != nil {
		return nil, errors.Wrap(err, "[NewCliConfig]failed to bind flag")
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "[NewCliConfig]failed to read config")
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, errors.Wrap(err, "[NewCliConfig]failed to unmarshal config")
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(config); err != nil {
		return nil, errors.Wrap(err, "[NewCliConfig]failed to validate config")
	}

	return &config, nil
}
