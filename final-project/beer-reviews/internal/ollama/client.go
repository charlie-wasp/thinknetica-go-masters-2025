package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Think  bool   `json:"think"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

type Client struct {
	url *url.URL
}

func NewClient(targetURL string) (*Client, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse target URL: %v", err)
	}
	target.Path = "/api/generate"
	return &Client{url: target}, nil
}

func (c *Client) Query(model string, prompt string) (string, error) {
	requestBody := ollamaRequest{
		Model:  model,
		Prompt: prompt,
		Think:  false,
		Stream: false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.url.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("Error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 1 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama responded with code %d", resp.StatusCode)
	}

	var resultBody ollamaResponse
	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&resultBody)
	if err != nil {
		return "", fmt.Errorf("Error parsing llm response: %w", err)
	}

	return resultBody.Response, nil
}
