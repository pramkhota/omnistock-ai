package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// AgentUsecase defines the interface for interacting with the AI
type AgentUsecase interface {
	ChatWithAgent(ctx context.Context, userMessage string) (string, error)
}

type agentUsecase struct {
	apiKey      string
	sopText     string
	toolsConfig []map[string]interface{} // Store tools configuration from JSON
}

// NewAgentUsecase initializes the AI Service using Native HTTP approach
func NewAgentUsecase() (AgentUsecase, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is not set in .env")
	}

	// 1. Load SOP text
	sopBytes, err := os.ReadFile("services/ai-agent-service/config/sop.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to load SOP config: %v", err)
	}

	// 2. Load Tools configuration
	toolsBytes, err := os.ReadFile("services/ai-agent-service/config/tools.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load tools config: %v", err)
	}

	var toolsConfig []map[string]interface{}
	if err := json.Unmarshal(toolsBytes, &toolsConfig); err != nil {
		return nil, fmt.Errorf("failed to parse tools json: %v", err)
	}

	return &agentUsecase{
		apiKey:      apiKey,
		sopText:     string(sopBytes),
		toolsConfig: toolsConfig,
	}, nil
}

// ChatWithAgent sends the message to Gemini and handles Function Calling (Tools)
func (u *agentUsecase) ChatWithAgent(ctx context.Context, userMessage string) (string, error) {
	if userMessage == "" {
		return "", errors.New("message cannot be empty")
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-flash-lite-latest:generateContent?key=%s", u.apiKey)

	// Inject tools and SOP into the JSON payload dynamically
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{{"text": userMessage}},
			},
		},
		"systemInstruction": map[string]interface{}{
			"parts": []map[string]interface{}{{"text": u.sopText}},
		},
		"tools": u.toolsConfig, // Injected from tools.json!
	}

	jsonValue, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Google API Error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the JSON response to support both text and functionCall
	var respData struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text         string `json:"text,omitempty"`
					FunctionCall *struct {
						Name string                 `json:"name"`
						Args map[string]interface{} `json:"args"`
					} `json:"functionCall,omitempty"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(bodyBytes, &respData); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	if len(respData.Candidates) > 0 && len(respData.Candidates[0].Content.Parts) > 0 {
		part := respData.Candidates[0].Content.Parts[0]

		// Check if the AI decided to call a tool
		if part.FunctionCall != nil {
			toolName := part.FunctionCall.Name
			toolArgs := part.FunctionCall.Args

			// For now, just return the decision
			return fmt.Sprintf("⚡ [AI DECISION]: AI wants to execute tool '%s' with args: %v", toolName, toolArgs), nil
		}

		// Return standard text response if no tool was called
		return part.Text, nil
	}

	return "", errors.New("received empty response from AI")
}
