package services

import (
	"ares_api/internal/api/dto"
	repo "ares_api/internal/interfaces/repository"
	"ares_api/internal/models"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MemoryService struct {
	Repo repo.MemoryRepository
}

func NewMemoryService(repo repo.MemoryRepository) *MemoryService {
	return &MemoryService{Repo: repo}
}

func (s *MemoryService) Learn(userID uint, eventType string, payload interface{}, sessionID *uuid.UUID) error {
	// Convert payload to JSONB type
	jsonbPayload, ok := payload.(map[string]interface{})
	if !ok {
		jsonbPayload = map[string]interface{}{"data": payload}
	}

	snapshot := &models.MemorySnapshot{
		Timestamp: time.Now(),
		EventType: eventType,
		Payload:   models.JSONB(jsonbPayload),
		UserID:    userID,
		SessionID: sessionID,
	}

	return s.Repo.SaveSnapshot(snapshot)
}

func (s *MemoryService) Recall(userID uint, limit int) ([]dto.MemoryRecallResponse, error) {
	snapshots, err := s.Repo.GetRecentSnapshots(userID, limit)
	if err != nil {
		return nil, err
	}

	return s.convertToDTO(snapshots), nil
}

func (s *MemoryService) RecallByEventType(userID uint, eventType string, limit int) ([]dto.MemoryRecallResponse, error) {
	snapshots, err := s.Repo.GetSnapshotsByEventType(userID, eventType, limit)
	if err != nil {
		return nil, err
	}

	return s.convertToDTO(snapshots), nil
}

func (s *MemoryService) RecallBySessionID(sessionID uuid.UUID, limit int) ([]dto.MemoryRecallResponse, error) {
	snapshots, err := s.Repo.GetSnapshotsBySessionID(sessionID, limit)
	if err != nil {
		return nil, err
	}

	return s.convertToDTO(snapshots), nil
}

func (s *MemoryService) convertToDTO(snapshots []models.MemorySnapshot) []dto.MemoryRecallResponse {
	responses := make([]dto.MemoryRecallResponse, len(snapshots))
	for i, snapshot := range snapshots {
		var sessionIDStr *string
		if snapshot.SessionID != nil {
			str := snapshot.SessionID.String()
			sessionIDStr = &str
		}

		responses[i] = dto.MemoryRecallResponse{
			ID:        snapshot.ID,
			Timestamp: snapshot.Timestamp.Format(time.RFC3339),
			EventType: snapshot.EventType,
			Payload:   map[string]interface{}(snapshot.Payload),
			UserID:    snapshot.UserID,
			SessionID: sessionIDStr,
		}
	}

	return responses
}

// ImportConversation parses and imports a conversation into memory
func (s *MemoryService) ImportConversation(userID uint, content string, source string, tags []string) (int, uint, error) {
	// Default source if not provided
	if source == "" {
		source = "manual_paste"
	}

	// Default tags if not provided
	if len(tags) == 0 {
		tags = []string{"genesis_conversation"}
	}

	// Parse conversation - look for common patterns
	// Pattern 1: "User: ... Assistant: ..."
	// Pattern 2: JSON array of messages
	messageCount := 0
	sessionID := uuid.New()

	// Try to detect format
	if strings.Contains(content, `"role"`) && strings.Contains(content, `"content"`) {
		// Likely JSON format - but for now we'll store as single entry
		// TODO: Implement proper JSON parsing
	}

	// Split by common delimiters
	userPattern := regexp.MustCompile(`(?i)(user|human|you):\s*`)
	assistantPattern := regexp.MustCompile(`(?i)(assistant|ai|claude|ares):\s*`)

	// Simple heuristic: split into chunks and store each
	lines := strings.Split(content, "\n")
	currentMessage := ""
	currentRole := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this line starts a new message
		if userPattern.MatchString(line) {
			// Save previous message if exists
			if currentMessage != "" {
				s.saveConversationChunk(userID, currentRole, currentMessage, tags, &sessionID)
				messageCount++
			}
			currentRole = "user"
			currentMessage = userPattern.ReplaceAllString(line, "")
		} else if assistantPattern.MatchString(line) {
			// Save previous message if exists
			if currentMessage != "" {
				s.saveConversationChunk(userID, currentRole, currentMessage, tags, &sessionID)
				messageCount++
			}
			currentRole = "assistant"
			currentMessage = assistantPattern.ReplaceAllString(line, "")
		} else {
			// Continuation of current message
			if currentMessage != "" {
				currentMessage += "\n" + line
			} else {
				currentMessage = line
			}
		}
	}

	// Save last message
	if currentMessage != "" {
		s.saveConversationChunk(userID, currentRole, currentMessage, tags, &sessionID)
		messageCount++
	}

	// If no messages were parsed (no user/assistant markers), store entire content as single memory
	if messageCount == 0 {
		payload := map[string]interface{}{
			"content": content,
			"tags":    tags,
			"source":  source,
		}
		snapshot := &models.MemorySnapshot{
			Timestamp: time.Now(),
			EventType: "conversation_import",
			Payload:   models.JSONB(payload),
			UserID:    userID,
			SessionID: &sessionID,
		}
		s.Repo.SaveSnapshot(snapshot)
		messageCount = 1
	}

	// Create ConversationImport record
	// TODO: Add ConversationImport repository method
	// For now, this is tracked via memory_snapshots

	return messageCount, userID, nil
}

func (s *MemoryService) saveConversationChunk(userID uint, role string, content string, tags []string, sessionID *uuid.UUID) {
	payload := map[string]interface{}{
		"role":    role,
		"content": content,
		"tags":    tags,
	}
	snapshot := &models.MemorySnapshot{
		Timestamp: time.Now(),
		EventType: "conversation_message",
		Payload:   models.JSONB(payload),
		UserID:    userID,
		SessionID: sessionID,
	}
	s.Repo.SaveSnapshot(snapshot)
}
