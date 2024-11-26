package domain

import (
	"github.com/pkg/errors"
	"gitlab-mr-reviewer/pkg/utils"
	"slices"
)

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"

	ModelGPT4O     = "gpt-4o"
	ModelGPT4OMini = "gpt-4o-mini"
)

var models = []string{ModelGPT4O, ModelGPT4OMini}

type CodeReviewMessagebox struct {
	Message        []Message
	MaxInputToken  int64
	MaxOutputToken int64
	Model          string
}

type Message struct {
	Role    string
	Content string
}

func NewCodeReviewMessageBox(systemMessage string, model string, maxInputToken int64, maxOutputToken int64) (*CodeReviewMessagebox, error) {
	if maxInputToken <= 0 {
		maxInputToken = 10000
	}
	if maxOutputToken <= 0 {
		maxOutputToken = 10000
	}
	if !slices.Contains(models, model) {
		return nil, errors.Errorf("Not allow model %s, only accept models %s", model, models)
	}

	return &CodeReviewMessagebox{
		MaxInputToken:  maxInputToken,
		MaxOutputToken: maxOutputToken,
		Model:          model,
		Message: []Message{
			newSystemMessage(systemMessage),
		},
	}, nil
}

func NewUserMessage(content string) Message {
	return Message{
		Role:    RoleUser,
		Content: content,
	}
}
func NewAssistantMessage(content string) Message {
	return Message{
		Role:    RoleAssistant,
		Content: content,
	}
}
func newSystemMessage(content string) Message {
	return Message{
		Role:    RoleSystem,
		Content: content,
	}
}

func (c *CodeReviewMessagebox) AppendMessage(messages []Message) {
	c.Message = append(c.Message, messages...)
}

func (c *CodeReviewMessagebox) AddUserMessage(content string) error {
	m := NewUserMessage(content)
	countTokens, err := utils.CountTokens(c.Model, m.Content)
	if err != nil {
		return errors.Wrap(err, "Unable to count tokens")
	}
	if countTokens > c.MaxInputToken {
		return errors.Errorf("Token count (%d) exceeds maximum allowed (%d)", countTokens, c.MaxInputToken)
	}

	c.Message = append(c.Message, m)

	return nil
}

func (c *CodeReviewMessagebox) AddAssistantMessage(content string) {
	m := NewAssistantMessage(content)
	c.Message = append(c.Message, m)
}

func (c *CodeReviewMessagebox) GetLastAssistantMessage() (Message, error) {
	for i := len(c.Message) - 1; i >= 0; i-- {
		if c.Message[i].Role == RoleAssistant {
			return c.Message[i], nil
		}
	}
	return Message{}, errors.New("No assistant message found")
}
