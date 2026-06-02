package tui

import (
	"strings"
)

// NaturalCommandParser handles natural language command parsing
type NaturalCommandParser struct {
	commandMap map[string]string
	aliases    map[string][]string
}

// NewNaturalCommandParser creates a new natural language parser
func NewNaturalCommandParser() *NaturalCommandParser {
	parser := &NaturalCommandParser{
		commandMap: make(map[string]string),
		aliases:    make(map[string][]string),
	}

	parser.registerDefaultMappings()
	return parser
}

// registerDefaultMappings sets up the natural language to command mappings
func (p *NaturalCommandParser) registerDefaultMappings() {
	// Help commands
	p.addMapping("help", "/help")
	p.addMapping("show help", "/help")
	p.addMapping("what can you do", "/help")
	p.addMapping("commands", "/help")
	p.addMapping("show commands", "/help")

	// Clear commands
	p.addMapping("clear", "/clear")
	p.addMapping("clear conversation", "/clear")
	p.addMapping("clear history", "/clear")
	p.addMapping("clear messages", "/clear")
	p.addMapping("reset", "/clear")
	p.addMapping("start over", "/clear")
	p.addMapping("new conversation", "/clear")

	// Model commands
	p.addMapping("switch model", "/model")
	p.addMapping("change model", "/model")
	p.addMapping("use model", "/model")
	p.addMapping("model", "/model")
	p.addMapping("switch to", "/model")
	p.addMapping("change to", "/model")

	// Status commands
	p.addMapping("status", "/status")
	p.addMapping("show status", "/status")
	p.addMapping("info", "/status")
	p.addMapping("session info", "/status")

	// Cost commands
	p.addMapping("cost", "/cost")
	p.addMapping("show cost", "/cost")
	p.addMapping("usage", "/cost")
	p.addMapping("tokens", "/cost")
	p.addMapping("how much", "/cost")

	// Session commands
	p.addMapping("save", "/session save")
	p.addMapping("save session", "/session save")
	p.addMapping("load", "/session load")
	p.addMapping("load session", "/session load")
	p.addMapping("sessions", "/sessions")
	p.addMapping("list sessions", "/sessions")
	p.addMapping("show sessions", "/sessions")

	// Theme commands
	p.addMapping("theme", "/theme")
	p.addMapping("change theme", "/theme")
	p.addMapping("toggle theme", "/theme")
	p.addMapping("dark mode", "/theme")
	p.addMapping("light mode", "/theme")

	// Exit commands
	p.addMapping("exit", "/exit")
	p.addMapping("quit", "/exit")
	p.addMapping("bye", "/exit")
	p.addMapping("goodbye", "/exit")

	// Todo commands
	p.addMapping("todo", "/todo")
	p.addMapping("todos", "/todo")
	p.addMapping("show todos", "/todo")
}

// addMapping adds a natural language phrase to command mapping
func (p *NaturalCommandParser) addMapping(phrase, command string) {
	p.commandMap[strings.ToLower(phrase)] = command
}

// Parse attempts to parse natural language input into a command
// Returns (command, args, isNatural)
func (p *NaturalCommandParser) Parse(input string) (string, []string, bool) {
	trimmed := strings.TrimSpace(input)
	lower := strings.ToLower(trimmed)

	// If it starts with /, it's already a command
	if strings.HasPrefix(trimmed, "/") {
		parts := strings.Fields(trimmed)
		if len(parts) == 0 {
			return "", nil, false
		}
		return parts[0], parts[1:], false
	}

	// Try exact match first
	if cmd, ok := p.commandMap[lower]; ok {
		return cmd, nil, true
	}

	// Try prefix match with arguments
	for phrase, cmd := range p.commandMap {
		if strings.HasPrefix(lower, phrase+" ") {
			// Extract arguments after the phrase
			argsStr := strings.TrimSpace(trimmed[len(phrase):])
			args := strings.Fields(argsStr)
			return cmd, args, true
		}

		// Also try without space for single-word commands
		if strings.HasPrefix(lower, phrase) && len(lower) > len(phrase) {
			// Check if next char is not alphanumeric (word boundary)
			if len(lower) > len(phrase) {
				nextChar := lower[len(phrase)]
				if !isAlphanumeric(nextChar) {
					argsStr := strings.TrimSpace(trimmed[len(phrase):])
					args := strings.Fields(argsStr)
					return cmd, args, true
				}
			}
		}
	}

	// Not a recognized natural command
	return "", nil, false
}

// IsCommand checks if the input looks like a command (natural or slash)
func (p *NaturalCommandParser) IsCommand(input string) bool {
	cmd, _, _ := p.Parse(input)
	return cmd != ""
}

// GetSuggestions returns command suggestions for partial input
func (p *NaturalCommandParser) GetSuggestions(partial string) []string {
	lower := strings.ToLower(strings.TrimSpace(partial))
	if lower == "" {
		return nil
	}

	var suggestions []string
	seen := make(map[string]bool)

	for phrase := range p.commandMap {
		if strings.HasPrefix(phrase, lower) {
			// Get the command this maps to
			cmd := p.commandMap[phrase]
			if !seen[phrase] {
				suggestions = append(suggestions, phrase)
				seen[phrase] = true
			}
			// Also suggest the slash command
			if !seen[cmd] {
				suggestions = append(suggestions, cmd)
				seen[cmd] = true
			}
		}
	}

	return suggestions
}

// GetAllCommands returns all registered natural language phrases
func (p *NaturalCommandParser) GetAllCommands() []string {
	commands := make([]string, 0, len(p.commandMap))
	for phrase := range p.commandMap {
		commands = append(commands, phrase)
	}
	return commands
}

// isAlphanumeric checks if a byte is alphanumeric
func isAlphanumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

// FormatHelp returns a formatted help string showing natural language alternatives
func (p *NaturalCommandParser) FormatHelp() string {
	var b strings.Builder

	b.WriteString("Natural Language Commands:\n\n")
	b.WriteString("You can use natural language instead of slash commands:\n\n")

	// Group by command
	cmdGroups := make(map[string][]string)
	for phrase, cmd := range p.commandMap {
		cmdGroups[cmd] = append(cmdGroups[cmd], phrase)
	}

	// Common commands first
	commonCmds := []string{"/help", "/clear", "/model", "/status", "/cost"}

	for _, cmd := range commonCmds {
		if phrases, ok := cmdGroups[cmd]; ok {
			b.WriteString("  " + cmd + ":\n")
			for _, phrase := range phrases {
				b.WriteString("    - " + phrase + "\n")
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("Examples:\n")
	b.WriteString("  'help' or '/help'\n")
	b.WriteString("  'clear conversation' or '/clear'\n")
	b.WriteString("  'switch to gpt-4' or '/model gpt-4'\n")
	b.WriteString("  'show status' or '/status'\n")

	return b.String()
}
