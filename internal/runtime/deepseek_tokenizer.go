package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sugarme/tokenizer"
)

// deepseekTokenizer wraps the sugarme/tokenizer for accurate DeepSeek
// V3/V4 token counting. It loads the official tokenizer.json from
// HuggingFace (deepseek-ai/DeepSeek-V3) and provides a simple
// CountTokens API that replaces the old chars/4 heuristic.
type deepseekTokenizer struct {
	mu   sync.Mutex
	tk   *tokenizer.Tokenizer
	err  error
	init bool
}

var dsTokenizer = &deepseekTokenizer{}

// tokenizerPath returns the path to the bundled tokenizer.json file.
func tokenizerPath() string {
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		candidate := filepath.Join(dir, "internal", "runtime", "tokenizer.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	// Fall back to source tree relative path.
	return filepath.Join("internal", "runtime", "tokenizer.json")
}

// initTokenizer loads the DeepSeek tokenizer once.
func (dt *deepseekTokenizer) initTokenizer() {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	if dt.init {
		return
	}
	dt.init = true

	path := tokenizerPath()
	if _, statErr := os.Stat(path); statErr != nil {
		dt.err = statErr
		return
	}

	tk := tokenizer.NewTokenizerFromFile(path)
	if tk == nil {
		dt.err = fmt.Errorf("failed to load tokenizer from %s", path)
		return
	}
	dt.tk = tk
}

// CountTokens returns the number of tokens in the given text using
// the DeepSeek V3 tokenizer. Falls back to chars/4 if the tokenizer
// is unavailable.
func (dt *deepseekTokenizer) CountTokens(text string) int {
	dt.mu.Lock()
	init := dt.init
	dt.mu.Unlock()

	if !init {
		dt.initTokenizer()
	}

	dt.mu.Lock()
	tk := dt.tk
	dt.mu.Unlock()

	if tk == nil {
		if len(text) == 0 {
			return 0
		}
		return (len(text) + 3) / 4
	}

	enc, err := tk.EncodeSingle(text, false)
	if err != nil {
		return (len(text) + 3) / 4
	}
	return len(enc.GetIds())
}


