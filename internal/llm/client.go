/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ares_api/internal/config"
)

type OllamaClient struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewOllamaClient(cfg *config.Config) *OllamaClient {
	return &OllamaClient{
		baseURL: cfg.OllamaBaseURL,
		model:   cfg.OllamaModel,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// GenerateCompletion for strategy reasoning
func (c *OllamaClient) GenerateCompletion(prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":  c.model,
		"prompt": prompt,
	}
	body, _ := json.Marshal(reqBody)
	resp, err := c.client.Post(c.baseURL+"/generate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama failed with status %d", resp.StatusCode)
	}
	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Response, nil
}
