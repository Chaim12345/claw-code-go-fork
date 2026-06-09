package search

import (
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// Only actual Go language keywords — NOT package names or common types.
var goKeywords = map[string]bool{
	"break": true, "case": true, "chan": true, "const": true, "continue": true,
	"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
	"func": true, "go": true, "goto": true, "if": true, "import": true,
	"interface": true, "map": true, "package": true, "range": true, "return": true,
	"select": true, "struct": true, "switch": true, "type": true, "var": true,
	"true": true, "false": true, "nil": true,
}

var stopwords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
	"with": true, "by": true, "from": true, "is": true, "it": true, "as": true,
	"was": true, "are": true, "be": true, "this": true, "that": true, "not": true,
	"has": true, "have": true, "can": true, "will": true, "do": true, "if": true,
	"no": true, "all": true, "its": true, "my": true, "we": true, "our": true,
	"you": true, "your": true, "they": true, "them": true, "their": true,
	"so": true, "up": true, "out": true, "about": true, "into": true,
	"than": true, "then": true, "also": true, "just": true, "more": true,
	"some": true, "any": true, "each": true, "which": true, "when": true,
	"what": true, "how": true, "there": true, "here": true, "been": true,
	"were": true, "being": true, "had": true, "did": true, "get": true,
	"got": true, "use": true, "using": true, "used": true,
}

type SearchResult struct {
	FilePath string  `json:"file_path"`
	Score    float64 `json:"score"`
	Snippet  string  `json:"snippet"`
}

type Index struct {
	docs     []string
	docLen   map[string]int
	postings map[string]map[string][]int
	avgDL    float64
	n        int
	content  map[string]string
}

func NewIndex() *Index {
	return &Index{
		docLen:   make(map[string]int),
		postings: make(map[string]map[string][]int),
		content:  make(map[string]string),
	}
}

func (idx *Index) Build(dir string) error {
	idx.docs = nil
	idx.docLen = make(map[string]int)
	idx.postings = make(map[string]map[string][]int)
	idx.content = make(map[string]string)

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" || name == "scratch_repos" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		text := string(data)
		idx.content[rel] = text
		tokens := tokenize(text)
		idx.docLen[rel] = len(tokens)
		for pos, tok := range tokens {
			if idx.postings[tok] == nil {
				idx.postings[tok] = make(map[string][]int)
			}
			idx.postings[tok][rel] = append(idx.postings[tok][rel], pos)
		}
		idx.docs = append(idx.docs, rel)
		return nil
	})
}

func (idx *Index) Search(query string, limit int) []SearchResult {
	if query == "" || len(idx.docs) == 0 {
		return nil
	}
	if limit <= 0 {
		limit = 10
	}

	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		return nil
	}

	totalDL := 0
	for _, dl := range idx.docLen {
		totalDL += dl
	}
	idx.avgDL = float64(totalDL) / float64(len(idx.docs))
	idx.n = len(idx.docs)

	scores := make(map[string]float64)
	for _, term := range queryTokens {
		postings, ok := idx.postings[term]
		if !ok {
			continue
		}
		df := float64(len(postings))
		n := float64(idx.n)
		idf := math.Log(1 + (n-df+0.5)/(df+0.5))
		for doc, positions := range postings {
			tf := float64(len(positions))
			dl := float64(idx.docLen[doc])
			k1 := 1.2
			b := 0.75
			tfNorm := (tf * (k1 + 1)) / (tf + k1*(1-b+b*dl/idx.avgDL))
			scores[doc] += idf * tfNorm
		}
	}

	var results []SearchResult
	for doc, score := range scores {
		if score > 0 {
			results = append(results, SearchResult{
				FilePath: doc,
				Score:    score,
				Snippet:  snippet(idx.content[doc], queryTokens),
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

func tokenize(text string) []string {
	var tokens []string
	var buf []rune
	flush := func() {
		if len(buf) == 0 {
			return
		}
		lower := strings.ToLower(string(buf))
		if !goKeywords[lower] && !stopwords[lower] && len(lower) > 1 {
			tokens = append(tokens, lower)
		}
		// Split camelCase / PascalCase into sub-tokens for partial matching
		if len(buf) > 3 {
			var part []rune
			for _, r := range buf {
				if unicode.IsUpper(r) && len(part) > 0 {
					s := strings.ToLower(string(part))
					if !goKeywords[s] && !stopwords[s] && len(s) > 1 {
						tokens = append(tokens, s)
					}
					part = part[:0]
				}
				part = append(part, r)
			}
			if len(part) > 0 {
				s := strings.ToLower(string(part))
				if !goKeywords[s] && !stopwords[s] && len(s) > 1 && s != lower {
					tokens = append(tokens, s)
				}
			}
		}
		buf = buf[:0]
	}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			buf = append(buf, r)
		} else {
			flush()
		}
	}
	flush()
	return tokens
}

func snippet(content string, queryTokens []string) string {
	lines := strings.Split(content, "\n")
	bestLine := 0
	bestScore := 0
	qSet := make(map[string]bool)
	for _, t := range queryTokens {
		qSet[t] = true
	}
	for i, line := range lines {
		score := 0
		for _, tok := range tokenize(line) {
			if qSet[tok] {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			bestLine = i
		}
	}
	start := bestLine - 1
	if start < 0 {
		start = 0
	}
	end := bestLine + 2
	if end > len(lines) {
		end = len(lines)
	}
	if end > start+5 {
		end = start + 5
	}
	return strings.Join(lines[start:end], "\n")
}
