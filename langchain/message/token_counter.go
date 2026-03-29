package message

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
	"github.com/tmc/langchaingo/llms"
)

type TokenCounter struct {
	tokenizer *tiktoken.Tiktoken
}

func NewTokenCounter() *TokenCounter {
	tkm, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return &TokenCounter{tokenizer: nil}
	}

	return &TokenCounter{tokenizer: tkm}
}

func (tc *TokenCounter) CalculateTokens(msg llms.MessageContent) (int, error) {
	if tc.tokenizer == nil {
		text := tc.MessageToText(msg)
		return len(text) / 4, nil
	}

	text := tc.MessageToText(msg)
	tokens := tc.tokenizer.Encode(text, nil, nil)
	return len(tokens), nil
}

func (tc *TokenCounter) MessageToText(msg llms.MessageContent) string {
	var text string

	switch msg.Role {
	case llms.ChatMessageTypeSystem:
		text += "System: "
	case llms.ChatMessageTypeHuman:
		text += "Human: "
	case llms.ChatMessageTypeAI:
		text += "AI: "
	case llms.ChatMessageTypeTool:
		text += "Tool: "
	}

	for _, part := range msg.Parts {
		switch p := part.(type) {
		case llms.TextContent:
			text += p.Text
		case llms.ToolCall:
			if p.FunctionCall != nil {
				text += fmt.Sprintf("ToolCall[name=%s, args=%s] ",
					p.FunctionCall.Name, p.FunctionCall.Arguments)
			}
		case llms.ToolCallResponse:
			text += fmt.Sprintf("ToolResponse[name=%s, content=%s] ",
				p.Name, p.Content)
		}
	}

	return text
}
