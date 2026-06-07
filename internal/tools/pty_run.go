package tools

import (
	"bufio"
	"claw-code-go/internal/api"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/creack/pty"
)

// ptyDefaultTimeout caps the wall-clock time a single pty_run call spends
// waiting for output. Long-running daemons should be tested via the
// non-interactive pty_run with timeout=0 (returns the prompt after first
// newline) or via a separate watch command.
const ptyDefaultTimeout = 30 * time.Second

// ptyDefaultInputDelay is the gap between sending successive input strings
// to the PTY. Many interactive REPLs (python -i, node -i, gdb, sqlite3)
// need a brief pause after each line for the prompt to render before the
// next input is read.
const ptyDefaultInputDelay = 200 * time.Millisecond

// blockedPatterns contains dangerous command patterns that should never
// be executed. Patterns are matched case-insensitively against the full
// command string. Some patterns are exact strings, others are regexes.
var blockedPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^rm\s+-rf\s+/$`),
	regexp.MustCompile(`^rm\s+-r\s+/home`),
	regexp.MustCompile(`^rm\s+-fr\s+/var`),
	regexp.MustCompile(`^rm\s+--recursive\s+/`),
	regexp.MustCompile(`^rm\s+--force\s+/etc/passwd`),
	regexp.MustCompile(`^dd\s+if=/dev/zero\s+of=/dev/sda`),
	regexp.MustCompile(`^mkfs\.ext4\s+/dev/sda`),
	regexp.MustCompile(`^mkfs\s+/dev/sda`),
	regexp.MustCompile(`^>\s*/dev/sda$`),
	regexp.MustCompile(`^>/dev/sda$`),
	regexp.MustCompile(`:\(\)\s*\{\s*:\|:&\s*\};:`),
	regexp.MustCompile(`^chmod\s+777\s+/$`),
	regexp.MustCompile(`^chmod\s+-R\s+777\s+/`),
	regexp.MustCompile(`^chown\s+-R\s+root\s+/`),
	regexp.MustCompile(`^kill\s+-9\s+1$`),
	regexp.MustCompile(`^reboot$`),
	regexp.MustCompile(`^shutdown`),
	regexp.MustCompile(`^halt$`),
	regexp.MustCompile(`^poweroff$`),
	regexp.MustCompile(`^wget\s+.*\s+-O\s+/dev/sda`),
	regexp.MustCompile(`^curl\s+.*\s+/dev/sda`),
	regexp.MustCompile(`^ufw\s+disable$`),
	regexp.MustCompile(`^iptables\s+-F$`),
	regexp.MustCompile(`^iptables\s+--flush$`),
	regexp.MustCompile(`^systemctl\s+stop\s+sshd$`),
	regexp.MustCompile(`^systemctl\s+disable\s+firewalld$`),
}

// checkBlockedPattern returns an error if the command matches any
// blocked pattern. The check is case-insensitive.
func checkBlockedPattern(command string) error {
	lower := strings.ToLower(strings.TrimSpace(command))
	for _, re := range blockedPatterns {
		if re.MatchString(lower) {
			return fmt.Errorf("command blocked by safety pattern: %s", command)
		}
	}
	return nil
}

// PTYRunTool returns the tool definition for the pty_run tool. pty_run is
// used to drive interactive programs (skill REPLs, MCP debug shells, REPL
// prompts) where a plain bash invocation would hang waiting for stdin.
func PTYRunTool() api.Tool {
	return api.Tool{
		Name: "pty_run",
		Description: "Run a command in a pseudo-terminal (PTY) and drive it with a sequence of " +
			"stdin inputs. Returns the rendered terminal output (ANSI codes stripped). " +
			"Use this for interactive debugging of skills, MCP servers, and REPLs.",
		InputSchema: api.InputSchema{
			Type: "object",
			Properties: map[string]api.Property{
				"command": {
					Type:        "string",
					Description: "Command to run (e.g. \"python3 -i\", \"node -i\", \"sqlite3 :memory:\")",
				},
				"args": {
					Type:        "array",
					Description: "Optional command arguments (string array)",
				},
				"inputs": {
					Type:        "array",
					Description: "Optional list of strings to write to stdin in order, with a brief delay between them",
				},
				"timeout_sec": {
					Type:        "integer",
					Description: fmt.Sprintf("Total timeout in seconds (default %d, 0 = no timeout)", int(ptyDefaultTimeout.Seconds())),
				},
				"input_delay_ms": {
					Type:        "integer",
					Description: fmt.Sprintf("Delay between successive inputs in ms (default %d)", int(ptyDefaultInputDelay.Milliseconds())),
				},
			},
			Required: []string{"command"},
		},
	}
}

// ExecutePTYRun runs command under a PTY, optionally driving it with inputs,
// and returns the terminal output (ANSI escape codes stripped) as a string.
func ExecutePTYRun(input map[string]any) (string, error) {
	command, ok := input["command"].(string)
	if !ok || command == "" {
		return "", fmt.Errorf("pty_run: 'command' is required")
	}

	if err := checkBlockedPattern(command); err != nil {
		return "", err
	}

	var args []string
	if raw, ok := input["args"].([]any); ok {
		for _, v := range raw {
			if s, ok := v.(string); ok {
				args = append(args, s)
			}
		}
	}

	timeout := ptyDefaultTimeout
	if v, ok := input["timeout_sec"]; ok {
		switch n := v.(type) {
		case float64:
			timeout = time.Duration(n) * time.Second
		case int:
			timeout = time.Duration(n) * time.Second
		}
	}

	delay := ptyDefaultInputDelay
	if v, ok := input["input_delay_ms"]; ok {
		switch n := v.(type) {
		case float64:
			delay = time.Duration(n) * time.Millisecond
		case int:
			delay = time.Duration(n) * time.Millisecond
		}
	}

	var inputs []string
	if raw, ok := input["inputs"].([]any); ok {
		for _, v := range raw {
			if s, ok := v.(string); ok {
				inputs = append(inputs, s)
			}
		}
	}

	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()
	pt, err := pty.Start(cmd)
	if err != nil {
		return "", fmt.Errorf("pty_run: start pty: %w", err)
	}
	defer pt.Close()

	if len(inputs) > 0 {
		go func() {
			for _, line := range inputs {
				_, _ = pt.Write([]byte(line + "\n"))
				time.Sleep(delay)
			}
		}()
	}

	deadline := time.Now().Add(timeout)
	if timeout == 0 {
		deadline = time.Now().Add(1 << 62) // ~290 years — effectively no timeout
	}

	// Use a single reader goroutine with a buffered channel to avoid
	// per-rune goroutine creation (was ~10k goroutines per 10KB output).
	type readResult struct {
		r    rune
		rerr error
	}
	readCh := make(chan readResult, 256)
	go func() {
		reader := bufio.NewReader(pt)
		for {
			r, _, rerr := reader.ReadRune()
			select {
			case readCh <- readResult{r, rerr}:
			case <-time.After(500 * time.Millisecond):
				return // consumer gone or blocked
			}
			if rerr != nil {
				return
			}
		}
	}()

	var buf strings.Builder
readLoop:
	for {
		if time.Now().After(deadline) {
			_ = cmd.Process.Kill()
			buf.WriteString("\n[pty_run: timeout reached, process killed]\n")
			break
		}

		select {
		case res, ok := <-readCh:
			if !ok {
				break readLoop
			}
			if res.rerr != nil {
				if res.rerr != io.EOF {
					buf.WriteString(fmt.Sprintf("\n[pty_run: read error: %v]\n", res.rerr))
				}
				break readLoop
			}
			if res.r == '\r' {
				continue
			}
			if res.r == 0x1b {
				// Drop ANSI escape sequences (ESC [ ... letter).
				select {
				case peek, ok := <-readCh:
					if !ok {
						break readLoop
					}
					if peek.rerr != nil {
						break readLoop
					}
					// Consume until terminator (0x40-0x7e or BEL 0x07).
					for (peek.r < 0x40 || peek.r > 0x7e) && peek.r != 0x07 {
						select {
						case peek, ok = <-readCh:
							if !ok {
								break readLoop
							}
							if peek.rerr != nil {
								break readLoop
							}
						case <-time.After(500 * time.Millisecond):
							break readLoop
						}
					}
				case <-time.After(500 * time.Millisecond):
				}
				continue
			}
			buf.WriteRune(res.r)
		case <-time.After(500 * time.Millisecond):
			if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
				break readLoop
			}
		}
	}

	if err := cmd.Wait(); err != nil && !strings.Contains(err.Error(), "signal:") && !strings.Contains(err.Error(), "exit status") {
		return strings.TrimSpace(buf.String()), fmt.Errorf("pty_run: %w", err)
	}
	return strings.TrimSpace(buf.String()), nil
}
