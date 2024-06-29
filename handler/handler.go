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
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	normalizedQuery := strings.ToLower(strings.TrimSpace(input.Query))
	if normalizedQuery == "halo" || normalizedQuery == "hi" {
		response := model.APIResponse{
			Candidates: []model.Candidate{
				{
					Content: model.Content{
						Role: "assistant",
						Parts: []model.Part{
							{Text: "Halo! Saya Luma, AI Assistant yang bisa membantu kamu seputar Energi di Smarthome kamu. Data yang Anda berikan adalah tentang penggunaan peralatan rumah tangga di berbagai waktu dan kondisi. Apakah Anda ingin saya melakukan analisis atau memberikan informasi lebih lanjut tentang data ini?"},
						},
					},
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

	c.JSON(http.StatusOK, gin.H{
		"tapas_answer":           response.Answer,
		"gemini_recommendations": geminiResponse.Candidates,
	})
}
