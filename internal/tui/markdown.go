package tui

import (
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
)

// mdRenderer caches a glamour renderer keyed by viewport width so we
// don't allocate a new one on every View() call.
var (
	mdRendererMu sync.Mutex
	mdRenderer   *glamour.TermRenderer
	mdLastWidth  int
)

// RenderMarkdown converts markdown text to terminal-ANSI styled output
// using the glamour "dark" or "light" style depending on the active theme.
// The output width is clamped to w so code blocks and tables don't overflow.
func RenderMarkdown(md string, w int) string {
	if md == "" {
		return ""
	}
	// Ensure a minimum width so glamour doesn't panic on 0.
	if w < 40 {
		w = 40
	}
	// glamour adds its own margins — allow the content to use most of the width.
	renderWidth := w - 2
	if renderWidth < 40 {
		renderWidth = 40
	}

	mdRendererMu.Lock()
	defer mdRendererMu.Unlock()

	if mdRenderer == nil || mdLastWidth != renderWidth {
		var style string
		if currentTheme == LightTheme {
			style = "light"
		} else {
			style = "dark"
		}
		r, err := glamour.NewTermRenderer(
			glamour.WithStylesFromJSONBytes([]byte(glamourTheme(style))),
			glamour.WithWordWrap(renderWidth),
		)
		if err != nil {
			// Fallback: return the raw markdown if glamour init fails.
			return md
		}
		mdRenderer = r
		mdLastWidth = renderWidth
	}

	out, err := mdRenderer.Render(md)
	if err != nil {
		return md
	}
	// glamour often appends a trailing newline. Trim one so we don't
	// inject extra blank lines into the viewport.
	return strings.TrimSuffix(out, "\n")
}

// RenderUserMarkdown is a convenience wrapper that prefixes the rendered
// output with the user label.
func RenderUserMarkdown(text string, width int) string {
	label := userLabelStyle.Render("You") + ": "
	if !looksLikeMarkdown(text) {
		return label + text + "\n\n"
	}
	return label + "\n" + RenderMarkdown(text, width) + "\n\n"
}

// RenderAssistantMarkdown is a convenience wrapper that prefixes the rendered
// output with the assistant label.
func RenderAssistantMarkdown(text string, width int) string {
	label := assistantLabelStyle.Render("Claude") + ": "
	if !looksLikeMarkdown(text) {
		return label + text + "\n"
	}
	return label + "\n" + RenderMarkdown(text, width) + "\n"
}

// looksLikeMarkdown returns true if the text contains markdown syntax
// markers worth rendering (code fences, lists, headings, tables, bold, etc.).
// Short plain-text messages skip the glamour renderer entirely — this
// avoids the latency hit on fast chat turns where there's nothing to format.
func looksLikeMarkdown(text string) bool {
	// Multi-paragraph text is likely worth formatting.
	if strings.Count(text, "\n\n") >= 1 {
		return true
	}
	markers := []string{"```", "# ", "## ", "### ", "- ", "* ", "1. ", "| ", "**", "`", "["}
	for _, m := range markers {
		if strings.Contains(text, m) {
			return true
		}
	}
	return false
}

// glamourTheme returns a JSON theme for glamour that blends with the
// current TUI color scheme. We keep a simple embedded theme so we don't
// need external files.
func glamourTheme(mode string) string {
	if mode == "light" {
		return lightGlamourJSON
	}
	return darkGlamourJSON
}

const darkGlamourJSON = `{
  "dark": {
    "document": {
      "style": "",
      "margin": 0
    },
    "block_quote": {
      "prefix": "▍ ",
      "style": "#757575"
    },
    "paragraph": { "style": "", "margin": 0 },
    "heading": {
      "style": "bold",
      "color": "#d787ff",
      "margin": 0
    },
    "h1": {
      "style": "bold",
      "color": "#d787ff",
      "background_color": "",
      "prefix": "",
      "suffix": "",
      "margin": 0
    },
    "h2": {
      "style": "bold",
      "color": "#d787ff",
      "margin": 0
    },
    "h3": {
      "style": "bold",
      "color": "#d787ff",
      "margin": 0
    },
    "h4": {
      "style": "bold",
      "color": "#d787ff",
      "margin": 0
    },
    "strong": { "style": "bold", "color": "#ffffff" },
    "emph": { "style": "italic", "color": "#ffffff" },
    "code": {
      "style": "",
      "color": "#ff87af",
      "background_color": "#303030",
      "margin": 0
    },
    "code_block": {
      "style": "",
      "color": "#e0e0e0",
      "background_color": "",
      "margin": 0
    },
    "link": { "style": "underline", "color": "#5fafff" },
    "link_text": { "style": "underline", "color": "#5fafff" },
    "list": { "style": "", "margin": 0, "level_indent": 2 },
    "item": { "style": "", "margin": 0 },
    "hr": { "style": "", "color": "#757575" },
    "image": { "style": "", "color": "#757575" },
    "table": { "style": "", "margin": 0 },
    "table_header": { "style": "bold", "color": "#ffffff" },
    "definition_description": { "style": "", "margin": 0 },
    "html_block": { "style": "", "margin": 0 }
  }
}`

const lightGlamourJSON = `{
  "light": {
    "document": {
      "style": "",
      "margin": 0
    },
    "block_quote": {
      "prefix": "▍ ",
      "style": "#949494"
    },
    "paragraph": { "style": "", "margin": 0 },
    "heading": {
      "style": "bold",
      "color": "#af005f",
      "margin": 0
    },
    "h1": {
      "style": "bold",
      "color": "#af005f",
      "background_color": "",
      "prefix": "",
      "suffix": "",
      "margin": 0
    },
    "h2": {
      "style": "bold",
      "color": "#af005f",
      "margin": 0
    },
    "h3": {
      "style": "bold",
      "color": "#af005f",
      "margin": 0
    },
    "h4": {
      "style": "bold",
      "color": "#af005f",
      "margin": 0
    },
    "strong": { "style": "bold", "color": "#080808" },
    "emph": { "style": "italic", "color": "#080808" },
    "code": {
      "style": "",
      "color": "#af005f",
      "background_color": "#e8e8e8",
      "margin": 0
    },
    "code_block": {
      "style": "",
      "color": "#080808",
      "background_color": "",
      "margin": 0
    },
    "link": { "style": "underline", "color": "#005faf" },
    "link_text": { "style": "underline", "color": "#005faf" },
    "list": { "style": "", "margin": 0, "level_indent": 2 },
    "item": { "style": "", "margin": 0 },
    "hr": { "style": "", "color": "#949494" },
    "image": { "style": "", "color": "#949494" },
    "table": { "style": "", "margin": 0 },
    "table_header": { "style": "bold", "color": "#080808" },
    "definition_description": { "style": "", "margin": 0 },
    "html_block": { "style": "", "margin": 0 }
  }
}`
