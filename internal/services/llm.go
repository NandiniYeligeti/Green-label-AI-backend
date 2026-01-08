package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Recipe struct {
	Title       string   `json:"title"`
	TimeMinutes int      `json:"time_minutes"`
	Ingredients []string `json:"ingredients"`
	Steps       []string `json:"steps"`
}

type openAIResponsesRequest struct {
	Model           string `json:"model"`
	Input           string `json:"input"`
	MaxOutputTokens int    `json:"max_output_tokens,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
}

type openAIResponsesResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func GenerateRecipesForProduct(productName string, count int) ([]Recipe, error) {
	apiKey := strings.TrimSpace(os.Getenv("LLM_API_KEY"))
	if apiKey == "" {
		return nil, errors.New("LLM_API_KEY is not set")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("LLM_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := strings.TrimSpace(os.Getenv("LLM_MODEL"))
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	if count <= 0 {
		count = 2
	}
	if count > 3 {
		count = 3
	}

	input := fmt.Sprintf(
		"You are a helpful cooking assistant. Respond with ONLY valid JSON. No markdown.\n\nGive me %d simple recipes using '%s' as a main ingredient. Output MUST be a JSON array. Each item must have: title (string), time_minutes (number), ingredients (array of strings), steps (array of strings).",
		count,
		productName,
	)

	payload := openAIResponsesRequest{
		Model:           model,
		Input:           input,
		MaxOutputTokens: 700,
		Temperature:     0.7,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 25 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/responses", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("LLM API returned status: %d", resp.StatusCode)
	}

	var out openAIResponsesResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	content := strings.TrimSpace(out.OutputText)
	if content == "" {
		var sb strings.Builder
		for _, o := range out.Output {
			for _, c := range o.Content {
				if strings.TrimSpace(c.Text) != "" {
					sb.WriteString(c.Text)
				}
			}
		}
		content = strings.TrimSpace(sb.String())
	}
	if content == "" {
		return nil, errors.New("LLM returned empty response")
	}

	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var recipes []Recipe
	if err := json.Unmarshal([]byte(content), &recipes); err == nil {
		return recipes, nil
	}

	start := strings.Index(content, "[")
	end := strings.LastIndex(content, "]")
	if start >= 0 && end > start {
		candidate := content[start : end+1]
		if err := json.Unmarshal([]byte(candidate), &recipes); err == nil {
			return recipes, nil
		}
	}

	if n, err := strconv.Atoi(strconv.Itoa(count)); err == nil && n > 0 {
		return nil, errors.New("failed to parse recipes JSON from LLM response")
	}

	return nil, errors.New("failed to parse recipes JSON from LLM response")
}
