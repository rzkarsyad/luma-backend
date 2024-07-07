package service

import (
	"luma-backend/model"
	"luma-backend/repository"
)

type AIService struct {
	Connector *repository.AIModelConnector
	ChatRepo  *repository.ChatRepository
}

func (s *AIService) GetAIResponse(inputs model.Inputs, token string) (model.Response, error) {
	return s.Connector.ConnectAIModel(inputs, token)
}

func (s *AIService) GetGeminiRecommendation(sessionID, query string, table map[string][]string, token string) (model.APIResponse, error) {
	chatHistory, err := s.GetChatHistory(sessionID)
	if err != nil {
		return model.APIResponse{}, err
	}
	return s.Connector.GeminiRecommendationWithHistory(query, table, token, chatHistory)
}


func (s *AIService) SaveChatHistory(sessionID string, message model.Message) error {
	return s.ChatRepo.SaveMessage(sessionID, message)
}

func (s *AIService) GetChatHistory(sessionID string) ([]model.Message, error) {
	return s.ChatRepo.GetMessages(sessionID)
}
