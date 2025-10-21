/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package database

import (
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// WriteQueue provides resilient database writes with automatic retry
type WriteQueue struct {
	db    *gorm.DB
	queue []QueuedWrite
	mu    sync.Mutex

	maxQueueSize int
	retryDelay   time.Duration
}

// QueuedWrite represents a pending database operation
type QueuedWrite struct {
	Operation string      // "create", "update", "delete"
	Table     string      // Table name for logging
	Data      interface{} // The record to write
	Timestamp time.Time   // When it was queued
	Retries   int         // Number of retry attempts
}

// NewWriteQueue creates a new database write queue
func NewWriteQueue(db *gorm.DB, maxSize int) *WriteQueue {
	wq := &WriteQueue{
		db:           db,
		queue:        make([]QueuedWrite, 0),
		maxQueueSize: maxSize,
		retryDelay:   5 * time.Second,
	}

	// Start background processor
	go wq.processQueue()

	return wq
}

// Enqueue adds a write operation to the queue
func (wq *WriteQueue) Enqueue(operation, table string, data interface{}) error {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	// Check queue size limit
	if len(wq.queue) >= wq.maxQueueSize {
		log.Printf("[WRITEQUEUE][ERROR] Queue full (%d items), dropping oldest writes", wq.maxQueueSize)
		// Drop oldest 10% of queue
		dropCount := wq.maxQueueSize / 10
		wq.queue = wq.queue[dropCount:]
	}

	write := QueuedWrite{
		Operation: operation,
		Table:     table,
		Data:      data,
		Timestamp: time.Now(),
		Retries:   0,
	}

	wq.queue = append(wq.queue, write)
	log.Printf("[WRITEQUEUE][ENQUEUE] Queued %s for %s (queue size: %d)", operation, table, len(wq.queue))

	return nil
}

// processQueue continuously attempts to flush queued writes
func (wq *WriteQueue) processQueue() {
	ticker := time.NewTicker(wq.retryDelay)
	defer ticker.Stop()

	for range ticker.C {
		wq.flush()
	}
}

// flush attempts to write all queued operations
func (wq *WriteQueue) flush() {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	if len(wq.queue) == 0 {
		return
	}

	log.Printf("[WRITEQUEUE][FLUSH] Attempting to flush %d queued writes", len(wq.queue))

	// Test database connection
	sqlDB, err := wq.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		log.Printf("[WRITEQUEUE][WARN] Database still unavailable, keeping queue (%d items)", len(wq.queue))
		return
	}

	// Database is available - process queue
	processed := 0
	failed := []QueuedWrite{}

	for _, write := range wq.queue {
		var err error

		switch write.Operation {
		case "create":
			err = wq.db.Create(write.Data).Error
		case "update":
			err = wq.db.Save(write.Data).Error
		case "delete":
			err = wq.db.Delete(write.Data).Error
		default:
			log.Printf("[WRITEQUEUE][ERROR] Unknown operation: %s", write.Operation)
			continue
		}

		if err != nil {
			write.Retries++
			if write.Retries < 5 {
				// Retry later
				failed = append(failed, write)
				log.Printf("[WRITEQUEUE][RETRY] %s for %s failed (retry %d/5): %v",
					write.Operation, write.Table, write.Retries, err)
			} else {
				// Give up after 5 retries
				log.Printf("[WRITEQUEUE][DROP] Dropping %s for %s after 5 retries: %v",
					write.Operation, write.Table, err)
			}
		} else {
			processed++
			age := time.Since(write.Timestamp)
			log.Printf("[WRITEQUEUE][SUCCESS] Persisted %s for %s (age: %v)",
				write.Operation, write.Table, age.Round(time.Second))
		}
	}

	// Update queue with only failed writes
	wq.queue = failed

	if processed > 0 {
		log.Printf("[WRITEQUEUE][COMPLETE] Flushed %d writes, %d remaining in queue",
			processed, len(wq.queue))
	}
}

// Stats returns queue statistics
func (wq *WriteQueue) Stats() map[string]interface{} {
	wq.mu.Lock()
	defer wq.mu.Unlock()

	oldestAge := time.Duration(0)
	if len(wq.queue) > 0 {
		oldestAge = time.Since(wq.queue[0].Timestamp)
	}

	return map[string]interface{}{
		"queue_size":       len(wq.queue),
		"max_queue_size":   wq.maxQueueSize,
		"oldest_write_age": oldestAge.Round(time.Second).String(),
		"retry_delay":      wq.retryDelay.String(),
	}
}

// Size returns current queue size
func (wq *WriteQueue) Size() int {
	wq.mu.Lock()
	defer wq.mu.Unlock()
	return len(wq.queue)
}
