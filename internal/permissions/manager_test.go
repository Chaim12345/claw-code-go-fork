package permissions

import (
	"testing"
)

func TestNewManager_NilRules(t *testing.T) {
	m := NewManager(ModeDefault, nil)
	if m.Rules == nil {
		t.Error("NewManager should initialize empty Ruleset when nil")
	}
}

func TestManager_Check_BypassMode(t *testing.T) {
	m := NewManager(ModeBypassPermissions, nil)
	d := m.Check("bash", "rm -rf /")
	if d != DecisionAllow {
		t.Errorf("bypass mode: d=%v, want Allow", d)
	}
}

func TestManager_Check_PlanMode(t *testing.T) {
	m := NewManager(ModePlan, nil)
	d := m.Check("bash", "anything")
	if d != DecisionDeny {
		t.Errorf("plan mode: d=%v, want Deny", d)
	}
}

func TestManager_Check_DefaultMode_RuleMatch(t *testing.T) {
	rs := &Ruleset{
		Rules: []Rule{
			{Tool: "bash", Decision: DecisionAllow},
		},
	}
	m := NewManager(ModeDefault, rs)
	d := m.Check("bash", "")
	if d != DecisionAllow {
		t.Errorf("default+rule: d=%v, want Allow", d)
	}
}

func TestManager_Check_DefaultMode_NoRule(t *testing.T) {
	m := NewManager(ModeDefault, nil)
	d := m.Check("unknown_tool", "")
	if d != DecisionAsk {
		t.Errorf("default+no rule: d=%v, want Ask", d)
	}
}

func TestManager_Check_AcceptEdits_AllowFileTools(t *testing.T) {
	m := NewManager(ModeAcceptEdits, nil)
	tools := []string{"read_file", "glob", "grep", "write_file"}
	for _, tool := range tools {
		d := m.Check(tool, "")
		if d != DecisionAllow {
			t.Errorf("accept-edits+%s: d=%v, want Allow", tool, d)
		}
	}
}

func TestManager_Check_AcceptEdits_AskBash(t *testing.T) {
	m := NewManager(ModeAcceptEdits, nil)
	d := m.Check("bash", "echo hi")
	if d != DecisionAsk {
		t.Errorf("accept-edits+bash: d=%v, want Ask", d)
	}
}

func TestManager_Remember_CachesDecision(t *testing.T) {
	m := NewManager(ModeDefault, nil)
	m.Remember("bash", "", DecisionAllow, ScopeAlways)

	d := m.Check("bash", "")
	if d != DecisionAllow {
		t.Errorf("after Remember: d=%v, want Allow", d)
	}
}

func TestManager_Remember_ScopeOnce_NoCache(t *testing.T) {
	m := NewManager(ModeDefault, nil)
	m.Remember("bash", "", DecisionAllow, ScopeOnce)

	d := m.Check("bash", "")
	if d != DecisionAsk {
		t.Errorf("ScopeOnce should not cache: d=%v, want Ask", d)
	}
}
