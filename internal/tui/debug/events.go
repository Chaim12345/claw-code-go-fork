package debug

import (
	"fmt"
	"sync"
	"time"
)

// EventType represents the type of debug event
type EventType string

const (
	EventStateChange    EventType = "state_change"
	EventToolCall       EventType = "tool_call"
	EventToolComplete   EventType = "tool_complete"
	EventStreamChunk    EventType = "stream_chunk"
	EventStreamComplete EventType = "stream_complete"
	EventStreamError    EventType = "stream_error"
	EventRender         EventType = "render"
	EventKeyPress       EventType = "key_press"
	EventResize         EventType = "resize"
	EventPermission     EventType = "permission"
	EventSession        EventType = "session"
	EventCompaction     EventType = "compaction"
	EventError          EventType = "error"
	EventWarning        EventType = "warning"
	EventInfo           EventType = "info"
)

// Event represents a single debug event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Duration  time.Duration          `json:"duration,omitempty"`
}

// Logger handles debug event logging
type Logger struct {
	mu       sync.RWMutex
	events   []Event
	maxSize  int
	enabled  bool
	filters  map[EventType]bool
	handlers []EventHandler
	counter  int
}

// EventHandler is called when an event is logged
type EventHandler func(Event)

// NewLogger creates a new debug logger
func NewLogger(maxSize int) *Logger {
	return &Logger{
		events:   make([]Event, 0, maxSize),
		maxSize:  maxSize,
		enabled:  true,
		filters:  make(map[EventType]bool),
		handlers: make([]EventHandler, 0),
	}
}

// Log adds a new event to the logger
func (l *Logger) Log(eventType EventType, message string, data map[string]interface{}) {
	if !l.enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if this event type is filtered
	if filtered, ok := l.filters[eventType]; ok && filtered {
		return
	}

	l.counter++
	event := Event{
		ID:        fmt.Sprintf("evt_%d", l.counter),
		Type:      eventType,
		Timestamp: time.Now(),
		Message:   message,
		Data:      data,
	}

	// Add to buffer (circular)
	if len(l.events) >= l.maxSize {
		l.events = l.events[1:]
	}
	l.events = append(l.events, event)

	// Call handlers
	for _, handler := range l.handlers {
		go handler(event)
	}
}

// LogWithDuration logs an event with a duration measurement
func (l *Logger) LogWithDuration(eventType EventType, message string, data map[string]interface{}, duration time.Duration) {
	if !l.enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if filtered, ok := l.filters[eventType]; ok && filtered {
		return
	}

	l.counter++
	event := Event{
		ID:        fmt.Sprintf("evt_%d", l.counter),
		Type:      eventType,
		Timestamp: time.Now(),
		Message:   message,
		Data:      data,
		Duration:  duration,
	}

	if len(l.events) >= l.maxSize {
		l.events = l.events[1:]
	}
	l.events = append(l.events, event)

	for _, handler := range l.handlers {
		go handler(event)
	}
}

// GetEvents returns a copy of all events
func (l *Logger) GetEvents() []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()

	events := make([]Event, len(l.events))
	copy(events, l.events)
	return events
}

// GetEventsByType returns events filtered by type
func (l *Logger) GetEventsByType(eventType EventType) []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()

	filtered := make([]Event, 0)
	for _, e := range l.events {
		if e.Type == eventType {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// GetRecentEvents returns the N most recent events
func (l *Logger) GetRecentEvents(n int) []Event {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if n > len(l.events) {
		n = len(l.events)
	}

	events := make([]Event, n)
	copy(events, l.events[len(l.events)-n:])
	return events
}

// Clear removes all events
func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.events = make([]Event, 0, l.maxSize)
	l.counter = 0
}

// Enable enables event logging
func (l *Logger) Enable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = true
}

// Disable disables event logging
func (l *Logger) Disable() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.enabled = false
}

// IsEnabled returns whether logging is enabled
func (l *Logger) IsEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.enabled
}

// SetFilter sets whether to filter out a specific event type
func (l *Logger) SetFilter(eventType EventType, filtered bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.filters[eventType] = filtered
}

// AddHandler adds an event handler
func (l *Logger) AddHandler(handler EventHandler) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.handlers = append(l.handlers, handler)
}

// GetStats returns statistics about logged events
func (l *Logger) GetStats() map[EventType]int {
	l.mu.RLock()
	defer l.mu.RUnlock()

	stats := make(map[EventType]int)
	for _, e := range l.events {
		stats[e.Type]++
	}
	return stats
}

// Global logger instance
var globalLogger = NewLogger(1000)

// Global convenience functions
func Log(eventType EventType, message string, data map[string]interface{}) {
	globalLogger.Log(eventType, message, data)
}

func LogWithDuration(eventType EventType, message string, data map[string]interface{}, duration time.Duration) {
	globalLogger.LogWithDuration(eventType, message, data, duration)
}

func GetEvents() []Event {
	return globalLogger.GetEvents()
}

func GetEventsByType(eventType EventType) []Event {
	return globalLogger.GetEventsByType(eventType)
}

func GetRecentEvents(n int) []Event {
	return globalLogger.GetRecentEvents(n)
}

func Clear() {
	globalLogger.Clear()
}

func Enable() {
	globalLogger.Enable()
}

func Disable() {
	globalLogger.Disable()
}

func IsEnabled() bool {
	return globalLogger.IsEnabled()
}

func SetFilter(eventType EventType, filtered bool) {
	globalLogger.SetFilter(eventType, filtered)
}

func AddHandler(handler EventHandler) {
	globalLogger.AddHandler(handler)
}

func GetStats() map[EventType]int {
	return globalLogger.GetStats()
}
