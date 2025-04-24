package determinevictim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

// Request payload for Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	System string `json:"system"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// Response payload from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// CallOllama takes a model and prompt, sends a request to the Ollama API, and returns the result
func CallOllama(model, systemPrompt string, prompt string) (string, error) {
	url := "http://localhost:11434/api/generate"

	// Construct the request payload
	requestData := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		System: systemPrompt,
		Stream: false,
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send the POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Ollama can stream responses line by line; for simplicity, parse as a single response here
	var response OllamaResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Err(err).Str("response", string(body)).Msg("failed to unmarshal response")
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Response, nil
}
