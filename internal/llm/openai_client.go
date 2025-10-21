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

type OpenAIClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewOpenAIClient(cfg *config.Config) *OpenAIClient {
	return &OpenAIClient{
		apiKey:  cfg.OpenAIApiKey,
		baseURL: cfg.OpenAIBaseURL,
		model:   cfg.OpenAIModel,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// GetEmbedding for strategy configs
func (c *OpenAIClient) GetEmbedding(text string) ([]float64, error) {
	reqBody := map[string]interface{}{
		"model": c.model,
		"input": text,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", c.baseURL+"/embeddings", bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Bearer "+c.apiKey)
	req.Header.Add("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI failed with status %d", resp.StatusCode)
	}
	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("No embedding returned")
	}
	return result.Data[0].Embedding, nil
}
