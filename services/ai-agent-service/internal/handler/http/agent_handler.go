package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"omnistock-ai/services/ai-agent-service/internal/usecase"
)

type AgentHandler struct {
	usecase usecase.AgentUsecase
}

// NewAgentHandler initializes the handler and registers routes
func NewAgentHandler(router *gin.Engine, us usecase.AgentUsecase) {
	handler := &AgentHandler{
		usecase: us,
	}

	// Register the POST route for chatting with AI
	router.POST("/api/v1/chat", handler.Chat)
}

// ChatRequest represents the incoming JSON payload from the user
type ChatRequest struct {
	Message string `json:"message" binding:"required"`
	Role    string `json:"role"` // "picker" or "packer", defaults to "picker"
}

// Chat handles the HTTP request to interact with the AI agent
func (h *AgentHandler) Chat(c *gin.Context) {
	var req ChatRequest

	// 1. Extract JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message field is required"})
		return
	}

	if req.Role == "" {
		req.Role = "picker"
	}

	// 2. Send the message to the AI Usecase
	// We pass the Request Context so if the user closes the browser, the request can be cancelled
	reply, err := h.usecase.ChatWithAgent(c.Request.Context(), req.Role, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Return the AI's generated response
	c.JSON(http.StatusOK, gin.H{
		"reply": reply,
	})
}
