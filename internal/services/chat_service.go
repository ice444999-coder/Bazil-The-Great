package services

import (
	"ares_api/internal/api/dto"
	repo "ares_api/internal/interfaces/repository"
	"ares_api/internal/models"
	"ares_api/internal/ollama"
	"time"
)

type ChatService struct {
	Repo        repo.ChatRepository
	OllamaClient *ollama.Client
}

func NewChatService(repo repo.ChatRepository, client *ollama.Client) *ChatService {
	return &ChatService{Repo: repo, OllamaClient: client}
}

func (s *ChatService) SendMessage(userID uint, req dto.ChatRequest) (dto.ChatResponse, error) {


	respText, err := s.OllamaClient.Chat("llama3", req.Message)
	if err != nil {
		return dto.ChatResponse{}, err
	}

	chat := &models.Chat{
		UserID:   userID,
		Message:  req.Message,
		Response: respText,
		CreatedAt: time.Now(),
	}

	if err := s.Repo.Create(chat); err != nil {
		return dto.ChatResponse{}, err
	}

	return dto.ChatResponse{
		Message:  req.Message,
		Response: respText,
	}, nil
}

func (s *ChatService) GetHistory(userID uint, limit int) ([]dto.ChatHistoryResponse, error) {
	chats, err := s.Repo.GetByUserID(userID, limit)
	if err != nil {
		return nil, err
	}

	res := make([]dto.ChatHistoryResponse, len(chats))
	for i, c := range chats {
		res[i] = dto.ChatHistoryResponse{
			ID:        c.ID,
			UserID:    c.UserID,
			Message:   c.Message,
			Response:  c.Response,
			CreatedAt: c.CreatedAt.Format(time.RFC3339),
		}
	}

	return res, nil
}

