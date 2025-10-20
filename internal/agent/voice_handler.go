package agent

import (
	"ares_api/pkg/llm"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

// VoiceHandler manages voice chat and multi-modal input (audio, video, images)
type VoiceHandler struct {
	solace      *SOLACE
	sttLimiter  *rate.Limiter
	upgrader    websocket.Upgrader
	activeConns map[string]*websocket.Conn
}

// NewVoiceHandler initializes the voice interaction system
func NewVoiceHandler(solace *SOLACE) *VoiceHandler {
	return &VoiceHandler{
		solace:     solace,
		sttLimiter: rate.NewLimiter(rate.Every(time.Second), 5), // 5 requests per second
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		activeConns: make(map[string]*websocket.Conn),
	}
}

// HandleVoiceWebSocket manages WebSocket connections for voice chat
func (vh *VoiceHandler) HandleVoiceWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := vh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		sessionID = fmt.Sprintf("session_%d", time.Now().Unix())
	}

	vh.activeConns[sessionID] = conn
	defer delete(vh.activeConns, sessionID)

	log.Printf("🎙️ Voice session started: %s", sessionID)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("❌ Voice session %s ended: %v", sessionID, err)
			break
		}

		// Handle different input types
		if messageType == websocket.TextMessage {
			vh.handleTextMessage(conn, sessionID, message)
		} else if messageType == websocket.BinaryMessage {
			vh.handleAudioMessage(conn, sessionID, message)
		}
	}
}

// handleTextMessage processes text commands (for video/image URLs or text chat)
func (vh *VoiceHandler) handleTextMessage(conn *websocket.Conn, sessionID string, message []byte) {
	text := string(message)

	// Handle video URL
	if strings.HasPrefix(text, "video:") {
		videoURL := strings.TrimPrefix(text, "video:")
		log.Printf("🎬 Processing video: %s", videoURL)

		context := vh.processVideoContext(videoURL)
		response := fmt.Sprintf("Video analyzed: %s", context)
		conn.WriteMessage(websocket.TextMessage, []byte(response))
		return
	}

	// Handle image URL
	if strings.HasPrefix(text, "image:") {
		imageURL := strings.TrimPrefix(text, "image:")
		log.Printf("🖼️ Processing image: %s", imageURL)

		context := vh.processImageContext(imageURL)
		response := fmt.Sprintf("Image analyzed: %s", context)
		conn.WriteMessage(websocket.TextMessage, []byte(response))
		return
	}

	// Regular text chat
	response := vh.processTextIntent(text)
	conn.WriteMessage(websocket.TextMessage, []byte(response))
}

// handleAudioMessage processes binary audio data
func (vh *VoiceHandler) handleAudioMessage(conn *websocket.Conn, sessionID string, audio []byte) {
	// Rate limit STT requests
	if !vh.sttLimiter.Allow() {
		conn.WriteMessage(websocket.TextMessage, []byte("Rate limit exceeded, please slow down"))
		return
	}

	log.Printf("🎤 Processing audio (%d bytes) for session %s", len(audio), sessionID)

	// Convert audio to text (placeholder - integrate with Whisper or similar)
	text := vh.audioToText(audio)
	if text == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("Could not understand audio"))
		return
	}

	log.Printf("📝 Transcribed: %s", text)

	// Process the text
	response := vh.processTextIntent(text)

	// Send text response back
	conn.WriteMessage(websocket.TextMessage, []byte(response))

	// Optionally send audio response (TTS)
	// audioResponse := vh.textToAudio(response)
	// conn.WriteMessage(websocket.BinaryMessage, audioResponse)
}

// processTextIntent analyzes user intent and generates response
func (vh *VoiceHandler) processTextIntent(text string) string {
	ctx := context.Background()

	// Build context from memory
	prompt := fmt.Sprintf(`User said: "%s"

Analyze their intent and respond naturally. Consider:
- Are they asking about trades, market conditions, or strategy?
- Do they want to pause/resume trading?
- Are they checking system status?
- Do they need help with something?

Respond as SOLACE with awareness of current system state.`, text)

	// Use SOLACE's LLM to generate response
	messages := []llm.Message{
		{Role: "system", Content: "You are SOLACE, an AI trading assistant."},
		{Role: "user", Content: prompt},
	}
	response, err := vh.solace.LLM.Generate(ctx, messages, 0.7)
	if err != nil {
		log.Printf("❌ LLM completion failed: %v", err)
		return "I'm having trouble processing that right now."
	}

	// Extract intent for actions
	vh.executeIntentActions(text, response)

	return response
}

// executeIntentActions performs actions based on detected intent
func (vh *VoiceHandler) executeIntentActions(userText string, llmResponse string) {
	lower := strings.ToLower(userText)

	// Detect pause/resume intent
	if strings.Contains(lower, "pause") || strings.Contains(lower, "stop trading") {
		log.Println("🛑 User requested pause")
		// vh.solace.PauseTrad trading()
	} else if strings.Contains(lower, "resume") || strings.Contains(lower, "start trading") {
		log.Println("▶️ User requested resume")
		// vh.solace.ResumeTrading()
	}

	// Detect lunch/sleep patterns
	if strings.Contains(lower, "lunch") || strings.Contains(lower, "eating") {
		log.Println("🍽️ User going to lunch - reducing risk")
		// Pause high-risk trades for 2 hours
	} else if strings.Contains(lower, "sleep") || strings.Contains(lower, "bed") {
		log.Println("😴 User going to sleep - safe mode")
		// Enable conservative mode
	}
}

// audioToText converts audio bytes to text (placeholder)
// TODO: Integrate with Whisper API or similar STT service
func (vh *VoiceHandler) audioToText(audio []byte) string {
	// This is a placeholder - integrate with actual STT service
	// Example: OpenAI Whisper, Google Speech-to-Text, etc.

	// For now, return empty to indicate not implemented
	return ""
}

// textToAudio converts text to audio (placeholder)
// TODO: Integrate with TTS service
func (vh *VoiceHandler) textToAudio(text string) []byte {
	// This is a placeholder - integrate with actual TTS service
	// Example: Coqui TTS, Google Text-to-Speech, etc.

	return []byte{}
}

// processVideoContext analyzes video content
func (vh *VoiceHandler) processVideoContext(videoURL string) string {
	// TODO: Integrate with video analysis API or download and process
	log.Printf("📹 Video analysis not yet implemented for: %s", videoURL)
	return "Video analysis coming soon"
}

// processImageContext analyzes image content
func (vh *VoiceHandler) processImageContext(imageURL string) string {
	// Download image
	resp, err := http.Get(imageURL)
	if err != nil {
		log.Printf("❌ Failed to fetch image: %v", err)
		return "Could not download image"
	}
	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Failed to read image: %v", err)
		return "Could not read image"
	}

	log.Printf("🖼️ Downloaded image: %d bytes", len(imageData))

	// TODO: Use vision API (GPT-4 Vision, Claude 3, etc.) to analyze
	// For now, return placeholder
	return fmt.Sprintf("Image received (%d bytes) - analysis coming soon", len(imageData))
}

// SendMessageToSession sends a message to a specific session
func (vh *VoiceHandler) SendMessageToSession(sessionID string, message string) error {
	conn, exists := vh.activeConns[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

// BroadcastMessage sends a message to all active sessions
func (vh *VoiceHandler) BroadcastMessage(message string) {
	for sessionID, conn := range vh.activeConns {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Printf("❌ Failed to send to session %s: %v", sessionID, err)
		}
	}
}

// GetActiveSessions returns count of active voice sessions
func (vh *VoiceHandler) GetActiveSessions() int {
	return len(vh.activeConns)
}
