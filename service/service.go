package service

import (
	"luma-backend/model"
	"luma-backend/repository"
)

type AIService struct {
	Connector *repository.AIModelConnector
}

func (s *AIService) GetAIResponse(inputs model.Inputs, token string) (model.Response, error) {
	return s.Connector.ConnectAIModel(inputs, token)
}

func (s *AIService) GetGeminiRecommendation(query string, table map[string][]string, token string) (model.APIResponse, error) {
	return s.Connector.GeminiRecommendation(query, table, token)
}
