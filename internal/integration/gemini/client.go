package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/solidbit/integritypos/internal/config"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
	model      string
}

func NewClient(cfg config.AIConfig) *Client {
	model := cfg.Model
	if model == "" {
		model = "gemini-pro"
	}
	return &Client{
		httpClient: &http.Client{},
		apiKey:     cfg.APIKey,
		model:      model,
	}
}

func (c *Client) GenerateJSON(ctx context.Context, systemPrompt, userPrompt string) (json.RawMessage, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.model, c.apiKey)

	payload := map[string]interface{}{
		"system_instruction": map[string]interface{}{
			"parts": []map[string]string{{"text": systemPrompt}},
		},
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{{"text": userPrompt}},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":        0,
			"response_mime_type": "application/json",
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call gemini api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gemini api returned status: %d", resp.StatusCode)
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode gemini response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from gemini")
	}

	text := result.Candidates[0].Content.Parts[0].Text
	return json.RawMessage(text), nil
}
