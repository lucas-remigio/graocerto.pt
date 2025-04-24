package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lucas-remigio/wallet-tracker/config"
	"github.com/lucas-remigio/wallet-tracker/types"
)

type Client struct {
	apiKey string
}

func NewClient() *Client {
	apiKey := config.Envs.OpenAIKey
	if apiKey == "" {
		panic("OPENAI_API_KEY environment variable is not set")
	}
	return &Client{apiKey: apiKey}
}

func (c *Client) GenerateGPT4Response(prompt string) (string, error) {

	// Create the request payload
	request := types.GPTRequest{
		Model: "gpt-4", // Adjust to "gpt-4-turbo" if that's the specific version you want
		Messages: []types.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   1000,
		Temperature: 0.0,
	}

	// Convert the request payload to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Make the HTTP request
	url := "https://api.openai.com/v1/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		// Read the response body for debugging
		body := new(bytes.Buffer)
		body.ReadFrom(resp.Body)
		return "", fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, body.String())
	}

	// Parse the response
	var gptResponse types.GPTResponse
	err = json.NewDecoder(resp.Body).Decode(&gptResponse)
	if err != nil {
		return "", fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Extract and return the response text
	if len(gptResponse.Choices) == 0 {
		return "", fmt.Errorf("no choices returned in OpenAI response")
	}

	message, err := c.cleanAiMessage(gptResponse.Choices[0].Message.Content)
	if err != nil {
		return "", fmt.Errorf("failed to clean AI message: %w", err)
	}

	return message, nil
}

func (c *Client) cleanAiMessage(message string) (string, error) {
	// This function receives the message from OpenAi, and cleans everything before the { and after the }
	// so that only the json format is returned

	// Find the first occurrence of '{'
	start := bytes.Index([]byte(message), []byte("{"))
	if start == -1 {
		return "", fmt.Errorf("no JSON object found in the message")
	}

	// Find the last occurrence of '}'
	end := bytes.LastIndex([]byte(message), []byte("}"))
	if end == -1 {
		return "", fmt.Errorf("no JSON object found in the message")
	}

	// Extract the JSON object
	jsonObject := message[start : end+1]

	return string(jsonObject), nil

}
