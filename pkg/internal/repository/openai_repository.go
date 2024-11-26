package repository

import (
	"context"
	"fmt"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/pkg/errors"
	"gitlab-mr-reviewer/pkg/internal/domain"
	"gitlab-mr-reviewer/pkg/logging"
	"time"
)

type openaiRepository struct {
	client *openai.Client
	logger *logging.ZaprLogger
}

func NewOpenaiRepository(logger *logging.ZaprLogger, apiKey string) LLMRepository {
	client := openai.NewClient(
		option.WithAPIKey(apiKey),                 // defaults to os.LookupEnv("OPENAI_API_KEY")
		option.WithRequestTimeout(60*time.Second), // set default timeout
	)

	return &openaiRepository{
		client: client,
		logger: logger,
	}
}

func (r *openaiRepository) SummarizeRelativeChanges(ctx context.Context, input SummarizeRelativeChangesInput) (SummarizeRelativeChangesOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, len(input.MessageContext))
	for i, message := range input.MessageContext {
		msg, err := toChatCompletionMessage(message)
		if err != nil {
			return SummarizeRelativeChangesOutput{}, err
		}
		openaiMessages[i] = msg
	}

	resp, err := r.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Messages:            openai.F(openaiMessages),
			Model:               openai.F(openai.ChatModelGPT4o),
			MaxCompletionTokens: openai.Int(input.MaxOutputToken),
		},
	)
	if err != nil {
		return SummarizeRelativeChangesOutput{}, err
	}

	output := SummarizeRelativeChangesOutput{
		Messages: make([]domain.Message, len(resp.Choices)),
	}

	if len(resp.Choices) == 0 {
		return output, errors.New("response has no choices")
	}

	r.logger.Info(fmt.Sprintf("Response TotalTokens: %d", resp.Usage.TotalTokens))
	r.logger.Info(fmt.Sprintf("Response PromptTokens: %d", resp.Usage.PromptTokens))
	r.logger.Info(fmt.Sprintf("Response CompletionTokens: %d", resp.Usage.CompletionTokens))
	for i, choice := range resp.Choices {
		r.logger.Info(fmt.Sprintf("Received response role: %s", choice.Message.Role))
		r.logger.Info(fmt.Sprintf("Received response choice: %s", choice.Message.Content))
		//r.logger.Info(fmt.Sprintf("Received response json: %#v", choice.Message.JSON))

		output.Messages[i] = toDomainMessage(choice.Message)
	}

	return output, nil
}

func (r *openaiRepository) SummarizeReleaseNote(ctx context.Context, input SummarizeReleaseNoteInput) (SummarizeReleaseNoteOutput, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	openaiMessages := make([]openai.ChatCompletionMessageParamUnion, len(input.MessageContext))
	for i, message := range input.MessageContext {
		msg, err := toChatCompletionMessage(message)
		if err != nil {
			return SummarizeReleaseNoteOutput{}, err
		}
		openaiMessages[i] = msg
	}

	resp, err := r.client.Chat.Completions.New(
		ctx,
		openai.ChatCompletionNewParams{
			Messages:            openai.F(openaiMessages),
			Model:               openai.F(openai.ChatModelGPT4o),
			MaxCompletionTokens: openai.Int(input.MaxOutputToken),
		},
	)
	if err != nil {
		return SummarizeReleaseNoteOutput{}, err
	}

	output := SummarizeReleaseNoteOutput{
		Messages: make([]domain.Message, len(resp.Choices)),
	}

	if len(resp.Choices) == 0 {
		return output, errors.New("response has no choices")
	}

	r.logger.Info(fmt.Sprintf("Response TotalTokens: %d", resp.Usage.TotalTokens))
	r.logger.Info(fmt.Sprintf("Response PromptTokens: %d", resp.Usage.PromptTokens))
	r.logger.Info(fmt.Sprintf("Response CompletionTokens: %d", resp.Usage.CompletionTokens))
	for i, choice := range resp.Choices {
		r.logger.Info(fmt.Sprintf("Received response role: %s", choice.Message.Role))
		r.logger.Info(fmt.Sprintf("Received response choice: %s", choice.Message.Content))
		//r.logger.Info(fmt.Sprintf("Received response json: %#v", choice.Message.JSON))

		output.Messages[i] = toDomainMessage(choice.Message)
	}

	return output, nil

}

func toChatCompletionMessage(message domain.Message) (openai.ChatCompletionMessageParamUnion, error) {
	switch message.Role {
	case domain.RoleSystem:
		return openai.SystemMessage(message.Content), nil
	case domain.RoleUser:
		return openai.UserMessage(message.Content), nil
	case domain.RoleAssistant:
		return openai.AssistantMessage(message.Content), nil
	default:
		return nil, errors.Errorf("Unsupported role %s", message.Role)
	}
}

func toDomainMessage(openaiMessage openai.ChatCompletionMessage) domain.Message {
	return domain.NewAssistantMessage(openaiMessage.Content)
}
