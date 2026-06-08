package permissions

import (
	"testing"
)

func TestParsePermissionMode(t *testing.T) {
	tests := []struct {
		input string
		want  PermissionMode
		err   bool
	}{
		{"default", ModeDefault, false},
		{"", ModeDefault, false},
		{"accept-edits", ModeAcceptEdits, false},
		{"bypass", ModeBypassPermissions, false},
		{"auto", ModeBypassPermissions, false},
		{"plan", ModePlan, false},
		{"bogus", ModeDefault, true},
	}
	for _, tt := range tests {
		got, err := ParsePermissionMode(tt.input)
		if (err != nil) != tt.err {
			t.Errorf("ParsePermissionMode(%q) error = %v, wantErr %v", tt.input, err, tt.err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParsePermissionMode(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestPermissionMode_String(t *testing.T) {
	tests := []struct {
		mode PermissionMode
		want string
	}{
		{ModeDefault, "default"},
		{ModeAcceptEdits, "accept-edits"},
		{ModeBypassPermissions, "bypass"},
		{ModePlan, "plan"},
		{PermissionMode(999), "unknown"},
	}
	for _, tt := range tests {
		got := tt.mode.String()
		if got != tt.want {
			t.Errorf("PermissionMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}
