package permissions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRuleset_Match_Nil(t *testing.T) {
	var rs *Ruleset
	d, ok := rs.Match("bash", "rm -rf /")
	if ok {
		t.Error("nil Ruleset should not match")
	}
	if d != DecisionAsk {
		t.Errorf("nil Ruleset decision = %v, want DecisionAsk", d)
	}
}

func TestRuleset_Match_ToolWildcard(t *testing.T) {
	rs := &Ruleset{
		Rules: []Rule{
			{Tool: "*", Decision: DecisionAllow},
		},
	}
	d, ok := rs.Match("bash", "anything")
	if !ok || d != DecisionAllow {
		t.Errorf("wildcard match: d=%v ok=%v, want Allow/true", d, ok)
	}
}

func TestRuleset_Match_ToolSpecific(t *testing.T) {
	rs := &Ruleset{
		Rules: []Rule{
			{Tool: "bash", Decision: DecisionDeny},
			{Tool: "read_file", Decision: DecisionAllow},
		},
	}
	d, ok := rs.Match("bash", "")
	if !ok || d != DecisionDeny {
		t.Errorf("bash: d=%v ok=%v, want Deny/true", d, ok)
	}
	d, ok = rs.Match("read_file", "")
	if !ok || d != DecisionAllow {
		t.Errorf("read_file: d=%v ok=%v, want Allow/true", d, ok)
	}
}

func TestRuleset_Match_Pattern(t *testing.T) {
	rs := &Ruleset{
		Rules: []Rule{
			{Tool: "bash", Pattern: "git", Decision: DecisionAllow},
			{Tool: "bash", Decision: DecisionDeny},
		},
	}
	d, ok := rs.Match("bash", "git status")
	if !ok || d != DecisionAllow {
		t.Errorf("bash+git: d=%v ok=%v, want Allow/true", d, ok)
	}
	d, ok = rs.Match("bash", "rm -rf /")
	if !ok || d != DecisionDeny {
		t.Errorf("bash+rm: d=%v ok=%v, want Deny/true", d, ok)
	}
}

func TestRuleset_Match_NoMatch(t *testing.T) {
	rs := &Ruleset{
		Rules: []Rule{
			{Tool: "bash", Decision: DecisionDeny},
		},
	}
	d, ok := rs.Match("unknown_tool", "")
	if ok {
		t.Errorf("no match: ok=%v, want false", ok)
	}
	if d != DecisionAsk {
		t.Errorf("no match: d=%v, want DecisionAsk", d)
	}
}

func TestRuleset_Match_FirstWins(t *testing.T) {
	rs := &Ruleset{
		Rules: []Rule{
			{Tool: "bash", Decision: DecisionAllow},
			{Tool: "bash", Decision: DecisionDeny},
		},
	}
	d, ok := rs.Match("bash", "")
	if !ok || d != DecisionAllow {
		t.Errorf("first wins: d=%v ok=%v, want Allow/true", d, ok)
	}
}

func TestLoadRuleset_NotExist(t *testing.T) {
	rs, err := LoadRuleset("/nonexistent/path/settings.json")
	if err != nil {
		t.Fatalf("LoadRuleset: %v", err)
	}
	if len(rs.Rules) != 0 {
		t.Errorf("Rules = %d, want 0", len(rs.Rules))
	}
}

func TestLoadRuleset_WithRules(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	content := `{
		"rules": [
			{"tool": "bash", "pattern": "git", "decision": "allow"},
			{"tool": "write_file", "decision": "deny"}
		],
		"allowedTools": ["read_file"],
		"blockedTools": ["web_search"]
	}`
	os.WriteFile(path, []byte(content), 0o600)

	rs, err := LoadRuleset(path)
	if err != nil {
		t.Fatalf("LoadRuleset: %v", err)
	}

	if len(rs.Rules) != 4 {
		t.Errorf("Rules = %d, want 4", len(rs.Rules))
	}

	d, ok := rs.Match("bash", "git commit")
	if !ok || d != DecisionAllow {
		t.Errorf("bash+git: d=%v ok=%v, want Allow/true", d, ok)
	}

	d, ok = rs.Match("write_file", "anything")
	if !ok || d != DecisionDeny {
		t.Errorf("write_file: d=%v ok=%v, want Deny/true", d, ok)
	}

	d, ok = rs.Match("read_file", "")
	if !ok || d != DecisionAllow {
		t.Errorf("read_file: d=%v ok=%v, want Allow/true", d, ok)
	}

	d, ok = rs.Match("web_search", "")
	if !ok || d != DecisionDeny {
		t.Errorf("web_search: d=%v ok=%v, want Deny/true", d, ok)
	}
}

func TestLoadRuleset_ReadError(t *testing.T) {
	rs, err := LoadRuleset("/dev/null/is-not-a-dir/settings.json")
	if err == nil {
		t.Error("LoadRuleset with unreadable path should return error")
	}
	if rs != nil {
		t.Error("rs should be nil on error")
	}
}

func TestLoadRuleset_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	os.WriteFile(path, []byte("not json"), 0o600)

	_, err := LoadRuleset(path)
	if err == nil {
		t.Error("LoadRuleset with invalid JSON should return error")
	}
}

func TestRulesetFromLists(t *testing.T) {
	rs := RulesetFromLists([]string{"read_file", "glob"}, []string{"bash"})
	if len(rs.Rules) != 3 {
		t.Errorf("Rules = %d, want 3", len(rs.Rules))
	}

	// Allowed first
	d, ok := rs.Match("read_file", "")
	if !ok || d != DecisionAllow {
		t.Errorf("read_file: d=%v ok=%v, want Allow/true", d, ok)
	}

	d, ok = rs.Match("bash", "")
	if !ok || d != DecisionDeny {
		t.Errorf("bash: d=%v ok=%v, want Deny/true", d, ok)
	}
}

func TestRulesetFromLists_AllowWinsOnConflict(t *testing.T) {
	rs := RulesetFromLists([]string{"bash"}, []string{"bash"})
	d, ok := rs.Match("bash", "")
	if !ok {
		t.Fatal("expected match")
	}
	if d != DecisionAllow {
		t.Errorf("allow should win on conflict, got %v", d)
	}
}

func TestLoadRuleset_DefaultDecision(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	content := `{"rules": [{"tool": "bash", "decision": ""}]}`
	os.WriteFile(path, []byte(content), 0o600)

	rs, err := LoadRuleset(path)
	if err != nil {
		t.Fatalf("LoadRuleset: %v", err)
	}
	if len(rs.Rules) != 1 {
		t.Fatalf("Rules = %d, want 1", len(rs.Rules))
	}
	if rs.Rules[0].Decision != DecisionAsk {
		t.Errorf("empty decision should default to DecisionAsk, got %v", rs.Rules[0].Decision)
	}
}

func TestLoadRuleset_AllowDecision(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	content := `{"rules": [{"tool": "read_file", "decision": "allow"}]}`
	os.WriteFile(path, []byte(content), 0o600)

	rs, err := LoadRuleset(path)
	if err != nil {
		t.Fatalf("LoadRuleset: %v", err)
	}
	if rs.Rules[0].Decision != DecisionAllow {
		t.Errorf("allow decision: got %v, want DecisionAllow", rs.Rules[0].Decision)
	}
}

func TestLoadRuleset_DenyDecision(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	content := `{"rules": [{"tool": "bash", "decision": "deny"}]}`
	os.WriteFile(path, []byte(content), 0o600)

	rs, err := LoadRuleset(path)
	if err != nil {
		t.Fatalf("LoadRuleset: %v", err)
	}
	if rs.Rules[0].Decision != DecisionDeny {
		t.Errorf("deny decision: got %v, want DecisionDeny", rs.Rules[0].Decision)
	}
}
