package main

import (
	"claw-code-go/internal/api"
	"claw-code-go/internal/auth"
	"claw-code-go/internal/commands"
	"claw-code-go/internal/compat"
	"claw-code-go/internal/permissions"
	"claw-code-go/internal/runtime"
	"claw-code-go/internal/tools"
	"claw-code-go/internal/tui"
	"claw-code-go/internal/web"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	deepseekprovider "claw-code-go/internal/api/providers/deepseek"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Route diagnostic subcommands before flag parsing.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "dump-manifests":
			compat.RunDumpManifests(os.Args[2:])
			return
		case "bootstrap-plan":
			compat.RunBootstrapPlan(os.Args[2:])
			return
		case "print-system-prompt":
			compat.RunPrintSystemPrompt(os.Args[2:])
			return
		case "resume-session":
			compat.RunResumeSession(os.Args[2:])
			return
		case "ralph":
			runRalphSubcommand(os.Args[2:])
			return
		case "web":
			runWebSubcommand(os.Args[2:])
			return
		}
	}

	promptFlag := flag.String("prompt", "", "Run a single prompt and exit")
	modelFlag := flag.String("model", "", "Override the model to use")
	providerFlag := flag.String("provider", "", "AI provider: anthropic, openai, bedrock, vertex, foundry, deepseek (default: detected from env)")
	replFlag := flag.Bool("repl", false, "Run in interactive REPL mode (default when no --prompt)")
	sessionFlag := flag.String("session", "", "Session ID to load")
	sessionDirFlag := flag.String("session-dir", "", "Directory to store sessions")
	permModeFlag := flag.String("permission-mode", "default", "Permission mode: default, accept-edits, bypass, plan")
	deltaFlag := flag.Bool("delta", false, "Enable delta mode (DeepSeek only): send only new messages instead of full history")
	_ = replFlag

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: claw-code-go [subcommand] [options]\n\n")
		fmt.Fprintf(os.Stderr, "Subcommands:\n")
		fmt.Fprintf(os.Stderr, "  dump-manifests [--src <dir>] [--json]   List tools, slash commands, and source manifest\n")
		fmt.Fprintf(os.Stderr, "  bootstrap-plan [--json]                 Print the ordered startup phase plan\n")
		fmt.Fprintf(os.Stderr, "  print-system-prompt [--cwd] [--date]    Render the full system prompt\n")
		fmt.Fprintf(os.Stderr, "  resume-session <file> [commands...]     Replay a saved session file\n")
	fmt.Fprintf(os.Stderr, "  ralph [--spec <path>] [--max-iterations <n>]\n")
	fmt.Fprintf(os.Stderr, "                                          Run a Ralph loop against a spec/roadmap file\n")
	fmt.Fprintf(os.Stderr, "  web [--addr <host:port>] [--cwd <dir>]   Serve the TUI in a browser via WebSocket + wterm\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment variables:\n")
		fmt.Fprintf(os.Stderr, "  ANTHROPIC_API_KEY        Anthropic API key (takes precedence over stored credentials)\n")
		fmt.Fprintf(os.Stderr, "  OPENAI_API_KEY           OpenAI API key (takes precedence over stored credentials)\n")
		fmt.Fprintf(os.Stderr, "  ANTHROPIC_MODEL          Model to use (default: %s)\n", runtime.DefaultModel)
		fmt.Fprintf(os.Stderr, "  ANTHROPIC_BASE_URL       Base URL for the Anthropic API\n")
		fmt.Fprintf(os.Stderr, "  CLAUDE_CODE_USE_BEDROCK  Set to 1 to use AWS Bedrock (env-var fallback)\n")
		fmt.Fprintf(os.Stderr, "  CLAUDE_CODE_USE_VERTEX   Set to 1 to use Google Vertex AI (env-var fallback)\n")
		fmt.Fprintf(os.Stderr, "  CLAUDE_CODE_USE_FOUNDRY  Set to 1 to use Azure AI Foundry (env-var fallback)\n")
	}

	flag.Parse()

	cfg := runtime.LoadConfig()

	if *modelFlag != "" {
		cfg.Model = *modelFlag
	}
	if *providerFlag != "" {
		cfg.ProviderName = *providerFlag
	}
	if *sessionDirFlag != "" {
		cfg.SessionDir = *sessionDirFlag
	}
	if *deltaFlag {
		cfg.DeltaMode = true
	}

	// Resolve credentials and build the provider client. The TUI flow
	// falls back to a no-auth placeholder so the user can /login from
	// inside the chat; non-interactive subcommands handle the error
	// differently (see runRalphSubcommand).
	realClient, buildErr := buildProvider(cfg)
	if buildErr != nil {
		fmt.Fprintf(os.Stderr, "Note: %v\n", buildErr)
		fmt.Fprintln(os.Stderr, "      Use /login in the TUI to authenticate.")
		realClient = runtime.NewNoAuthClient()
	}

	loop := runtime.NewConversationLoop(cfg, realClient)

	// Wire up the permission manager (Phase 11).
	// CLI --permission-mode flag overrides the config-file value when set to a
	// non-default value. cfg.PermissionMode comes from the layered settings files.
	resolvedPermMode := cfg.PermissionMode
	if *permModeFlag != "default" {
		resolvedPermMode = *permModeFlag
	}
	permMode, err := permissions.ParsePermissionMode(resolvedPermMode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v; using default mode\n", err)
		permMode = permissions.ModeDefault
	}
	cfg.PermissionMode = permMode.String() // normalise back into Config

	ruleset, rErr := permissions.LoadRuleset(".claude/settings.json")
	if rErr != nil {
		ruleset = &permissions.Ruleset{}
	}
	// Merge allowedTools/blockedTools from layered config into the ruleset.
	if len(cfg.AllowedTools) > 0 || len(cfg.BlockedTools) > 0 {
		extra := permissions.RulesetFromLists(cfg.AllowedTools, cfg.BlockedTools)
		ruleset.Rules = append(ruleset.Rules, extra.Rules...)
	}
	loop.PermManager = permissions.NewManager(permMode, ruleset)

	// Connect to MCP servers defined in config (non-fatal errors printed inside).
	loop.InitMCPFromConfig(context.Background())

	if *sessionFlag != "" {
		sess, err := runtime.LoadSession(cfg.SessionDir, *sessionFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load session %s: %v\n", *sessionFlag, err)
		} else {
			loop.Session = sess
			// Restore provider-side session state (e.g. DeepSeek
			// chat_session_id) so delta mode can resume.
			if ssp, ok := realClient.(api.SessionStateProvider); ok && sess.ProviderSessionState != "" {
				if err := ssp.RestoreSessionState(sess.ProviderSessionState); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not restore provider session state: %v\n", err)
				}
			}
			fmt.Printf("Loaded session: %s\n", sess.ID)
		}
	}

	// Single prompt (non-interactive) mode — no TUI, plain stdout streaming.
	if *promptFlag != "" {
		if buildErr != nil {
			fmt.Fprintln(os.Stderr, "Error: cannot use --prompt without valid credentials.")
			fmt.Fprintln(os.Stderr, "Set ANTHROPIC_API_KEY or OPENAI_API_KEY, or run the TUI and use /login.")
			os.Exit(1)
		}
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			fmt.Fprintln(os.Stdout, "\nInterrupted. Saving session...")
			saveSessionSilent(cfg.SessionDir, loop)
			os.Exit(0)
		}()

		ctx := context.Background()
		if err := loop.SendMessage(ctx, *promptFlag); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		saveSessionSilent(cfg.SessionDir, loop)
		return
	}

	// Interactive TUI mode.
	runTUI(cfg, loop)
}

// runTUI starts the Bubble Tea TUI for interactive use.
func runTUI(cfg *runtime.Config, loop *runtime.ConversationLoop) {
	// Register slash commands (available for future non-TUI REPL mode).
	registry := commands.NewRegistry()
	commands.RegisterAuthCommands(registry)
	commands.RegisterMCPCommand(registry)
	_ = registry

	// Save session on SIGTERM (Ctrl+C is handled by Bubble Tea itself).
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	go func() {
		<-sigCh
		saveSessionSilent(cfg.SessionDir, loop)
		os.Exit(0)
	}()

	model := tui.NewModel(cfg, loop)
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // Enable mouse support (wheel, click, motion)
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}

	// Save session after the TUI exits (covers Ctrl+C via tea.Quit).
	saveSessionSilent(cfg.SessionDir, loop)
}

// saveSessionSilent saves the session, printing only to stderr on failure.
// Marshals provider-side session state before saving so delta mode can resume.
func saveSessionSilent(dir string, loop *runtime.ConversationLoop) {
	if ssp, ok := loop.Client.(api.SessionStateProvider); ok {
		loop.Session.ProviderSessionState = ssp.MarshalSessionState()
	}
	if err := runtime.SaveSession(dir, loop.Session); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save session: %v\n", err)
	}
}

// buildProvider resolves credentials via the multi-provider credential
// store, constructs the API client, and wires up any provider-specific
// tooling (e.g. DeepSeek native web search). Returns an error if no
// credentials are found or the provider cannot be created. Callers
// decide how to handle the error (TUI falls back to NoAuthClient so the
// user can /login; non-interactive subcommands should fail hard).
func buildProvider(cfg *runtime.Config) (api.APIClient, error) {
	provider, token, authMethod, credErr := auth.ResolveCredentials()
	if credErr != nil {
		return nil, fmt.Errorf("no credentials found: %w (set ANTHROPIC_API_KEY, OPENAI_API_KEY, or DEEPSEEK_TOKEN)", credErr)
	}
	cfg.ProviderName = provider
	cfg.AuthMethod = authMethod
	if authMethod == "oauth" {
		cfg.OAuthToken = token
	} else {
		cfg.APIKey = token
	}

	client, err := runtime.NewProviderClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create %s client: %w", cfg.ProviderName, err)
	}

	// Wire up DeepSeek native web search so the web_search tool uses
	// DeepSeek's server-side search (Instant model with search_enabled=true)
	// instead of Brave/DDG.
	if cfg.ProviderName == "deepseek" {
		if dsProvider, ok := client.(*deepseekprovider.Client); ok {
			tools.SetNativeSearch(func(ctx context.Context, query string, numResults int) (string, error) {
				return dsProvider.NativeSearch(ctx, query, numResults)
			})
			fmt.Fprintf(os.Stderr, "[web_search] DeepSeek native search enabled\n")
		}
	}

	return client, nil
}

// runRalphSubcommand is the entry point for `claw ralph [--spec ...]
// [--max-iterations ...]`. It wires the existing runtime loop into the
// Ralph driver, using the same credentials/provider resolution as the
// TUI and --prompt modes. Fails hard on missing credentials.
func runRalphSubcommand(args []string) {
	fs := flag.NewFlagSet("ralph", flag.ExitOnError)
	spec := fs.String("spec", runtime.DefaultRalphSpecPath, "Path to the spec/roadmap file")
	maxIter := fs.Int("max-iterations", runtime.DefaultRalphMaxIterations, "Maximum fresh-context iterations")
	selfDebug := fs.Bool("self-debug", true, "After exhausting retries, run one self-debug pass where the agent sees the last error and is tasked with fixing the root cause")
	delta := fs.Bool("delta", false, "Enable delta mode (DeepSeek only): pass cfg.DeltaMode=true to the provider. Currently a no-op for the deepseek web provider, but exposed for forward compatibility.")
	_ = fs.Parse(args)

	cfg := runtime.LoadConfig()
	// Ralph is an autonomous loop — same flags as the RunTask path.
	cfg.Autonomous = true
	if cfg.MaxTurns <= 0 {
		cfg.MaxTurns = 5 // per-iteration turn cap
	}
	if *delta {
		cfg.DeltaMode = true
	}

	provider, err := buildProvider(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	loop := runtime.NewConversationLoop(cfg, provider)

	ctx := context.Background()
	ralphCfg := runtime.DefaultRalphConfig()
	ralphCfg.SpecPath = *spec
	ralphCfg.MaxIterations = *maxIter
	ralphCfg.SelfDebug = *selfDebug

	fmt.Fprintf(os.Stderr, "[ralph] starting against %s (max %d iterations, self-debug=%v, delta=%v)\n", ralphCfg.SpecPath, ralphCfg.MaxIterations, ralphCfg.SelfDebug, cfg.DeltaMode)
	if err := runtime.RunRalphLoop(ctx, loop, ralphCfg); err != nil {
		fmt.Fprintf(os.Stderr, "[ralph] stopped: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "[ralph] done\n")
}

// runWebSubcommand starts the web server: spawns the TUI inside a PTY
// and bridges bytes to a browser via WebSocket. The browser renders
// the stream with wterm (Zig/WASM terminal emulator).
func runWebSubcommand(args []string) {
	fs := flag.NewFlagSet("web", flag.ExitOnError)
	addr := fs.String("addr", "127.0.0.1:7777", "Listen address (host:port)")
	cwd := fs.String("cwd", "", "Working directory for the spawned TUI (default: current dir)")
	provider := fs.String("provider", "", "Forward --provider to the spawned TUI")
	model := fs.String("model", "", "Forward --model to the spawned TUI")
	_ = fs.Parse(args)

	// The web subcommand itself doesn't need credentials — it just
	// spawns the TUI in a PTY and the TUI handles auth. But we
	// forward --provider/--model env-equivalents so the spawned
	// process picks them up.
	tuiArgs := []string{}
	if *provider != "" {
		tuiArgs = append(tuiArgs, "--provider", *provider)
	}
	if *model != "" {
		tuiArgs = append(tuiArgs, "--model", *model)
	}

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot resolve self path: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ctrl+C in the launching terminal: graceful shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintf(os.Stderr, "\n[web] shutting down…\n")
		cancel()
	}()

	srv := web.NewServer(web.Config{
		Addr:       *addr,
		BinaryPath: exe,
		Workdir:    *cwd,
		Args:       tuiArgs,
	})
	if err := srv.ListenAndServe(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "[web] error: %v\n", err)
		os.Exit(1)
	}
}
