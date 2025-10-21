/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisEventBus implements EventBus using Redis pub/sub
type RedisEventBus struct {
	client      *redis.Client
	ctx         context.Context
	cancel      context.CancelFunc
	subscribers map[string][]chan []byte
	mu          sync.RWMutex
	closed      bool
	pubsub      *redis.PubSub
}

// NewRedisEventBus creates a new Redis-backed event bus
func NewRedisEventBus(redisURL string) (*RedisEventBus, error) {
	// Parse Redis URL
	opts, err := redis.ParseURL(fmt.Sprintf("redis://%s", redisURL))
	if err != nil {
		return nil, fmt.Errorf("invalid redis URL: %w", err)
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	log.Printf("[EVENTBUS] ✅ Connected to Redis at %s", redisURL)

	appCtx, appCancel := context.WithCancel(context.Background())
	eb := &RedisEventBus{
		client:      client,
		ctx:         appCtx,
		cancel:      appCancel,
		subscribers: make(map[string][]chan []byte),
		pubsub:      client.Subscribe(appCtx),
	}

	// Start subscriber goroutine
	go eb.receiveMessages()

	return eb, nil
}

// Publish publishes an event to a topic
func (eb *RedisEventBus) Publish(topic string, data interface{}) error {
	eb.mu.RLock()
	if eb.closed {
		eb.mu.RUnlock()
		return fmt.Errorf("event bus is closed")
	}
	eb.mu.RUnlock()

	// Serialize data
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Publish to Redis
	err = eb.client.Publish(eb.ctx, topic, payload).Err()
	if err != nil {
		return fmt.Errorf("failed to publish to redis: %w", err)
	}

	log.Printf("[EVENTBUS] 📤 Published to Redis topic: %s", topic)
	return nil
}

// Subscribe subscribes to a topic and calls the handler for each event
func (eb *RedisEventBus) Subscribe(topic string, handler func([]byte)) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		log.Printf("[EVENTBUS] ⚠️  Cannot subscribe to %s: event bus is closed", topic)
		return
	}

	// Create channel for this handler
	ch := make(chan []byte, 100)
	eb.subscribers[topic] = append(eb.subscribers[topic], ch)

	// Subscribe to Redis topic if first subscriber
	if len(eb.subscribers[topic]) == 1 {
		if err := eb.pubsub.Subscribe(eb.ctx, topic); err != nil {
			log.Printf("[EVENTBUS] ❌ Failed to subscribe to Redis topic %s: %v", topic, err)
			return
		}
		log.Printf("[EVENTBUS] 📥 Subscribed to Redis topic: %s", topic)
	}

	// Start handler goroutine
	go func() {
		for data := range ch {
			handler(data)
		}
	}()

	log.Printf("[EVENTBUS] ✅ Handler registered for topic: %s", topic)
}

// receiveMessages receives messages from Redis and distributes to subscribers
func (eb *RedisEventBus) receiveMessages() {
	ch := eb.pubsub.Channel()
	for {
		select {
		case msg := <-ch:
			if msg == nil {
				continue
			}

			eb.mu.RLock()
			handlers, ok := eb.subscribers[msg.Channel]
			eb.mu.RUnlock()

			if !ok {
				continue
			}

			// Distribute to all handlers
			payload := []byte(msg.Payload)
			for _, handler := range handlers {
				select {
				case handler <- payload:
				default:
					log.Printf("[EVENTBUS] ⚠️  Handler channel full for topic: %s", msg.Channel)
				}
			}

		case <-eb.ctx.Done():
			log.Println("[EVENTBUS] Stopped receiving Redis messages")
			return
		}
	}
}

// Close closes the event bus and all subscriptions
func (eb *RedisEventBus) Close() error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return nil
	}

	eb.closed = true
	eb.cancel()

	// Close all handler channels
	for topic, handlers := range eb.subscribers {
		for _, ch := range handlers {
			close(ch)
		}
		log.Printf("[EVENTBUS] 🔌 Unsubscribed from topic: %s", topic)
	}

	// Close Redis pubsub
	if err := eb.pubsub.Close(); err != nil {
		log.Printf("[EVENTBUS] ⚠️  Error closing Redis pubsub: %v", err)
	}

	// Close Redis client
	if err := eb.client.Close(); err != nil {
		log.Printf("[EVENTBUS] ⚠️  Error closing Redis client: %v", err)
	}

	log.Println("[EVENTBUS] ✅ Redis EventBus closed")
	return nil
}

// GetSubscriberCount returns the number of subscribers for a topic
func (eb *RedisEventBus) GetSubscriberCount(topic string) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.subscribers[topic])
}

// Health returns the health status of Redis event bus
func (eb *RedisEventBus) Health() map[string]interface{} {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	totalSubscribers := 0
	for _, handlers := range eb.subscribers {
		totalSubscribers += len(handlers)
	}

	return map[string]interface{}{
		"status":            "healthy",
		"type":              "redis",
		"topics":            len(eb.subscribers),
		"total_subscribers": totalSubscribers,
		"note":              "Events are persisted in Redis",
	}
}
