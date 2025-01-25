package types

type OpenAI interface {
	GenerateGPT4Response(prompt string) (string, error)
}

// GPTRequest represents the structure of the request to OpenAI's GPT-4 API.
type GPTRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

// Message represents a single message in a conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GPTResponse represents the structure of the response from OpenAI's API.
type GPTResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}
