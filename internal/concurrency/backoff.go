package concurrency

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ============================================
// EXPONENTIAL BACKOFF RECOVERY SYSTEM
// Fault tolerance and retry logic for trading operations
// ============================================

// BackoffConfig defines the configuration for exponential backoff
type BackoffConfig struct {
	InitialDelay time.Duration // Starting delay
	MaxDelay     time.Duration // Maximum delay
	Multiplier   float64       // Delay multiplier
	Jitter       bool          // Add random jitter
	MaxRetries   int           // Maximum number of retries (-1 for unlimited)
}

// DefaultBackoffConfig returns a sensible default configuration
func DefaultBackoffConfig() BackoffConfig {
	return BackoffConfig{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
		MaxRetries:   10,
	}
}

// ExponentialBackoff implements exponential backoff with jitter
type ExponentialBackoff struct {
	config     BackoffConfig
	attempts   int
	lastDelay  time.Duration
	totalDelay time.Duration
}

// NewExponentialBackoff creates a new exponential backoff instance
func NewExponentialBackoff(config BackoffConfig) *ExponentialBackoff {
	return &ExponentialBackoff{
		config:    config,
		attempts:  0,
		lastDelay: config.InitialDelay,
	}
}

// Reset resets the backoff state
func (eb *ExponentialBackoff) Reset() {
	eb.attempts = 0
	eb.lastDelay = eb.config.InitialDelay
	eb.totalDelay = 0
}

// NextDelay calculates the next delay duration
func (eb *ExponentialBackoff) NextDelay() time.Duration {
	if eb.config.MaxRetries >= 0 && eb.attempts >= eb.config.MaxRetries {
		return 0 // No more retries
	}

	delay := eb.lastDelay

	// Apply jitter if enabled
	if eb.config.Jitter {
		// Add random jitter of ±25%
		jitterFactor := 0.75 + rand.Float64()*0.5 // 0.75 to 1.25
		delay = time.Duration(float64(delay) * jitterFactor)
	}

	// Cap at max delay
	if delay > eb.config.MaxDelay {
		delay = eb.config.MaxDelay
	}

	// Calculate next delay for next attempt
	eb.lastDelay = time.Duration(float64(eb.lastDelay) * eb.config.Multiplier)
	if eb.lastDelay > eb.config.MaxDelay {
		eb.lastDelay = eb.config.MaxDelay
	}

	eb.attempts++
	eb.totalDelay += delay

	return delay
}

// Attempts returns the number of attempts made
func (eb *ExponentialBackoff) Attempts() int {
	return eb.attempts
}

// TotalDelay returns the total delay accumulated
func (eb *ExponentialBackoff) TotalDelay() time.Duration {
	return eb.totalDelay
}

// ShouldRetry returns true if another retry should be attempted
func (eb *ExponentialBackoff) ShouldRetry() bool {
	if eb.config.MaxRetries < 0 {
		return true // Unlimited retries
	}
	return eb.attempts < eb.config.MaxRetries
}

// CircuitBreaker implements a circuit breaker pattern
type CircuitBreaker struct {
	name         string
	state        CircuitState
	failures     int
	lastFailTime time.Time
	successes    int
	config       CircuitBreakerConfig
}

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreakerConfig defines circuit breaker configuration
type CircuitBreakerConfig struct {
	Name             string
	FailureThreshold int           // Failures before opening
	RecoveryTimeout  time.Duration // Time to wait before trying half-open
	SuccessThreshold int           // Successes needed to close from half-open
	Timeout          time.Duration // Request timeout
	ExpectedFailures []string      // Expected failure messages (don't count as failures)
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5
	}
	if config.RecoveryTimeout == 0 {
		config.RecoveryTimeout = 60 * time.Second
	}
	if config.SuccessThreshold == 0 {
		config.SuccessThreshold = 3
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	return &CircuitBreaker{
		name:   config.Name,
		state:  StateClosed,
		config: config,
	}
}

// Call executes a function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker %s is open", cb.name)
	}

	err := fn()
	cb.recordResult(err)

	return err
}

// canExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) canExecute() bool {
	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailTime) >= cb.config.RecoveryTimeout {
			cb.state = StateHalfOpen
			cb.successes = 0
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// recordResult records the result of a call
func (cb *CircuitBreaker) recordResult(err error) {
	isFailure := err != nil

	// Check if this is an expected failure
	if isFailure && cb.isExpectedFailure(err) {
		isFailure = false
	}

	switch cb.state {
	case StateClosed:
		if isFailure {
			cb.failures++
			cb.lastFailTime = time.Now()
			if cb.failures >= cb.config.FailureThreshold {
				cb.state = StateOpen
			}
		} else {
			cb.failures = 0 // Reset on success
		}

	case StateHalfOpen:
		if isFailure {
			cb.state = StateOpen
			cb.failures++
			cb.lastFailTime = time.Now()
		} else {
			cb.successes++
			if cb.successes >= cb.config.SuccessThreshold {
				cb.state = StateClosed
				cb.failures = 0
				cb.successes = 0
			}
		}
	}
}

// isExpectedFailure checks if an error is expected and shouldn't count as a failure
func (cb *CircuitBreaker) isExpectedFailure(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	for _, expected := range cb.config.ExpectedFailures {
		if strings.Contains(errMsg, expected) {
			return true
		}
	}

	return false
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitState {
	return cb.state
}

// Stats returns circuit breaker statistics
func (cb *CircuitBreaker) Stats() map[string]interface{} {
	return map[string]interface{}{
		"name":              cb.name,
		"state":             cb.stateString(),
		"failures":          cb.failures,
		"successes":         cb.successes,
		"last_failure":      cb.lastFailTime,
		"failure_threshold": cb.config.FailureThreshold,
		"recovery_timeout":  cb.config.RecoveryTimeout,
		"success_threshold": cb.config.SuccessThreshold,
	}
}

func (cb *CircuitBreaker) stateString() string {
	switch cb.state {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// RetryWithBackoff executes a function with exponential backoff retry
func RetryWithBackoff(fn func() error, config BackoffConfig) error {
	backoff := NewExponentialBackoff(config)

	var lastErr error
	for backoff.ShouldRetry() {
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err
		delay := backoff.NextDelay()

		if delay == 0 {
			break // No more retries
		}

		time.Sleep(delay)
	}

	return fmt.Errorf("operation failed after %d attempts: %w", backoff.Attempts(), lastErr)
}

// AdaptiveBackoff adjusts backoff based on system load
type AdaptiveBackoff struct {
	baseBackoff *ExponentialBackoff
	loadFactor  float64
	lastAdjust  time.Time
}

// NewAdaptiveBackoff creates a new adaptive backoff
func NewAdaptiveBackoff(config BackoffConfig) *AdaptiveBackoff {
	return &AdaptiveBackoff{
		baseBackoff: NewExponentialBackoff(config),
		loadFactor:  1.0,
		lastAdjust:  time.Now(),
	}
}

// NextDelay returns the next delay adjusted for system load
func (ab *AdaptiveBackoff) NextDelay() time.Duration {
	baseDelay := ab.baseBackoff.NextDelay()
	adjustedDelay := time.Duration(float64(baseDelay) * ab.loadFactor)

	// Cap at reasonable maximum
	maxDelay := 5 * time.Minute
	if adjustedDelay > maxDelay {
		adjustedDelay = maxDelay
	}

	return adjustedDelay
}

// AdjustLoadFactor adjusts the backoff based on system metrics
func (ab *AdaptiveBackoff) AdjustLoadFactor(cpuUsage, memoryUsage float64) {
	// Increase backoff when system is heavily loaded
	loadPressure := (cpuUsage + memoryUsage) / 200.0 // Normalize to 0-1

	// Adjust load factor (0.5 to 3.0)
	ab.loadFactor = 1.0 + (loadPressure * 2.0)
	if ab.loadFactor < 0.5 {
		ab.loadFactor = 0.5
	}
	if ab.loadFactor > 3.0 {
		ab.loadFactor = 3.0
	}

	ab.lastAdjust = time.Now()
}

// Reset resets the adaptive backoff
func (ab *AdaptiveBackoff) Reset() {
	ab.baseBackoff.Reset()
	ab.loadFactor = 1.0
}

// FailureRateTracker tracks failure rates for adaptive behavior
type FailureRateTracker struct {
	failures    *AtomicCounter
	totalCalls  *AtomicCounter
	windowStart time.Time
	windowSize  time.Duration
}

// NewFailureRateTracker creates a new failure rate tracker
func NewFailureRateTracker(windowSize time.Duration) *FailureRateTracker {
	return &FailureRateTracker{
		failures:    NewAtomicCounter(0),
		totalCalls:  NewAtomicCounter(0),
		windowStart: time.Now(),
		windowSize:  windowSize,
	}
}

// RecordCall records a call result
func (frt *FailureRateTracker) RecordCall(success bool) {
	frt.totalCalls.Increment()
	if !success {
		frt.failures.Increment()
	}

	// Reset window if needed
	if time.Since(frt.windowStart) >= frt.windowSize {
		frt.failures.Store(0)
		frt.totalCalls.Store(0)
		frt.windowStart = time.Now()
	}
}

// FailureRate returns the current failure rate (0.0 to 1.0)
func (frt *FailureRateTracker) FailureRate() float64 {
	total := frt.totalCalls.Load()
	if total == 0 {
		return 0.0
	}

	failures := frt.failures.Load()
	return float64(failures) / float64(total)
}

// ShouldThrottle returns true if the failure rate indicates throttling
func (frt *FailureRateTracker) ShouldThrottle(threshold float64) bool {
	return frt.FailureRate() > threshold
}
