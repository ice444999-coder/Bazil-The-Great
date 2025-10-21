/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package services

import (
	"ares_api/internal/api/dto"
	repo "ares_api/internal/interfaces/repository"
	"ares_api/internal/models"
	"ares_api/pkg/llm"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type ChatService struct {
	Repo          repo.ChatRepository
	MemoryRepo    repo.MemoryRepository
	LLMClient     *llm.Client
	SystemContext *SystemContextService
	ACEEnabled    bool                                // ACE consciousness cycle enabled
	contextMgrs   map[uint]*llm.ChatWithContextWindow // Per-user context managers
	mu            sync.RWMutex
}

func NewChatService(repo repo.ChatRepository, llmClient *llm.Client) *ChatService {
	return &ChatService{
		Repo:        repo,
		LLMClient:   llmClient,
		contextMgrs: make(map[uint]*llm.ChatWithContextWindow),
	}
}

// NewChatServiceWithMemory creates a chat service with enhanced memory and system context
func NewChatServiceWithMemory(chatRepo repo.ChatRepository, memoryRepo repo.MemoryRepository, llmClient *llm.Client) *ChatService {
	return &ChatService{
		Repo:          chatRepo,
		MemoryRepo:    memoryRepo,
		LLMClient:     llmClient,
		SystemContext: NewSystemContextService(memoryRepo),
		contextMgrs:   make(map[uint]*llm.ChatWithContextWindow),
	}
}

// SetACEEnabled enables ACE framework for consciousness-aware decision making
func (s *ChatService) SetACEEnabled(enabled bool) {
	s.ACEEnabled = enabled
	if enabled {
		log.Println("🧠 ACE Framework enabled for ChatService")
	}
}

// ACEQualityAssessment holds quality assessment data for ACE
type ACEQualityAssessment struct {
	UserID   uint
	Message  string
	Response string
	Context  map[string]interface{}
}

// getOrCreateContextManager gets or creates a context manager for a user
func (s *ChatService) getOrCreateContextManager(userID uint) *llm.ChatWithContextWindow {
	s.mu.RLock()
	mgr, exists := s.contextMgrs[userID]
	s.mu.RUnlock()

	if exists {
		return mgr
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if mgr, exists := s.contextMgrs[userID]; exists {
		return mgr
	}

	// Create new context manager with 2-hour rolling window
	mgr = llm.NewChatWithContextWindow(s.LLMClient)

	// Load recent chat history from database to restore context
	history, err := s.Repo.GetByUserID(userID, 20)
	if err == nil && len(history) > 0 {
		log.Printf("📚 Loading %d previous messages for user %d into context", len(history), userID)
		for _, chat := range history {
			// Add each historical message pair to context (user + assistant)
			userMsg := llm.Message{Role: "user", Content: chat.Message}
			assistantMsg := llm.Message{Role: "assistant", Content: chat.Response}

			mgr.ContextManager.AddMessage(userMsg, llm.EstimateTokens(chat.Message))
			mgr.ContextManager.AddMessage(assistantMsg, llm.EstimateTokens(chat.Response))
		}
	}

	s.contextMgrs[userID] = mgr

	log.Printf("✅ Created context manager for user %d (2-hour rolling window, 150k token budget)", userID)
	return mgr
}

func (s *ChatService) SendMessage(userID uint, req dto.ChatRequest) (dto.ChatResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get user's context manager (maintains 2-hour rolling window)
	contextMgr := s.getOrCreateContextManager(userID)

	// Check token budget before proceeding
	stats := contextMgr.GetContextStats()
	log.Printf("[User %d] Context Stats: %d/%d tokens (%.1f%% used), %d messages in window",
		userID, stats.UsedTokens, stats.MaxTokens, stats.UtilizationPct, stats.MessageCount)

	// 🧠 MEMORY ENHANCEMENT: Enrich message with system context and memories
	enrichedMessage := req.Message
	if s.SystemContext != nil {
		var err error
		enrichedMessage, err = s.SystemContext.EnrichPrompt(userID, req.Message)
		if err != nil {
			log.Printf("⚠️ Failed to enrich prompt with system context: %v (continuing with original)", err)
			enrichedMessage = req.Message
		} else {
			log.Printf("🧠 Enhanced message with system context and memory")
		}
	}

	// Generate response with automatic context management (using enriched message)
	respText, err := contextMgr.SendMessage(ctx, enrichedMessage, llm.TempGeneral)
	if err != nil {
		// Check if it's a token budget error
		if stats.RemainingTokens < 1000 {
			log.Printf("⚠️ User %d approaching token limit, resetting context window", userID)
			contextMgr.ResetContext()
			return dto.ChatResponse{}, fmt.Errorf("token budget exceeded, context window reset - please try again")
		}
		return dto.ChatResponse{}, err
	}

	// 🧠 ACE FRAMEWORK: Quality assessment and pattern learning (if enabled)
	if s.ACEEnabled {
		log.Printf("🧠 ACE: Response quality assessment queued for user %d", userID)
		// Note: ACE quality scoring happens in controller layer to avoid import cycles
	}

	// Store in database
	chat := &models.Chat{
		UserID:   userID,
		Message:  req.Message,
		Response: respText,
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
