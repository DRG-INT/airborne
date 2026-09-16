package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"airborne/internal/exit"
)

type OpenAI struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewOpenAI(apiKey, baseURL string) *OpenAI {
	return &OpenAI{
		apiKey:     apiKey,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

func (o *OpenAI) Name() string {
	return "openai-responses"
}

func (o *OpenAI) Chat(ctx context.Context, model, systemPrompt, userPrompt string, stream bool) (<-chan StreamEvent, error) {
	if o.apiKey == "" {
		return nil, fmt.Errorf("%s: OPENAI_API_KEY not set", exit.ProviderFailure)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	messages := []OMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	endpoint := o.baseURL + "/chat/completions"
	if strings.Contains(o.baseURL, "openai.com") || strings.Contains(endpoint, "/v1") {
		endpoint = o.baseURL + "/chat/completions"
	}

	reqBody := OpenAIRequest{
		Model:    model,
		Messages: messages,
		Stream:   stream,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	ch := make(chan StreamEvent, 100)

	if !stream {
		go func() {
			defer close(ch)
			resp, err := o.httpClient.Do(req)
			if err != nil {
				ch <- StreamEvent{Error: fmt.Errorf("provider request failed: %w", err)}
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				respBody, _ := io.ReadAll(resp.Body)
				ch <- StreamEvent{Error: fmt.Errorf("provider returned status %d: %s", resp.StatusCode, string(respBody))}
				return
			}

			var openaiResp OpenAIResponse
			if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
				ch <- StreamEvent{Error: fmt.Errorf("decode response: %w", err)}
				return
			}

			for _, choice := range openaiResp.Choices {
				ch <- StreamEvent{Delta: choice.Message.Content}
			}
			ch <- StreamEvent{Done: true, Usage: &openaiResp.Usage}
		}()
		return ch, nil
	}

	go func() {
		defer close(ch)
		resp, err := o.httpClient.Do(req)
		if err != nil {
			ch <- StreamEvent{Error: fmt.Errorf("provider request failed: %w", err)}
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			respBody, _ := io.ReadAll(resp.Body)
			ch <- StreamEvent{Error: fmt.Errorf("provider returned status %d: %s", resp.StatusCode, string(respBody))}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(string(scanner.Bytes()))
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "data: ") {
				line = strings.TrimPrefix(line, "data: ")
			}
			if strings.HasPrefix(line, "data:") {
				line = strings.TrimPrefix(line, "data:")
				line = strings.TrimSpace(line)
			}
			if line == "[DONE]" {
				ch <- StreamEvent{Done: true}
				return
			}

			var sse OpenAIStreamResponse
			if err := json.Unmarshal([]byte(line), &sse); err != nil {
				continue
			}
			for _, choice := range sse.Choices {
				if choice.Delta.Content != "" {
					ch <- StreamEvent{Delta: choice.Delta.Content}
				}
				if choice.FinishReason != "" {
					if sse.Usage != nil {
						ch <- StreamEvent{Done: true, Usage: sse.Usage}
					} else {
						ch <- StreamEvent{Done: true}
					}
					return
				}
			}
		}
	}()

	return ch, nil
}
