package eventbus

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// EventBus interface for event publication and subscription
type EventBusInterface interface {
	Publish(topic string, data interface{}) error
	Subscribe(topic string, handler func([]byte))
	Close() error
	GetSubscriberCount(topic string) int
	Health() map[string]interface{}
}

// EventBus handles in-memory event publishing and subscription
// NOTE: This is an in-memory implementation. Events are lost on restart.
// Can be upgraded to Redis when Docker is available.
type EventBus struct {
	subscribers map[string][]chan []byte
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// Event represents a generic event with metadata
type Event struct {
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Version   string                 `json:"version"`
}

// NewEventBus creates a new in-memory event bus
func NewEventBus() *EventBus {
	ctx, cancel := context.WithCancel(context.Background())
	log.Println("[EVENTBUS] ✅ Initialized in-memory EventBus")
	return &EventBus{
		subscribers: make(map[string][]chan []byte),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// NewEventBusWithRedis creates EventBus with Redis support (if URL provided)
// Falls back to in-memory if Redis connection fails
func NewEventBusWithRedis(redisURL string) EventBusInterface {
	if redisURL != "" {
		// Try Redis first
		redisEB, err := NewRedisEventBus(redisURL)
		if err != nil {
			log.Printf("[EVENTBUS] ⚠️  Failed to connect to Redis: %v", err)
			log.Println("[EVENTBUS] Falling back to in-memory EventBus")
			return NewEventBus()
		}
		return redisEB
	}

	// Default to in-memory
	return NewEventBus()
}

// Publish publishes an event to all subscribers of the topic
func (eb *EventBus) Publish(topic string, data interface{}) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("[EVENTBUS][ERROR] Failed to marshal event data for topic %s: %v", topic, err)
		return err
	}

	// Send to all subscribers (non-blocking)
	subscribers, exists := eb.subscribers[topic]
	if !exists || len(subscribers) == 0 {
		// No subscribers, event is dropped (this is expected behavior)
		log.Printf("[EVENTBUS][DEBUG] No subscribers for topic: %s", topic)
		return nil
	}

	// Send to each subscriber with timeout protection
	for _, ch := range subscribers {
		select {
		case ch <- jsonData:
			// Event delivered successfully
		case <-time.After(100 * time.Millisecond):
			// Subscriber is slow/blocked, skip to prevent blocking publisher
			log.Printf("[EVENTBUS][WARN] Subscriber for topic %s is slow, skipping delivery", topic)
		case <-eb.ctx.Done():
			// EventBus is shutting down
			return eb.ctx.Err()
		}
	}

	log.Printf("[EVENTBUS][INFO] Published event to topic: %s (%d subscribers)", topic, len(subscribers))
	return nil
}

// Subscribe subscribes to a topic and calls the handler for each event
func (eb *EventBus) Subscribe(topic string, handler func([]byte)) {
	eb.mu.Lock()

	// Create buffered channel for this subscriber
	ch := make(chan []byte, 100) // Buffer 100 events

	if eb.subscribers[topic] == nil {
		eb.subscribers[topic] = []chan []byte{}
	}
	eb.subscribers[topic] = append(eb.subscribers[topic], ch)

	subscriberCount := len(eb.subscribers[topic])
	eb.mu.Unlock()

	log.Printf("[EVENTBUS][INFO] New subscriber for topic: %s (total: %d)", topic, subscriberCount)

	// Start goroutine to process events
	go func() {
		for {
			select {
			case msg := <-ch:
				// Call handler with event data
				handler(msg)
			case <-eb.ctx.Done():
				// EventBus is shutting down
				return
			}
		}
	}()
}

// PublishEvent publishes a typed event (helper method)
func (eb *EventBus) PublishEvent(eventType string, version string, data map[string]interface{}) error {
	event := Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
		Version:   version,
	}
	return eb.Publish(eventType, event)
}

// Close gracefully shuts down the event bus
func (eb *EventBus) Close() error {
	log.Println("[EVENTBUS][INFO] Shutting down event bus...")
	eb.cancel()

	// Close all subscriber channels
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for topic, subscribers := range eb.subscribers {
		for _, ch := range subscribers {
			close(ch)
		}
		log.Printf("[EVENTBUS][INFO] Closed %d subscribers for topic: %s", len(subscribers), topic)
	}

	eb.subscribers = make(map[string][]chan []byte)
	log.Println("[EVENTBUS][INFO] Event bus shut down complete")
	return nil
}

// GetSubscriberCount returns the number of subscribers for a topic
func (eb *EventBus) GetSubscriberCount(topic string) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	subscribers, exists := eb.subscribers[topic]
	if !exists {
		return 0
	}
	return len(subscribers)
}

// GetTopics returns all active topics
func (eb *EventBus) GetTopics() []string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	topics := make([]string, 0, len(eb.subscribers))
	for topic := range eb.subscribers {
		topics = append(topics, topic)
	}
	return topics
}

// Health returns the health status of the event bus
func (eb *EventBus) Health() map[string]interface{} {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	totalSubscribers := 0
	for _, subscribers := range eb.subscribers {
		totalSubscribers += len(subscribers)
	}

	return map[string]interface{}{
		"status":            "healthy",
		"type":              "in-memory",
		"topics":            len(eb.subscribers),
		"total_subscribers": totalSubscribers,
		"note":              "Events are not persisted (in-memory only)",
	}
}
