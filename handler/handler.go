package handler

import (
	"net/http"
	"os"

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

	// Simpan pesan pengguna ke riwayat percakapan
	userMessage := model.Message{
		Role: "user",
		Parts: []model.Part{
			{Text: input.Query},
		},
	}
	err := h.Service.SaveChatHistory(input.SessionID, userMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving chat history"})
		return
	}

	// Buat payload untuk model
	inputs := model.Inputs{
		Table: h.Table,
		Query: input.Query,
	}

	// Ambil response dari model TAPAS
	token := os.Getenv("HUGGINGFACE_TOKEN")
	if token == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "HUGGINGFACE_TOKEN environment variable not set"})
		return
	}

	response, err := h.Service.GetAIResponse(inputs, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to AI model"})
		return
	}

	// Simpan response dari model TAPAS ke riwayat percakapan
	assistantMessage := model.Message{
		Role: "assistant",
		Parts: []model.Part{
			{Text: response.Answer},
		},
	}
	err = h.Service.SaveChatHistory(input.SessionID, assistantMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving chat history"})
		return
	}

	// Ambil rekomendasi dari model Gemini
	apiKey := os.Getenv("API_KEY_GEMINI")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API_KEY_GEMINI environment variable not set"})
		return
	}

	geminiResponse, err := h.Service.GetGeminiRecommendation(input.Query, h.Table, apiKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting Gemini recommendation"})
		return
	}

	for i := range geminiResponse.Candidates {
		geminiResponse.Candidates[i].Content.Role = "assistant"
	}

	// Simpan rekomendasi dari model Gemini ke riwayat percakapan
	for _, candidate := range geminiResponse.Candidates {
		assistantMessage := model.Message{
			Role:  "assistant",
			Parts: candidate.Content.Parts,
		}
		err = h.Service.SaveChatHistory(input.SessionID, assistantMessage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving chat history"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tapas_response":         response,
		"gemini_recommendations": geminiResponse.Candidates,
	})
}
