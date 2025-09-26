package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Client struct {
	BaseURL string
}

// NewClientFromEnv initializes a new Ollama client from environment variables
func NewClientFromEnv() *Client {
	baseURL := os.Getenv("OLLAMA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:11434/api" // default Ollama endpoint
	}
	return &Client{BaseURL: baseURL}
}

// Chat sends a message to the Ollama chat endpoint
func (c *Client) Chat(model, message string) (string, error) {
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": message},
		},
	}
	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(fmt.Sprintf("%s/chat", c.BaseURL), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama chat failed: %s", string(body))
	}

	// Ollama streams responses, so we’ll capture incrementally
	var output string
	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		var part struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Done bool `json:"done"`
		}
		if err := decoder.Decode(&part); err != nil {
			return "", err
		}
		output += part.Message.Content
		if part.Done {
			break
		}
	}
	return output, nil
}

// Generate sends a prompt to the Ollama generate endpoint
func (c *Client) Generate(model, prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
	}
	log.Printf("Generating with model=%s, prompt=%s", model, prompt)

	data, _ := json.Marshal(reqBody)

	resp, err := http.Post(fmt.Sprintf("%s/generate", c.BaseURL), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama generate failed: %s", string(body))
	}

	// Capture streaming output
	var output string
	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		var part struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		if err := decoder.Decode(&part); err != nil {
			return "", err
		}
		output += part.Response
		if part.Done {
			break
		}
	}
	return output, nil
}


