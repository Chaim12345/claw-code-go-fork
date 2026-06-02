package runtime

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"claw-code-go/internal/api"
)

// MessageImportance tracks the importance score of a message
type MessageImportance struct {
	Message api.Message
	Score   float64
	Index   int
	Reason  string
}

// HierarchicalLevel represents one level in the hierarchical compaction
type HierarchicalLevel struct {
	Name            string
	MessageCount    int
	CompressionRate float64
	Summary         string
	TokenCount      int
}

// HybridContextManager combines all three context management strategies
type HybridContextManager struct {
	config              *Config
	client              api.APIClient
	levels              []HierarchicalLevel
	importanceThreshold float64
}

// NewHybridContextManager creates a new hybrid context manager
func NewHybridContextManager(cfg *Config, client api.APIClient) *HybridContextManager {
	return &HybridContextManager{
		config:              cfg,
		client:              client,
		importanceThreshold: 5.0,
	}
}

// CalculateImportance scores a message based on multiple factors
func CalculateImportance(msg api.Message, index int, total int) float64 {
	score := 0.0

	// Recency bonus (exponential decay)
	recencyFactor := float64(index) / float64(total)
	score += recencyFactor * 10.0

	// Content type scoring
	for _, block := range msg.Content {
		switch block.Type {
		case "tool_use":
			score += 15.0 // Tool calls are critical
		case "tool_result":
			// Check result importance
			if containsError(block) {
				score += 20.0 // Errors are very important
			} else if containsFileModification(block) {
				score += 12.0 // File changes are important
			} else {
				score += 5.0 // Regular results
			}
		case "text":
			// Analyze text content
			if containsDecision(block.Text) {
				score += 10.0
			}
			if containsQuestion(block.Text) {
				score += 8.0
			}
			if containsError(block) {
				score += 15.0
			}
			// Length penalty for very long messages
			if len(block.Text) > 5000 {
				score -= 2.0
			}
		}
	}

	// Role bonus
	if msg.Role == "user" {
		score += 5.0 // User messages are important
	}

	return score
}

// containsError checks if a content block contains error indicators
func containsError(block api.ContentBlock) bool {
	text := strings.ToLower(block.Text)
	for _, inner := range block.Content {
		text += " " + strings.ToLower(inner.Text)
	}
	return strings.Contains(text, "error") ||
		strings.Contains(text, "failed") ||
		strings.Contains(text, "exception") ||
		strings.Contains(text, "panic")
}

// containsFileModification checks if a tool result indicates file changes
func containsFileModification(block api.ContentBlock) bool {
	for _, inner := range block.Content {
		text := strings.ToLower(inner.Text)
		if strings.Contains(text, "wrote") ||
			strings.Contains(text, "modified") ||
			strings.Contains(text, "created") ||
			strings.Contains(text, "updated") ||
			strings.Contains(text, "deleted") {
			return true
		}
	}
	return false
}

// containsDecision checks if text contains decision-making language
func containsDecision(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "decided") ||
		strings.Contains(lower, "chose") ||
		strings.Contains(lower, "selected") ||
		strings.Contains(lower, "will use") ||
		strings.Contains(lower, "going to")
}

// containsQuestion checks if text contains questions
func containsQuestion(text string) bool {
	return strings.Contains(text, "?")
}

// PruneByImportance removes low-importance messages to reach target token count
func PruneByImportance(messages []api.Message, targetTokens int) []api.Message {
	if len(messages) == 0 {
		return messages
	}

	// Score all messages
	scored := make([]MessageImportance, len(messages))
	for i, msg := range messages {
		scored[i] = MessageImportance{
			Message: msg,
			Score:   CalculateImportance(msg, i, len(messages)),
			Index:   i,
		}
	}

	// Sort by score (descending)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// Keep messages until we hit target
	kept := make([]MessageImportance, 0)
	currentTokens := 0

	for _, sm := range scored {
		msgTokens := EstimateTokens([]api.Message{sm.Message})
		if currentTokens+msgTokens <= targetTokens {
			kept = append(kept, sm)
			currentTokens += msgTokens
		}
	}

	// Sort back to original order
	sort.Slice(kept, func(i, j int) bool {
		return kept[i].Index < kept[j].Index
	})

	// Extract messages
	result := make([]api.Message, len(kept))
	for i, sm := range kept {
		result[i] = sm.Message
	}

	return result
}

// summarizeSegment creates a summary of a message segment.
// Uses the caller-provided client (which the Manage call site already
// wraps with NewCompactClient to guarantee high-capacity summarization
// even when the active model has a small context window, e.g. DeepSeek
// Expert at ~40k tokens).
func summarizeSegment(ctx context.Context, client api.APIClient, messages []api.Message, maxTokens int) (string, error) {
	if len(messages) == 0 {
		return "", nil
	}

	transcript := buildTranscript(messages)

	req := api.CreateMessageRequest{
		Model:     "", // empty = let the provider choose its own default (no hardcoded Anthropic model)
		MaxTokens: maxTokens,
		Messages: []api.Message{
			{
				Role: "user",
				Content: []api.ContentBlock{
					{Type: "text", Text: transcript},
				},
			},
		},
		Stream: true,
	}

	ch, err := client.StreamResponse(ctx, req)
	if err != nil {
		return "", fmt.Errorf("summarize segment: stream response: %w", err)
	}

	return collectStreamText(ch)
}

// CompactHierarchical creates multi-level summaries of the conversation
func (h *HybridContextManager) CompactHierarchical(ctx context.Context, messages []api.Message) error {
	if len(messages) == 0 {
		return nil
	}

	// Define compression levels
	levelDefs := []struct {
		name      string
		keep      int
		compress  float64
		maxTokens int
	}{
		{"recent", 10, 1.0, 0},      // Keep verbatim
		{"detailed", 20, 0.5, 1024}, // 50% compression
		{"summary", 30, 0.25, 512},  // 75% compression
		{"overview", -1, 0.1, 256},  // 90% compression (rest)
	}

	h.levels = make([]HierarchicalLevel, 0, len(levelDefs))
	processed := 0

	for _, level := range levelDefs {
		start := processed
		end := start + level.keep
		if level.keep < 0 {
			end = len(messages) // All remaining
		}
		if end > len(messages) {
			end = len(messages)
		}

		if start >= end {
			continue
		}

		segment := messages[start:end]

		if level.compress >= 1.0 {
			// Keep verbatim
			h.levels = append(h.levels, HierarchicalLevel{
				Name:         level.name,
				MessageCount: len(segment),
				Summary:      "", // Not summarized
				TokenCount:   EstimateTokens(segment),
			})
		} else {
			// Summarize
			summary, err := summarizeSegment(ctx, h.client, segment, level.maxTokens)
			if err != nil {
				return fmt.Errorf("compact level %s: %w", level.name, err)
			}

			h.levels = append(h.levels, HierarchicalLevel{
				Name:            level.name,
				MessageCount:    len(segment),
				CompressionRate: level.compress,
				Summary:         summary,
				TokenCount:      len(summary) / charsPerToken,
			})
		}

		processed = end
	}

	return nil
}

// BuildHierarchicalPrompt creates a system prompt with hierarchical context
func (h *HybridContextManager) BuildHierarchicalPrompt() string {
	if len(h.levels) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<hierarchical_context>\n")
	sb.WriteString("The conversation history has been organized into multiple levels for efficient context management:\n\n")

	// Iterate from oldest to newest
	for i := len(h.levels) - 1; i >= 0; i-- {
		level := h.levels[i]
		if level.Summary != "" {
			fmt.Fprintf(&sb, "## %s Context (%d messages, ~%d tokens)\n",
				strings.Title(level.Name), level.MessageCount, level.TokenCount)
			sb.WriteString(level.Summary)
			sb.WriteString("\n\n")
		}
	}

	sb.WriteString("</hierarchical_context>\n")
	return sb.String()
}

// Manage applies the hybrid context management strategy
func (h *HybridContextManager) Manage(ctx context.Context, messages []api.Message) ([]api.Message, string, error) {
	currentTokens := EstimateTokens(messages)
	maxTokens := h.config.CompactionMaxInputTokens
	if maxTokens <= 0 {
		maxTokens = h.config.MaxTokens
	}
	threshold := int(float64(maxTokens) * h.config.CompactionThreshold)

	if currentTokens < threshold {
		return messages, "", nil // No action needed
	}

	// Step 1: Prune low-importance messages (Path 3)
	targetAfterPrune := int(float64(maxTokens) * 0.6) // 60% of max
	pruned := PruneByImportance(messages, targetAfterPrune)

	prunedTokens := EstimateTokens(pruned)
	if prunedTokens < threshold {
		return pruned, "", nil // Pruning was enough
	}

	// Step 2: Apply hierarchical summarization (Path 2)
	if err := h.CompactHierarchical(ctx, pruned); err != nil {
		return pruned, "", fmt.Errorf("hierarchical compaction: %w", err)
	}

	// Step 3: Build final context with summaries
	systemPrompt := h.BuildHierarchicalPrompt()

	// Keep only recent messages (Level 0)
	var recentMessages []api.Message
	if len(h.levels) > 0 {
		recentLevel := h.levels[0]
		if recentLevel.MessageCount > 0 && recentLevel.MessageCount <= len(pruned) {
			recentMessages = pruned[len(pruned)-recentLevel.MessageCount:]
		} else {
			// Fallback: keep last 10 messages
			keepCount := 10
			if keepCount > len(pruned) {
				keepCount = len(pruned)
			}
			recentMessages = pruned[len(pruned)-keepCount:]
		}
	} else {
		// Fallback: keep last 10 messages
		keepCount := 10
		if keepCount > len(pruned) {
			keepCount = len(pruned)
		}
		recentMessages = pruned[len(pruned)-keepCount:]
	}

	return recentMessages, systemPrompt, nil
}

// AdaptiveWindowSize calculates the optimal number of recent messages to keep
func AdaptiveWindowSize(messages []api.Message, minKeep, maxKeep int) int {
	if len(messages) <= minKeep {
		return len(messages)
	}

	// Count critical messages in the recent window
	criticalCount := 0
	checkWindow := maxKeep
	if checkWindow > len(messages) {
		checkWindow = len(messages)
	}

	for i := len(messages) - 1; i >= len(messages)-checkWindow && i >= 0; i-- {
		if isCriticalMessage(messages[i]) {
			criticalCount++
		}
	}

	// Keep more messages if many are critical
	keep := minKeep + (criticalCount / 2)
	if keep < minKeep {
		keep = minKeep
	}
	if keep > maxKeep {
		keep = maxKeep
	}

	return keep
}

// isCriticalMessage determines if a message is critical to keep
func isCriticalMessage(msg api.Message) bool {
	for _, block := range msg.Content {
		switch block.Type {
		case "tool_use", "tool_result":
			return true
		case "text":
			// Check for error indicators
			lower := strings.ToLower(block.Text)
			if strings.Contains(lower, "error") ||
				strings.Contains(lower, "failed") ||
				strings.Contains(lower, "exception") {
				return true
			}
		}
	}
	return false
}
