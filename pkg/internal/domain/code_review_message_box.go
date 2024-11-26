package domain

import (
	"github.com/pkg/errors"
	"gitlab-mr-reviewer/pkg/utils"
)

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"

	ModelGPT4O     = "gpt-4o"
	ModelGPT4OMini = "gpt-4o-mini"
)

type CodeReviewMessagebox struct {
	Message        []Message
	MaxInputToken  int64
	MaxOutputToken int64
}

type Message struct {
	Role    string
	Content string
}

func NewCodeReviewMessageBox(systemMessage string) *CodeReviewMessagebox {
	return &CodeReviewMessagebox{
		MaxInputToken:  10000,
		MaxOutputToken: 10000,
		Message: []Message{
			newSystemMessage(systemMessage),
		},
	}
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
	countTokens, err := utils.CountTokensWithGPT4OMini(m.Content)
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
