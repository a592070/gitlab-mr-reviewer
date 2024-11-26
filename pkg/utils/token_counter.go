package utils

import (
	"github.com/pkoukk/tiktoken-go"
)

const (
	ModelGPT4OMini = "gpt-4o-mini"
)

func CountTokens(modelName, content string) (int64, error) {
	tkm, err := tiktoken.EncodingForModel(modelName)
	if err != nil {
		return 0, err
	}
	// encode
	token := tkm.Encode(content, nil, nil)

	return int64(len(token)), nil
}

func CountTokensWithGPT4OMini(content string) (int64, error) {
	return CountTokens(ModelGPT4OMini, content)
}
