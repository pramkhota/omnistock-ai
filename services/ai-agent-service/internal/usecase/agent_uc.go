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
	ChatWithAgent(ctx context.Context, role string, userMessage string) (string, error)
}

type agentUsecase struct {
	apiKey      string
	pickerSOP   string
	packerSOP   string
	toolsConfig []map[string]interface{}
}

// NewAgentUsecase initializes the AI Service
func NewAgentUsecase() (AgentUsecase, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is not set in .env")
	}

	pickerBytes, err := os.ReadFile("services/ai-agent-service/config/sop.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to load Picker SOP: %v", err)
	}

	packerBytes, err := os.ReadFile("services/ai-agent-service/config/sop_packer.txt")
	if err != nil {
		// If the file doesn't exist yet, we just leave it empty or handle it, but we know we created it.
		return nil, fmt.Errorf("failed to load Packer SOP: %v", err)
	}

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
		pickerSOP:   string(pickerBytes),
		packerSOP:   string(packerBytes),
		toolsConfig: toolsConfig,
	}, nil
}

func (u *agentUsecase) ChatWithAgent(ctx context.Context, role string, userMessage string) (string, error) {
	if userMessage == "" {
		return "", errors.New("message cannot be empty")
	}

	activeSOP := u.pickerSOP
	if role == "packer" {
		activeSOP = u.packerSOP
	}

	contents := []map[string]interface{}{
		{
			"role":  "user",
			"parts": []map[string]interface{}{{"text": userMessage}},
		},
	}

	var finalDecisions []string

	for i := 0; i < 4; i++ {
		reply, toolName, toolArgs, rawParts, err := u.callGeminiAPI(ctx, activeSOP, contents)
		if err != nil {
			return "", err
		}

		if toolName == "" {
			if len(finalDecisions) > 0 {
				return fmt.Sprintf("⚡ [FINAL AI DECISIONS]: %v\nMessage: %s", finalDecisions, reply), nil
			}
			return reply, nil
		}

		// Handle intermediate tools (ReAct loop)
		if toolName == "check_stock" || toolName == "calc_packaging" {
			fmt.Printf("🔍 AI called %s with args: %v\n", toolName, toolArgs)
			
			contents = append(contents, map[string]interface{}{
				"role": "model",
				"parts": rawParts,
			})

			var toolResp map[string]interface{}
			
			if toolName == "check_stock" {
				mockStockQty := 0
				sku, _ := toolArgs["sku"].(string)
				if sku == "SKU-HAVE-SPARE" || sku == "SKU001" {
					mockStockQty = 5
				}
				toolResp = map[string]interface{}{"stock_qty": mockStockQty}
			} else if toolName == "calc_packaging" {
				toolResp = map[string]interface{}{
					"items": []map[string]interface{}{
						{"sku": "SKU001", "width": 10, "height": 20, "length": 15},
						{"sku": "SKU002", "width": 5, "height": 5, "length": 5},
					},
				}
			}

			contents = append(contents, map[string]interface{}{
				"role": "user",
				"parts": []map[string]interface{}{
					{
						"functionResponse": map[string]interface{}{
							"name": toolName,
							"response": toolResp,
						},
					},
				},
			})
			continue
		}

		// Handle final decision tools
		if toolName == "swap_item" || toolName == "abort_order" || toolName == "report_damage_only" || toolName == "request_repick" || toolName == "return_to_shelf" {
			decisionMsg := fmt.Sprintf("Tool '%s' with args: %v", toolName, toolArgs)
			finalDecisions = append(finalDecisions, decisionMsg)
			
			contents = append(contents, map[string]interface{}{
				"role": "model",
				"parts": rawParts,
			})
			contents = append(contents, map[string]interface{}{
				"role": "user",
				"parts": []map[string]interface{}{
					{
						"functionResponse": map[string]interface{}{
							"name": toolName,
							"response": map[string]interface{}{"status": "success"},
						},
					},
				},
			})
			
			continue
		}

		return fmt.Sprintf("⚠️ AI called unknown tool '%s'", toolName), nil
	}

	if len(finalDecisions) > 0 {
		return fmt.Sprintf("⚡ [FINAL AI DECISIONS]: %v\n(Loop ended)", finalDecisions), nil
	}
	return "", errors.New("AI loop limit reached")
}

func (u *agentUsecase) callGeminiAPI(ctx context.Context, sopText string, contents []map[string]interface{}) (reply string, toolName string, toolArgs map[string]interface{}, rawParts []map[string]interface{}, err error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-flash-lite-latest:generateContent?key=%s", u.apiKey)

	reqBody := map[string]interface{}{
		"contents": contents,
		"systemInstruction": map[string]interface{}{
			"parts": []map[string]interface{}{{"text": sopText}},
		},
		"tools": u.toolsConfig,
	}

	jsonValue, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return "", "", nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", nil, nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", nil, nil, fmt.Errorf("Google API Error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var respData struct {
		Candidates []struct {
			Content struct {
				Parts []map[string]interface{} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(bodyBytes, &respData); err != nil {
		return "", "", nil, nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if len(respData.Candidates) > 0 && len(respData.Candidates[0].Content.Parts) > 0 {
		parts := respData.Candidates[0].Content.Parts
		
		// Find if there's a functionCall
		for _, p := range parts {
			if fc, ok := p["functionCall"].(map[string]interface{}); ok {
				name, _ := fc["name"].(string)
				args, _ := fc["args"].(map[string]interface{})
				return "", name, args, parts, nil
			}
		}

		// Otherwise return text
		if txt, ok := parts[0]["text"].(string); ok {
			return txt, "", nil, parts, nil
		}
	}

	return "", "", nil, nil, errors.New("received empty response from AI")
}
