package web

// truncateText returns at most maxLen runes of s, appending "…" if
// truncated. Keeps byte representation under control for preview text.
func truncateText(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "\u2026"
}
