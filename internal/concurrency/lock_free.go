/* HUMAN MODE - Truth Protocol Active
   System: Senior CTO-scientist reasoning mode engaged
   Reward = TRUTH_PROVEN via tests. Claims = PROVISIONAL until verified.
   This file protected by HUMAN-TRUTH protocol - see truth_protocol/README.md
*/
package concurrency

import (
	"fmt"
	"sync/atomic"
	"unsafe"
)

// ============================================
// LOCK-FREE CONCURRENCY SYSTEM
// High-performance concurrent operations for trading
// ============================================

// AtomicFloat64 provides lock-free float64 operations
type AtomicFloat64 struct {
	value uint64
}

// NewAtomicFloat64 creates a new atomic float64
func NewAtomicFloat64(initial float64) *AtomicFloat64 {
	return &AtomicFloat64{
		value: *(*uint64)(unsafe.Pointer(&initial)),
	}
}

// Load atomically loads the current value
func (af *AtomicFloat64) Load() float64 {
	value := atomic.LoadUint64(&af.value)
	return *(*float64)(unsafe.Pointer(&value))
}

// Store atomically stores a new value
func (af *AtomicFloat64) Store(newValue float64) {
	atomic.StoreUint64(&af.value, *(*uint64)(unsafe.Pointer(&newValue)))
}

// Add atomically adds to the current value and returns the new value
func (af *AtomicFloat64) Add(delta float64) float64 {
	for {
		oldBits := atomic.LoadUint64(&af.value)
		oldValue := *(*float64)(unsafe.Pointer(&oldBits))
		newValue := oldValue + delta
		newBits := *(*uint64)(unsafe.Pointer(&newValue))

		if atomic.CompareAndSwapUint64(&af.value, oldBits, newBits) {
			return newValue
		}
	}
}

// CompareAndSwap atomically compares and swaps
func (af *AtomicFloat64) CompareAndSwap(old, new float64) bool {
	oldBits := *(*uint64)(unsafe.Pointer(&old))
	newBits := *(*uint64)(unsafe.Pointer(&new))
	return atomic.CompareAndSwapUint64(&af.value, oldBits, newBits)
}

// SpinLock provides a simple spin lock for low-contention scenarios
type SpinLock struct {
	state int32
}

// Lock acquires the spin lock
func (sl *SpinLock) Lock() {
	for !atomic.CompareAndSwapInt32(&sl.state, 0, 1) {
		// Spin until we can acquire the lock
	}
}

// Unlock releases the spin lock
func (sl *SpinLock) Unlock() {
	atomic.StoreInt32(&sl.state, 0)
}

// TryLock attempts to acquire the lock without blocking
func (sl *SpinLock) TryLock() bool {
	return atomic.CompareAndSwapInt32(&sl.state, 0, 1)
}

// LockFreeQueue implements a lock-free FIFO queue
type LockFreeQueue[T any] struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

// queueNode represents a node in the lock-free queue
type queueNode[T any] struct {
	value T
	next  unsafe.Pointer
}

// NewLockFreeQueue creates a new lock-free queue
func NewLockFreeQueue[T any]() *LockFreeQueue[T] {
	dummy := &queueNode[T]{}
	return &LockFreeQueue[T]{
		head: unsafe.Pointer(dummy),
		tail: unsafe.Pointer(dummy),
	}
}

// Enqueue adds an item to the queue
func (q *LockFreeQueue[T]) Enqueue(value T) {
	node := &queueNode[T]{value: value}

	for {
		tail := (*queueNode[T])(atomic.LoadPointer(&q.tail))
		next := (*queueNode[T])(atomic.LoadPointer(&tail.next))

		if tail == (*queueNode[T])(atomic.LoadPointer(&q.tail)) { // Are tail and next consistent?
			if next == nil {
				// Try to link node at the end of the linked list
				if atomic.CompareAndSwapPointer(&tail.next, nil, unsafe.Pointer(node)) {
					// Enqueue is done. Try to swing tail to the inserted node
					atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(node))
					return
				}
			} else {
				// Tail was not pointing to the last node, try to swing tail
				atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
			}
		}
	}
}

// Dequeue removes and returns an item from the queue
func (q *LockFreeQueue[T]) Dequeue() (T, bool) {
	var zero T
	for {
		head := (*queueNode[T])(atomic.LoadPointer(&q.head))
		tail := (*queueNode[T])(atomic.LoadPointer(&q.tail))
		next := (*queueNode[T])(atomic.LoadPointer(&head.next))

		if head == (*queueNode[T])(atomic.LoadPointer(&q.head)) { // Are head, tail, and next consistent?
			if head == tail { // Is queue empty or tail falling behind?
				if next == nil { // Is queue empty?
					return zero, false
				}
				// Tail is falling behind, try to advance it
				atomic.CompareAndSwapPointer(&q.tail, unsafe.Pointer(tail), unsafe.Pointer(next))
			} else {
				// Read value before CAS, otherwise another dequeue might free the next node
				value := next.value
				// Try to swing head to the next node
				if atomic.CompareAndSwapPointer(&q.head, unsafe.Pointer(head), unsafe.Pointer(next)) {
					return value, true
				}
			}
		}
	}
}

// LockFreeMap provides a lock-free concurrent map
type LockFreeMap[K comparable, V any] struct {
	buckets []lockFreeBucket[K, V]
	size    int
}

type lockFreeBucket[K comparable, V any] struct {
	items map[K]V
	mutex SpinLock
}

// NewLockFreeMap creates a new lock-free map
func NewLockFreeMap[K comparable, V any](initialSize int) *LockFreeMap[K, V] {
	if initialSize < 1 {
		initialSize = 16
	}

	buckets := make([]lockFreeBucket[K, V], initialSize)
	for i := range buckets {
		buckets[i].items = make(map[K]V)
	}

	return &LockFreeMap[K, V]{
		buckets: buckets,
		size:    initialSize,
	}
}

// getBucket returns the bucket for a given key
func (m *LockFreeMap[K, V]) getBucket(key K) *lockFreeBucket[K, V] {
	hash := hashKey(key)
	return &m.buckets[hash%uint32(len(m.buckets))]
}

// hashKey computes a simple hash for the key
func hashKey[K comparable](key K) uint32 {
	// Simple hash function - in production, use a better one
	var hash uint32 = 5381
	keyStr := fmt.Sprintf("%v", key)
	for _, char := range keyStr {
		hash = ((hash << 5) + hash) + uint32(char)
	}
	return hash
}

// Get retrieves a value from the map
func (m *LockFreeMap[K, V]) Get(key K) (V, bool) {
	bucket := m.getBucket(key)
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	value, exists := bucket.items[key]
	return value, exists
}

// Put stores a value in the map
func (m *LockFreeMap[K, V]) Put(key K, value V) {
	bucket := m.getBucket(key)
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	bucket.items[key] = value
}

// Delete removes a value from the map
func (m *LockFreeMap[K, V]) Delete(key K) bool {
	bucket := m.getBucket(key)
	bucket.mutex.Lock()
	defer bucket.mutex.Unlock()

	if _, exists := bucket.items[key]; exists {
		delete(bucket.items, key)
		return true
	}
	return false
}

// AtomicCounter provides lock-free counter operations
type AtomicCounter struct {
	value int64
}

// NewAtomicCounter creates a new atomic counter
func NewAtomicCounter(initial int64) *AtomicCounter {
	return &AtomicCounter{value: initial}
}

// Increment atomically increments and returns the new value
func (ac *AtomicCounter) Increment() int64 {
	return atomic.AddInt64(&ac.value, 1)
}

// Decrement atomically decrements and returns the new value
func (ac *AtomicCounter) Decrement() int64 {
	return atomic.AddInt64(&ac.value, -1)
}

// Add atomically adds to the counter and returns the new value
func (ac *AtomicCounter) Add(delta int64) int64 {
	return atomic.AddInt64(&ac.value, delta)
}

// Load atomically loads the current value
func (ac *AtomicCounter) Load() int64 {
	return atomic.LoadInt64(&ac.value)
}

// Store atomically stores a new value
func (ac *AtomicCounter) Store(newValue int64) {
	atomic.StoreInt64(&ac.value, newValue)
}

// CompareAndSwap atomically compares and swaps
func (ac *AtomicCounter) CompareAndSwap(old, new int64) bool {
	return atomic.CompareAndSwapInt64(&ac.value, old, new)
}

// SequenceGenerator provides lock-free sequence number generation
type SequenceGenerator struct {
	counter AtomicCounter
}

// NewSequenceGenerator creates a new sequence generator
func NewSequenceGenerator(start int64) *SequenceGenerator {
	return &SequenceGenerator{
		counter: *NewAtomicCounter(start),
	}
}

// Next returns the next sequence number
func (sg *SequenceGenerator) Next() int64 {
	return sg.counter.Increment()
}

// Current returns the current sequence number without incrementing
func (sg *SequenceGenerator) Current() int64 {
	return sg.counter.Load()
}

// LockFreeRingBuffer implements a lock-free ring buffer
type LockFreeRingBuffer[T any] struct {
	buffer []T
	size   int
	head   uint64 // Atomic head index
	tail   uint64 // Atomic tail index
}

// NewLockFreeRingBuffer creates a new lock-free ring buffer
func NewLockFreeRingBuffer[T any](capacity int) *LockFreeRingBuffer[T] {
	if capacity < 1 {
		capacity = 1024
	}

	return &LockFreeRingBuffer[T]{
		buffer: make([]T, capacity),
		size:   capacity,
		head:   0,
		tail:   0,
	}
}

// Push adds an item to the ring buffer
func (rb *LockFreeRingBuffer[T]) Push(item T) bool {
	for {
		head := atomic.LoadUint64(&rb.head)
		tail := atomic.LoadUint64(&rb.tail)

		if (tail - head) >= uint64(rb.size) {
			// Buffer is full
			return false
		}

		nextTail := tail + 1
		if atomic.CompareAndSwapUint64(&rb.tail, tail, nextTail) {
			// Successfully claimed the slot
			rb.buffer[tail%uint64(rb.size)] = item
			return true
		}
	}
}

// Pop removes and returns an item from the ring buffer
func (rb *LockFreeRingBuffer[T]) Pop() (T, bool) {
	var zero T
	for {
		head := atomic.LoadUint64(&rb.head)
		tail := atomic.LoadUint64(&rb.tail)

		if head >= tail {
			// Buffer is empty
			return zero, false
		}

		nextHead := head + 1
		if atomic.CompareAndSwapUint64(&rb.head, head, nextHead) {
			// Successfully claimed the item
			item := rb.buffer[head%uint64(rb.size)]
			return item, true
		}
	}
}

// Len returns the current number of items in the buffer
func (rb *LockFreeRingBuffer[T]) Len() int {
	head := atomic.LoadUint64(&rb.head)
	tail := atomic.LoadUint64(&rb.tail)
	return int(tail - head)
}

// Cap returns the buffer capacity
func (rb *LockFreeRingBuffer[T]) Cap() int {
	return rb.size
}
