package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"claw-code-go/internal/auth"
)

func TestNewRegistry_HasBuiltins(t *testing.T) {
	r := NewRegistry()
	cmds := r.List()
	if len(cmds) < 5 {
		t.Errorf("List() = %d commands, want >= 5", len(cmds))
	}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	called := false
	r.Register(Command{
		Name:        "custom",
		Description: "test",
		Handler: func(args string, loop interface{}) error {
			called = true
			return nil
		},
	})

	ok, err := r.Execute("/custom", nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !ok {
		t.Error("Execute returned false for registered command")
	}
	if !called {
		t.Error("handler was not called")
	}
}

func TestRegistry_Execute_NotCommand(t *testing.T) {
	r := NewRegistry()
	ok, err := r.Execute("hello world", nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if ok {
		t.Error("Execute should return false for non-command input")
	}
}

func TestRegistry_Execute_UnknownCommand(t *testing.T) {
	r := NewRegistry()
	ok, err := r.Execute("/nope", nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !ok {
		t.Error("unknown command should still return true (handled)")
	}
}

func TestRegistry_Execute_Exit(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute("/exit", nil)
	if err != ErrExit {
		t.Errorf("Execute /exit: err = %v, want ErrExit", err)
	}
}

func TestRegistry_Execute_Quit(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute("/quit", nil)
	if err != ErrExit {
		t.Errorf("Execute /quit: err = %v, want ErrExit", err)
	}
}

func TestRegistry_Execute_WithArgs(t *testing.T) {
	r := NewRegistry()
	var gotArgs string
	r.Register(Command{
		Name: "test",
		Handler: func(args string, loop interface{}) error { gotArgs = args; return nil },
	})
	r.Execute("/test arg1 arg2", nil)
	if gotArgs != "arg1 arg2" {
		t.Errorf("args = %q, want %q", gotArgs, "arg1 arg2")
	}
}

func TestRegistry_Execute_CaseInsensitive(t *testing.T) {
	r := NewRegistry()
	called := false
	r.Register(Command{
		Name: "mycmd",
		Handler: func(args string, loop interface{}) error { called = true; return nil },
	})
	r.Execute("/MYCMD", nil)
	if !called {
		t.Error("command should be case-insensitive")
	}
}

func TestRegistry_Execute_StripsSlash(t *testing.T) {
	r := NewRegistry()
	called := false
	r.Register(Command{
		Name:        "/leading-slash",
		Description: "test",
		Handler: func(args string, loop interface{}) error { called = true; return nil },
	})
	r.Execute("/leading-slash", nil)
	if !called {
		t.Error("Register should strip leading slash from name")
	}
}

func TestRegisterAuthCommands(t *testing.T) {
	r := NewRegistry()
	RegisterAuthCommands(r)
	cmds := r.List()
	found := false
	for _, c := range cmds {
		if c.Name == "auth" {
			found = true
			break
		}
	}
	if !found {
		t.Error("/auth command not registered")
	}
}

func TestHandleAuthCommand_DefaultStatus(t *testing.T) {
	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth", nil)
	if err != nil {
		t.Errorf("/auth (default): %v", err)
	}
}

func TestHandleAuthCommand_Logout(t *testing.T) {
	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth logout", nil)
	if err != nil {
		t.Errorf("/auth logout: %v", err)
	}
}

func TestHandleAuthCommand_UnknownSub(t *testing.T) {
	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth foobar", nil)
	if err != nil {
		t.Errorf("/auth foobar: %v", err)
	}
}

func TestRegisterMCPCommand(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	cmds := r.List()
	found := false
	for _, c := range cmds {
		if c.Name == "mcp" {
			found = true
			break
		}
	}
	if !found {
		t.Error("/mcp command not registered")
	}
}

func TestMCPCommand_NoLoop(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	_, err := r.Execute("/mcp list", nil)
	if err != nil {
		t.Errorf("/mcp with nil loop: %v", err)
	}
}

func TestMCPCommand_NoArgs(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	loop := &fakeMCPLoop{}
	_, err := r.Execute("/mcp", loop)
	if err != nil {
		t.Errorf("/mcp with no args: %v", err)
	}
}

func TestMCPCommand_List(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	loop := &fakeMCPLoop{}
	_, err := r.Execute("/mcp list", loop)
	if err != nil {
		t.Errorf("/mcp list: %v", err)
	}
}

func TestMCPCommand_Connect(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	loop := &fakeMCPLoop{}
	_, err := r.Execute("/mcp connect myserver", loop)
	if err != nil {
		t.Errorf("/mcp connect: %v", err)
	}
	if !loop.connected {
		t.Error("MCPConnect not called")
	}
}

func TestMCPCommand_ConnectNoName(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	loop := &fakeMCPLoop{}
	_, err := r.Execute("/mcp connect", loop)
	if err != nil {
		t.Errorf("/mcp connect (no name): %v", err)
	}
}

func TestMCPCommand_Disconnect(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	loop := &fakeMCPLoop{}
	_, err := r.Execute("/mcp disconnect myserver", loop)
	if err != nil {
		t.Errorf("/mcp disconnect: %v", err)
	}
	if !loop.disconnected {
		t.Error("MCPDisconnect not called")
	}
}

func TestMCPCommand_DisconnectNoName(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	loop := &fakeMCPLoop{}
	_, err := r.Execute("/mcp disconnect", loop)
	if err != nil {
		t.Errorf("/mcp disconnect (no name): %v", err)
	}
}

func TestMCPCommand_UnknownSub(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	loop := &fakeMCPLoop{}
	_, err := r.Execute("/mcp foobar", loop)
	if err != nil {
		t.Errorf("/mcp foobar: %v", err)
	}
}

type fakeMCPLoop struct {
	connected   bool
	disconnected bool
}

func (f *fakeMCPLoop) MCPConnect(_ context.Context, _ string) error {
	f.connected = true
	return nil
}

func (f *fakeMCPLoop) MCPDisconnect(_ string) error {
	f.disconnected = true
	return nil
}

func (f *fakeMCPLoop) MCPList() string {
	return "No MCP servers connected.\n"
}

type fakeMCPLoopErrors struct {
	connectErr    error
	disconnectErr error
}

func (f *fakeMCPLoopErrors) MCPConnect(_ context.Context, _ string) error {
	return f.connectErr
}

func (f *fakeMCPLoopErrors) MCPDisconnect(_ string) error {
	return f.disconnectErr
}

func (f *fakeMCPLoopErrors) MCPList() string {
	return "No MCP servers connected.\n"
}

func TestMCPCommand_ConnectError(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	loop := &fakeMCPLoopErrors{connectErr: errors.New("connection refused")}
	_, err := r.Execute("/mcp connect badserver", loop)
	if err != nil {
		t.Errorf("/mcp connect error: expected nil, got %v", err)
	}
}

func TestMCPCommand_DisconnectError(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	loop := &fakeMCPLoopErrors{disconnectErr: errors.New("not found")}
	_, err := r.Execute("/mcp disconnect badserver", loop)
	if err != nil {
		t.Errorf("/mcp disconnect error: expected nil, got %v", err)
	}
}

func TestMCPCommand_WrongLoopType(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	_, err := r.Execute("/mcp list", "not-an-mcp-loop")
	if err != nil {
		t.Errorf("/mcp with wrong loop type: expected nil, got %v", err)
	}
}

func TestHandleAuthCommand_LoginSuccess(t *testing.T) {
	orig := authStartOAuthFlow
	defer func() { authStartOAuthFlow = orig }()
	authStartOAuthFlow = func() (*auth.TokenData, error) {
		return &auth.TokenData{
			AccessToken:  "test-access",
			RefreshToken: "test-refresh",
			ExpiresAt:    time.Now().Add(time.Hour),
			TokenType:    "Bearer",
			Scope:        "openid",
		}, nil
	}

	origSave := authSaveTokens
	defer func() { authSaveTokens = origSave }()
	authSaveTokens = func(_ *auth.TokenData) error { return nil }

	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth login", nil)
	if err != nil {
		t.Errorf("/auth login success: %v", err)
	}
}

func TestHandleAuthCommand_LoginOAuthError(t *testing.T) {
	orig := authStartOAuthFlow
	defer func() { authStartOAuthFlow = orig }()
	authStartOAuthFlow = func() (*auth.TokenData, error) {
		return nil, errors.New("oauth server unreachable")
	}

	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth login", nil)
	if err == nil {
		t.Error("/auth login oauth error: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "login:") {
		t.Errorf("error should wrap 'login:', got %v", err)
	}
}

func TestHandleAuthCommand_LoginSaveError(t *testing.T) {
	orig := authStartOAuthFlow
	defer func() { authStartOAuthFlow = orig }()
	authStartOAuthFlow = func() (*auth.TokenData, error) {
		return &auth.TokenData{
			AccessToken: "test-access",
			ExpiresAt:   time.Now().Add(time.Hour),
			TokenType:   "Bearer",
		}, nil
	}

	origSave := authSaveTokens
	defer func() { authSaveTokens = origSave }()
	authSaveTokens = func(_ *auth.TokenData) error { return errors.New("disk full") }

	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth login", nil)
	if err == nil {
		t.Error("/auth login save error: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "save tokens:") {
		t.Errorf("error should wrap 'save tokens:', got %v", err)
	}
}

func TestHandleAuthCommand_LogoutError(t *testing.T) {
	orig := authClearTokens
	defer func() { authClearTokens = orig }()
	authClearTokens = func() error { return errors.New("permission denied") }

	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth logout", nil)
	if err == nil {
		t.Error("/auth logout error: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "logout:") {
		t.Errorf("error should wrap 'logout:', got %v", err)
	}
}

func TestHandleAuthCommand_StatusOAuthWithExpiry(t *testing.T) {
	orig := authGetStatus
	defer func() { authGetStatus = orig }()
	authGetStatus = func() auth.Status {
		return auth.Status{
			Authenticated: true,
			Method:        "oauth",
			ExpiresAt:     time.Date(2025, 12, 25, 10, 30, 0, 0, time.UTC),
			HasRefresh:    true,
		}
	}

	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth status", nil)
	if err != nil {
		t.Errorf("/auth status oauth: %v", err)
	}
}

func TestHandleAuthCommand_StatusOAuthZeroExpiry(t *testing.T) {
	orig := authGetStatus
	defer func() { authGetStatus = orig }()
	authGetStatus = func() auth.Status {
		return auth.Status{
			Authenticated: true,
			Method:        "oauth",
			ExpiresAt:     time.Time{},
			HasRefresh:    false,
		}
	}

	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth status", nil)
	if err != nil {
		t.Errorf("/auth status oauth zero expiry: %v", err)
	}
}

func TestHandleAuthCommand_StatusAPIKey(t *testing.T) {
	orig := authGetStatus
	defer func() { authGetStatus = orig }()
	authGetStatus = func() auth.Status {
		return auth.Status{
			Authenticated: true,
			Method:        "api_key",
			ExpiresAt:     time.Time{},
			HasRefresh:    false,
		}
	}

	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth status", nil)
	if err != nil {
		t.Errorf("/auth status api_key: %v", err)
	}
}

func TestRegistry_Execute_Help(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute("/help", nil)
	if err != nil {
		t.Errorf("/help: %v", err)
	}
}

type fakeSessionHolder struct {
	cleared bool
}

func (f *fakeSessionHolder) ClearSession() {
	f.cleared = true
}

func TestRegistry_Execute_Clear_WithHolder(t *testing.T) {
	r := NewRegistry()
	holder := &fakeSessionHolder{}
	_, err := r.Execute("/clear", holder)
	if err != nil {
		t.Errorf("/clear with holder: %v", err)
	}
	if !holder.cleared {
		t.Error("ClearSession() was not called")
	}
}

func TestRegistry_Execute_Clear_WithoutHolder(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute("/clear", nil)
	if err != nil {
		t.Errorf("/clear without holder: %v", err)
	}
}

func TestRegistry_Execute_Clear_WrongType(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute("/clear", "not-a-holder")
	if err != nil {
		t.Errorf("/clear wrong type: %v", err)
	}
}

type fakeSessionLister struct {
	sessions []string
	err      error
}

func (f *fakeSessionLister) ListSessions() ([]string, error) {
	return f.sessions, f.err
}

func TestRegistry_Execute_SessionList_WithSessions(t *testing.T) {
	r := NewRegistry()
	sl := &fakeSessionLister{sessions: []string{"sess-001", "sess-002"}}
	_, err := r.Execute("/session-list", sl)
	if err != nil {
		t.Errorf("/session-list with sessions: %v", err)
	}
}

func TestRegistry_Execute_SessionList_Empty(t *testing.T) {
	r := NewRegistry()
	sl := &fakeSessionLister{sessions: []string{}}
	_, err := r.Execute("/session-list", sl)
	if err != nil {
		t.Errorf("/session-list empty: %v", err)
	}
}

func TestRegistry_Execute_SessionList_Error(t *testing.T) {
	r := NewRegistry()
	sl := &fakeSessionLister{err: errors.New("read error")}
	_, err := r.Execute("/session-list", sl)
	if err == nil {
		t.Error("/session-list error: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "list sessions:") {
		t.Errorf("error should wrap 'list sessions:', got %v", err)
	}
}

func TestRegistry_Execute_SessionList_NoInterface(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute("/session-list", nil)
	if err != nil {
		t.Errorf("/session-list nil loop: %v", err)
	}
}

func TestRegistry_Execute_SessionList_WrongType(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute("/session-list", 42)
	if err != nil {
		t.Errorf("/session-list wrong type: %v", err)
	}
}

type fakeCompactor struct {
	summary string
	err     error
}

func (f *fakeCompactor) CompactNow(_ context.Context) (string, error) {
	return f.summary, f.err
}

func TestRegistry_Execute_Compact_Success(t *testing.T) {
	r := NewRegistry()
	c := &fakeCompactor{summary: "Session was about fixing auth bugs."}
	_, err := r.Execute("/compact", c)
	if err != nil {
		t.Errorf("/compact success: %v", err)
	}
}

func TestRegistry_Execute_Compact_LongPreview(t *testing.T) {
	r := NewRegistry()
	longSummary := strings.Repeat("x", 300)
	c := &fakeCompactor{summary: longSummary}
	_, err := r.Execute("/compact", c)
	if err != nil {
		t.Errorf("/compact long preview: %v", err)
	}
}

func TestRegistry_Execute_Compact_EmptySummary(t *testing.T) {
	r := NewRegistry()
	c := &fakeCompactor{summary: ""}
	_, err := r.Execute("/compact", c)
	if err != nil {
		t.Errorf("/compact empty summary: %v", err)
	}
}

func TestRegistry_Execute_Compact_Error(t *testing.T) {
	r := NewRegistry()
	c := &fakeCompactor{err: errors.New("provider error")}
	_, err := r.Execute("/compact", c)
	if err == nil {
		t.Error("/compact error: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "compact:") {
		t.Errorf("error should wrap 'compact:', got %v", err)
	}
}

func TestRegistry_Execute_Compact_NoInterface(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute("/compact", nil)
	if err != nil {
		t.Errorf("/compact nil loop: %v", err)
	}
}

func TestRegistry_Execute_Compact_WrongType(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute("/compact", []string{"not", "a", "compactor"})
	if err != nil {
		t.Errorf("/compact wrong type: %v", err)
	}
}

func TestRegistry_Execute_HelpOutput(t *testing.T) {
	r := NewRegistry()
	RegisterAuthCommands(r)
	RegisterMCPCommand(r)
	cmds := r.List()
	_, err := r.Execute("/help", nil)
	if err != nil {
		t.Fatalf("/help: %v", err)
	}
	if len(cmds) < 7 {
		t.Errorf("expected >= 7 commands with auth+mcp registered, got %d", len(cmds))
	}
}

func TestHandleAuthCommand_LoginBranch(t *testing.T) {
	orig := authStartOAuthFlow
	defer func() { authStartOAuthFlow = orig }()
	authStartOAuthFlow = func() (*auth.TokenData, error) {
		return &auth.TokenData{AccessToken: "at", ExpiresAt: time.Now().Add(time.Hour), TokenType: "Bearer"}, nil
	}

	origSave := authSaveTokens
	defer func() { authSaveTokens = origSave }()
	authSaveTokens = func(_ *auth.TokenData) error { return nil }

	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth login", nil)
	if err != nil {
		t.Errorf("/auth login branch: %v", err)
	}
}

func TestMCPCommand_ConnectSuccess(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	loop := &fakeMCPLoop{}
	_, err := r.Execute("/mcp connect goodserver", loop)
	if err != nil {
		t.Errorf("/mcp connect success: %v", err)
	}
}

func TestMCPCommand_DisconnectSuccess(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	loop := &fakeMCPLoop{}
	_, err := r.Execute("/mcp disconnect goodserver", loop)
	if err != nil {
		t.Errorf("/mcp disconnect success: %v", err)
	}
}

func TestRegistry_ListContainsAllBuiltins(t *testing.T) {
	r := NewRegistry()
	cmds := r.List()
	names := make(map[string]bool, len(cmds))
	for _, c := range cmds {
		names[c.Name] = true
	}
	for _, name := range []string{"help", "exit", "quit", "clear", "session-list", "compact"} {
		if !names[name] {
			t.Errorf("builtin /%s not found in List()", name)
		}
	}
}

func TestMCPCommand_NilLoopList(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	_, err := r.Execute("/mcp list", nil)
	if err != nil {
		t.Errorf("/mcp list with nil loop: %v", err)
	}
}

func TestMCPCommand_NilLoopConnect(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	_, err := r.Execute("/mcp connect foo", nil)
	if err != nil {
		t.Errorf("/mcp connect with nil loop: %v", err)
	}
}

func TestMCPCommand_NilLoopDisconnect(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	_, err := r.Execute("/mcp disconnect foo", nil)
	if err != nil {
		t.Errorf("/mcp disconnect with nil loop: %v", err)
	}
}

func TestMCPCommand_NilLoopNoArgs(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	_, err := r.Execute("/mcp", nil)
	if err != nil {
		t.Errorf("/mcp no args with nil loop: %v", err)
	}
}

func TestHandleAuthCommand_StatusUnauthenticated(t *testing.T) {
	orig := authGetStatus
	defer func() { authGetStatus = orig }()
	authGetStatus = func() auth.Status {
		return auth.Status{Authenticated: false, Method: "", ExpiresAt: time.Time{}, HasRefresh: false}
	}

	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth status", nil)
	if err != nil {
		t.Errorf("/auth status unauthenticated: %v", err)
	}
}

func TestHandleAuthCommand_LoginExplicitArg(t *testing.T) {
	orig := authStartOAuthFlow
	defer func() { authStartOAuthFlow = orig }()
	authStartOAuthFlow = func() (*auth.TokenData, error) {
		return &auth.TokenData{AccessToken: "at", ExpiresAt: time.Now().Add(time.Hour), TokenType: "Bearer"}, nil
	}

	origSave := authSaveTokens
	defer func() { authSaveTokens = origSave }()
	authSaveTokens = func(_ *auth.TokenData) error { return nil }

	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth login extra-args", nil)
	if err != nil {
		t.Errorf("/auth login with extra args: %v", err)
	}
}

func TestHandleAuthCommand_DefaultNoArgs(t *testing.T) {
	orig := authGetStatus
	defer func() { authGetStatus = orig }()
	authGetStatus = func() auth.Status {
		return auth.Status{Authenticated: false, Method: ""}
	}

	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth", nil)
	if err != nil {
		t.Errorf("/auth default (no args): %v", err)
	}
}

func TestRegistry_Execute_HelpIteration(t *testing.T) {
	r := NewRegistry()
	r.Register(Command{Name: "z-cmd", Description: "Z test", Handler: func(string, interface{}) error { return nil }})
	r.Register(Command{Name: "a-cmd", Description: "A test", Handler: func(string, interface{}) error { return nil }})
	_, err := r.Execute("/help", nil)
	if err != nil {
		t.Errorf("/help with extra commands: %v", err)
	}
}

func TestRegistry_Execute_ClearSessionPrint(t *testing.T) {
	r := NewRegistry()
	holder := &fakeSessionHolder{}
	ok, err := r.Execute("/clear", holder)
	if !ok {
		t.Error("/clear should return true")
	}
	if err != nil {
		t.Errorf("/clear: %v", err)
	}
	if !holder.cleared {
		t.Error("ClearSession not called")
	}
}

func TestRegistry_Execute_CompactExact200(t *testing.T) {
	r := NewRegistry()
	exactly200 := strings.Repeat("x", 200)
	c := &fakeCompactor{summary: exactly200}
	_, err := r.Execute("/compact", c)
	if err != nil {
		t.Errorf("/compact exact 200: %v", err)
	}
}

func TestRegistry_Execute_CompactShortPreview(t *testing.T) {
	r := NewRegistry()
	c := &fakeCompactor{summary: "Short summary"}
	_, err := r.Execute("/compact", c)
	if err != nil {
		t.Errorf("/compact short preview: %v", err)
	}
}

func TestRegistry_Execute_SessionListMultiple(t *testing.T) {
	r := NewRegistry()
	sl := &fakeSessionLister{sessions: []string{"a", "b", "c", "d", "e"}}
	_, err := r.Execute("/session-list", sl)
	if err != nil {
		t.Errorf("/session-list multiple: %v", err)
	}
}

func TestRegistry_Execute_ExitErrorMessage(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute("/exit", nil)
	if err == nil {
		t.Error("/exit should return ErrExit")
	}
	if err.Error() != "exit" {
		t.Errorf("/exit error message = %q, want %q", err.Error(), "exit")
	}
}

func TestHandleAuthCommand_LoginSaveSuccessMessage(t *testing.T) {
	orig := authStartOAuthFlow
	defer func() { authStartOAuthFlow = orig }()
	authStartOAuthFlow = func() (*auth.TokenData, error) {
		return &auth.TokenData{
			AccessToken:  "at",
			RefreshToken: "rt",
			ExpiresAt:    time.Now().Add(2 * time.Hour),
			TokenType:    "Bearer",
			Scope:        "openid profile",
		}, nil
	}

	origSave := authSaveTokens
	defer func() { authSaveTokens = origSave }()
	var savedTD *auth.TokenData
	authSaveTokens = func(td *auth.TokenData) error { savedTD = td; return nil }

	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth login", nil)
	if err != nil {
		t.Fatalf("/auth login save success: %v", err)
	}
	if savedTD == nil {
		t.Error("SaveTokens was not called")
	}
	if savedTD.AccessToken != "at" {
		t.Errorf("saved AccessToken = %q, want %q", savedTD.AccessToken, "at")
	}
}

func TestMCPCommand_WrongLoopTypeConnect(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	_, err := r.Execute("/mcp connect foo", 123)
	if err != nil {
		t.Errorf("/mcp connect wrong loop type: expected nil, got %v", err)
	}
}

func TestMCPCommand_WrongLoopTypeDisconnect(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	_, err := r.Execute("/mcp disconnect foo", 123)
	if err != nil {
		t.Errorf("/mcp disconnect wrong loop type: expected nil, got %v", err)
	}
}

func TestMCPCommand_WrongLoopTypeNoArgs(t *testing.T) {
	r := NewRegistry()
	RegisterMCPCommand(r)
	_, err := r.Execute("/mcp", 123)
	if err != nil {
		t.Errorf("/mcp no args wrong loop type: expected nil, got %v", err)
	}
}

func TestRegistry_Execute_UnknownCommandMessage(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute("/nonexistent", nil)
	if err != nil {
		t.Errorf("/nonexistent: %v", err)
	}
}

func TestHandleAuthCommand_LogoutSuccess(t *testing.T) {
	orig := authClearTokens
	defer func() { authClearTokens = orig }()
	cleared := false
	authClearTokens = func() error { cleared = true; return nil }

	r := NewRegistry()
	RegisterAuthCommands(r)
	_, err := r.Execute("/auth logout", nil)
	if err != nil {
		t.Errorf("/auth logout success: %v", err)
	}
	if !cleared {
		t.Error("ClearTokens was not called")
	}
}

func TestRegistry_Execute_CompactErrorMessage(t *testing.T) {
	r := NewRegistry()
	c := &fakeCompactor{err: fmt.Errorf("compact: context cancelled")}
	_, err := r.Execute("/compact", c)
	if err == nil {
		t.Error("/compact with error: expected error")
	}
}

func TestRegistry_Execute_SessionListErrorMessage(t *testing.T) {
	r := NewRegistry()
	sl := &fakeSessionLister{err: fmt.Errorf("list sessions: permission denied")}
	_, err := r.Execute("/session-list", sl)
	if err == nil {
		t.Error("/session-list with error: expected error")
	}
}
