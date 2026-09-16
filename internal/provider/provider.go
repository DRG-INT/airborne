package provider

import (
	"context"
	"io"
)

type ModelProvider interface {
	Name() string
	Chat(ctx context.Context, model, systemPrompt, userPrompt string, stream bool) (<-chan StreamEvent, error)
}

type StreamEvent struct {
	Delta string
	Done  bool
	Error error
	Usage *Usage
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type Request struct {
	Model    string
	Messages []Message
	Stream   bool
}

type Message struct {
	Role    string
	Content string
}

type ResponseWriter interface {
	Write(b []byte) (int, error)
	WriteString(s string) (int, error)
	Close() error
}

type OpenAIRequest struct {
	Model    string     `json:"model"`
	Messages []OMessage `json:"messages"`
	Stream   bool       `json:"stream"`
}

type OMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *Usage `json:"usage"`
}

type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Index        int      `json:"index"`
		FinishReason string   `json:"finish_reason"`
		Message      OMessage `json:"message"`
	} `json:"choices"`
	Usage Usage `json:"usage"`
}

func NewReader(r io.Reader) io.Reader {
	return r
}
