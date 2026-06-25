package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/solidbit/integritypos/internal/config"
)

type Client struct {
	httpClient    *http.Client
	baseURL       string
	token         string
	phoneNumberID string
}

func NewClient(cfg config.WhatsAppConfig) *Client {
	return &Client{
		httpClient:    &http.Client{},
		baseURL:       "https://graph.facebook.com/v18.0",
		token:         cfg.Token,
		phoneNumberID: cfg.PhoneNumberID,
	}
}

func (c *Client) GetPhoneNumberID() string {
	return c.phoneNumberID
}

func (c *Client) SendTextMessage(ctx context.Context, to, text string) error {
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text": map[string]string{
			"body": text,
		},
	}
	return c.sendRequest(ctx, payload)
}

type InteractiveOption struct {
	ID          string
	Title       string
	Description string
}

func (c *Client) SendInteractiveList(ctx context.Context, to, header, body, footer string, options []InteractiveOption) error {
	var rows []map[string]string
	for _, opt := range options {
		rows = append(rows, map[string]string{
			"id":          opt.ID,
			"title":       opt.Title,
			"description": opt.Description,
		})
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "interactive",
		"interactive": map[string]interface{}{
			"type": "list",
			"header": map[string]string{
				"type": "text",
				"text": header,
			},
			"body": map[string]string{
				"text": body,
			},
			"footer": map[string]string{
				"text": footer,
			},
			"action": map[string]interface{}{
				"button": "Seleccionar",
				"sections": []map[string]interface{}{
					{
						"title": "Opciones",
						"rows":  rows,
					},
				},
			},
		},
	}
	return c.sendRequest(ctx, payload)
}

func (c *Client) sendRequest(ctx context.Context, payload interface{}) error {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/%s/messages", c.baseURL, c.phoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("whatsapp api returned status: %d", resp.StatusCode)
	}

	return nil
}
