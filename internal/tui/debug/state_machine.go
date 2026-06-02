package debug

import (
	"fmt"
	"sync"
	"time"
)

// State represents a TUI state
type State string

const (
	StateInput         State = "input"
	StateBusy          State = "busy"
	StatePicker        State = "picker"
	StateHelp          State = "help"
	StatePermission    State = "permission"
	StateLoginProvider State = "login_provider"
	StateLoginMethod   State = "login_method"
	StateLoginAPIKey   State = "login_api_key"
	StateLoginOAuth    State = "login_oauth"
	StateAskUser       State = "ask_user"
	StatePalette       State = "palette"
	StateSessionPicker State = "session_picker"
	StateMention       State = "mention"
	StateTodoPanel     State = "todo_panel"
	StateDebugPanel    State = "debug_panel"
)

// StateTransition represents a state change
type StateTransition struct {
	From      State
	To        State
	Timestamp time.Time
	Reason    string
	Duration  time.Duration
}

// StateMachine tracks state transitions and validates them
type StateMachine struct {
	mu               sync.RWMutex
	current          State
	previous         State
	history          []StateTransition
	maxHistory       int
	stateEnterTime   time.Time
	validTransitions map[State][]State
	metrics          map[State]*StateMetrics
}

// StateMetrics tracks metrics for a specific state
type StateMetrics struct {
	EnterCount    int
	TotalDuration time.Duration
	AvgDuration   time.Duration
	MaxDuration   time.Duration
	MinDuration   time.Duration
}

// NewStateMachine creates a new state machine tracker
func NewStateMachine(initialState State, maxHistory int) *StateMachine {
	sm := &StateMachine{
		current:          initialState,
		previous:         initialState,
		history:          make([]StateTransition, 0, maxHistory),
		maxHistory:       maxHistory,
		stateEnterTime:   time.Now(),
		validTransitions: make(map[State][]State),
		metrics:          make(map[State]*StateMetrics),
	}

	// Define valid state transitions
	sm.defineTransitions()

	return sm
}

// defineTransitions sets up the valid state transition graph
func (sm *StateMachine) defineTransitions() {
	// Input state can transition to most states
	sm.validTransitions[StateInput] = []State{
		StateBusy, StatePicker, StateHelp, StatePalette,
		StateSessionPicker, StateMention, StateTodoPanel,
		StateDebugPanel, StateLoginProvider,
	}

	// Busy state can transition to input, permission, or ask_user
	sm.validTransitions[StateBusy] = []State{
		StateInput, StatePermission, StateAskUser,
	}

	// Most overlay states can return to input or busy
	sm.validTransitions[StatePicker] = []State{StateInput}
	sm.validTransitions[StateHelp] = []State{StateInput}
	sm.validTransitions[StatePalette] = []State{StateInput}
	sm.validTransitions[StateSessionPicker] = []State{StateInput}
	sm.validTransitions[StateMention] = []State{StateInput}
	sm.validTransitions[StateTodoPanel] = []State{StateInput}
	sm.validTransitions[StateDebugPanel] = []State{StateInput, StateBusy}

	// Permission and ask_user return to busy
	sm.validTransitions[StatePermission] = []State{StateBusy}
	sm.validTransitions[StateAskUser] = []State{StateBusy}

	// Login flow transitions
	sm.validTransitions[StateLoginProvider] = []State{StateInput, StateLoginMethod, StateLoginAPIKey}
	sm.validTransitions[StateLoginMethod] = []State{StateLoginProvider, StateLoginAPIKey, StateLoginOAuth}
	sm.validTransitions[StateLoginAPIKey] = []State{StateInput}
	sm.validTransitions[StateLoginOAuth] = []State{StateInput}
}

// Transition changes the current state and logs the transition
func (sm *StateMachine) Transition(to State, reason string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	from := sm.current

	// Validate transition
	if !sm.isValidTransition(from, to) {
		err := fmt.Errorf("invalid state transition: %s -> %s", from, to)
		Log(EventError, err.Error(), map[string]interface{}{
			"from":   string(from),
			"to":     string(to),
			"reason": reason,
		})
		return err
	}

	// Calculate duration in previous state
	duration := time.Since(sm.stateEnterTime)

	// Update metrics
	sm.updateMetrics(from, duration)

	// Record transition
	transition := StateTransition{
		From:      from,
		To:        to,
		Timestamp: time.Now(),
		Reason:    reason,
		Duration:  duration,
	}

	if len(sm.history) >= sm.maxHistory {
		sm.history = sm.history[1:]
	}
	sm.history = append(sm.history, transition)

	// Update state
	sm.previous = from
	sm.current = to
	sm.stateEnterTime = time.Now()

	// Log the transition
	Log(EventStateChange, fmt.Sprintf("State: %s -> %s", from, to), map[string]interface{}{
		"from":     string(from),
		"to":       string(to),
		"reason":   reason,
		"duration": duration.String(),
	})

	return nil
}

// isValidTransition checks if a transition is allowed
func (sm *StateMachine) isValidTransition(from, to State) bool {
	// Same state is always valid
	if from == to {
		return true
	}

	validStates, ok := sm.validTransitions[from]
	if !ok {
		return false
	}

	for _, valid := range validStates {
		if valid == to {
			return true
		}
	}

	return false
}

// updateMetrics updates metrics for a state
func (sm *StateMachine) updateMetrics(state State, duration time.Duration) {
	metrics, ok := sm.metrics[state]
	if !ok {
		metrics = &StateMetrics{
			MinDuration: duration,
		}
		sm.metrics[state] = metrics
	}

	metrics.EnterCount++
	metrics.TotalDuration += duration

	if duration > metrics.MaxDuration {
		metrics.MaxDuration = duration
	}
	if duration < metrics.MinDuration || metrics.MinDuration == 0 {
		metrics.MinDuration = duration
	}

	metrics.AvgDuration = metrics.TotalDuration / time.Duration(metrics.EnterCount)
}

// GetCurrent returns the current state
func (sm *StateMachine) GetCurrent() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

// GetPrevious returns the previous state
func (sm *StateMachine) GetPrevious() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.previous
}

// GetHistory returns the state transition history
func (sm *StateMachine) GetHistory() []StateTransition {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	history := make([]StateTransition, len(sm.history))
	copy(history, sm.history)
	return history
}

// GetMetrics returns metrics for all states
func (sm *StateMachine) GetMetrics() map[State]*StateMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	metrics := make(map[State]*StateMetrics)
	for state, m := range sm.metrics {
		metricsCopy := *m
		metrics[state] = &metricsCopy
	}
	return metrics
}

// GetStateMetrics returns metrics for a specific state
func (sm *StateMachine) GetStateMetrics(state State) *StateMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if m, ok := sm.metrics[state]; ok {
		metricsCopy := *m
		return &metricsCopy
	}
	return nil
}

// GetTimeInCurrentState returns how long we've been in the current state
func (sm *StateMachine) GetTimeInCurrentState() time.Duration {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return time.Since(sm.stateEnterTime)
}

// Reset resets the state machine to initial state
func (sm *StateMachine) Reset(initialState State) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.current = initialState
	sm.previous = initialState
	sm.history = make([]StateTransition, 0, sm.maxHistory)
	sm.stateEnterTime = time.Now()
	sm.metrics = make(map[State]*StateMetrics)
}

// Global state machine instance
var globalStateMachine = NewStateMachine(StateInput, 100)

// Global convenience functions
func TransitionState(to State, reason string) error {
	return globalStateMachine.Transition(to, reason)
}

func GetCurrentState() State {
	return globalStateMachine.GetCurrent()
}

func GetPreviousState() State {
	return globalStateMachine.GetPrevious()
}

func GetStateHistory() []StateTransition {
	return globalStateMachine.GetHistory()
}

func GetStateMetrics() map[State]*StateMetrics {
	return globalStateMachine.GetMetrics()
}

func GetTimeInCurrentState() time.Duration {
	return globalStateMachine.GetTimeInCurrentState()
}

func ResetStateMachine(initialState State) {
	globalStateMachine.Reset(initialState)
}
