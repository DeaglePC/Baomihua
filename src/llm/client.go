package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"baomihua/config"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature float32   `json:"temperature"`
}

type ChatCompletionChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

// Result models the final JSON outcome expected from the LLM
type Result struct {
	Explanation string `json:"explanation"`
	Command     string `json:"command"`
}

// StreamCompletion sends the request to the LLM and streams the response back via a channel
func StreamCompletion(prompt string, ctx EnvContext, contentChan chan<- string, errChan chan<- error) {
	if GlobalRegistry == nil {
		InitRegistry()
		// Attempt to load without forcing refresh, ignore error if it fails to load some models
		_ = GlobalRegistry.LoadModels(false)
	}

	model := config.GetModel()

	actualModelName := model
	var vendorName string
	parts := strings.SplitN(model, "/", 2)
	if len(parts) == 2 {
		vendorName = parts[0]
		actualModelName = parts[1]
	}

	var provider Provider
	var err error

	if vendorName != "" {
		provider, err = GlobalRegistry.GetProviderByName(vendorName)
		if err != nil {
			provider, err = GlobalRegistry.GetProviderForModel(actualModelName)
		}
	} else {
		provider, err = GlobalRegistry.GetProviderForModel(actualModelName)
	}

	if err != nil {
		errChan <- err
		close(contentChan)
		close(errChan)
		return
	}

	provider.StreamCompletion(actualModelName, prompt, ctx, contentChan, errChan)
}

func (p *OpenAICompatibleProvider) StreamCompletion(model, prompt string, ctx EnvContext, contentChan chan<- string, errChan chan<- error) {
	defer close(contentChan)
	defer close(errChan)

	sysPrompt := BuildSystemPrompt(ctx)

	reqBody := ChatRequest{
		Model: model,
		Messages: []Message{
			{Role: "system", Content: sysPrompt},
			{Role: "user", Content: prompt},
		},
		Stream:      true,
		Temperature: 0.1,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		errChan <- fmt.Errorf("failed to marshal request: %w", err)
		return
	}

	url := strings.TrimRight(p.vendor.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		errChan <- fmt.Errorf("failed to create request: %w", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if p.vendor.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.vendor.APIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errChan <- fmt.Errorf("request failed: %w", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk ChatCompletionChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue // Skip malformed chunks instead of failing the stream
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				contentChan <- content
			}
		}
	}

	if err := scanner.Err(); err != nil {
		errChan <- fmt.Errorf("error reading stream: %w", err)
	}
}

// ParseResult parses the accumulated raw string from the stream into Result
func ParseResult(raw string) (*Result, error) {
	// Attempt to strip potential markdown codeblock formatting that models ignore prompt instructions for sometimes
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```json") {
		raw = strings.TrimPrefix(raw, "```json")
	} else if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```")
	}
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var res Result
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w (raw response: %s)", err, raw)
	}
	return &res, nil
}
