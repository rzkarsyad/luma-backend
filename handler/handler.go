package handler

import (
	"net/http"
	"os"
	"strings"

	"luma-backend/model"
	"luma-backend/service"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	Service *service.AIService
	Table   map[string][]string
}

func (h *AIHandler) HandleRequest(c *gin.Context) {
	var input struct {
		Query     string `json:"query"`
		SessionID string `json:"session_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not found in Authorization header"})
		return
	}

	sessionID := c.GetHeader("session_id")
	if sessionID == "" {
		sessionID = input.SessionID
	}
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session ID not found in headers or body"})
		return
	}

	userMessage := model.Message{
		Role: "user",
		Parts: []model.Part{
			{Text: input.Query},
		},
	}
	err := h.Service.SaveChatHistory(sessionID, userMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving chat history"})
		return
	}

	normalizedQuery := strings.ToLower(strings.TrimSpace(input.Query))
	if normalizedQuery == "halo" || normalizedQuery == "hi" {
		response := struct {
			Answer          string            `json:"answer"`
			Recommendations []model.Candidate `json:"recommendations"`
		}{
			Answer: "",
			Recommendations: []model.Candidate{
				{
					Content: model.Content{
						Role: "assistant",
						Parts: []model.Part{
							{Text: "Halo! Saya Luma, AI Assistant yang bisa membantu kamu seputar penggunaan energi di Smarthome kamu. Data yang Anda berikan adalah tentang penggunaan peralatan rumah tangga di berbagai waktu dan kondisi. Apakah Anda ingin saya melakukan analisis atau memberikan informasi lebih lanjut tentang data ini?"},
						},
					},
					FinishReason: "",
					Index:        0,
				},
			},
		}
		c.JSON(http.StatusOK, response)
		return
	}

	inputs := model.Inputs{
		Table: h.Table,
		Query: input.Query,
	}

	huggingfaceToken := os.Getenv("HUGGINGFACE_TOKEN")
	if huggingfaceToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "HUGGINGFACE_TOKEN environment variable not set"})
		return
	}

	response, err := h.Service.GetAIResponse(inputs, huggingfaceToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to AI model"})
		return
	}

	assistantMessage := model.Message{
		Role: "assistant",
		Parts: []model.Part{
			{Text: response.Answer},
		},
	}
	err = h.Service.SaveChatHistory(sessionID, assistantMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving chat history"})
		return
	}

	apiKey := os.Getenv("API_KEY_GEMINI")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API_KEY_GEMINI environment variable not set"})
		return
	}

	geminiResponse, err := h.Service.GetGeminiRecommendation(sessionID, input.Query, h.Table, apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting Gemini recommendation"})
		return
	}

	for i := range geminiResponse.Candidates {
		geminiResponse.Candidates[i].Content.Role = "assistant"
	}

	for _, candidate := range geminiResponse.Candidates {
		assistantMessage := model.Message{
			Role:  "assistant",
			Parts: candidate.Content.Parts,
		}
		err := h.Service.SaveChatHistory(sessionID, assistantMessage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving chat history"})
			return
		}
	}

	fullResponse := struct {
		Answer          string            `json:"answer"`
		Recommendations []model.Candidate `json:"recommendations"`
	}{
		Answer:          response.Answer,
		Recommendations: geminiResponse.Candidates,
	}

	c.JSON(http.StatusOK, fullResponse)
}

func (h *AIHandler) GetChatHistory(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID not provided"})
		return
	}

	messages, err := h.Service.GetChatHistory(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving chat history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}
