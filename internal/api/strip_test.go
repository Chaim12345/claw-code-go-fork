package api

import "testing"

func TestStripToolCalls(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{
			name: "passthrough_plain_text",
			in:   "Hello, world!",
			want: "Hello, world!",
		},
		{
			name: "invoke_wrapper_with_invoke",
			in:   "I will run bash.\n<tool_calls>\n<invoke name=\"bash\">\n<parameter name=\"command\">ls</parameter>\n</invoke>\n</tool_calls>\nDone.",
			want: "I will run bash.\n\nDone.",
		},
		{
			name: "bare_invoke",
			in:   "Let me check.\n<invoke name=\"bash\">\n<parameter name=\"command\">pwd</parameter>\n</invoke>\nFinished.",
			want: "Let me check.\n\nFinished.",
		},
		{
			name: "function_wrapper",
			in:   "Writing file.\n<function_calls>\n<function name=\"write_file\">\n<parameter name=\"path\">/tmp/x</parameter>\n</function>\n</function_calls>\nOk.",
			want: "Writing file.\n\nOk.",
		},
		{
			name: "bare_function",
			in:   "Writing file.\n<function name=\"write_file\">\n<parameter name=\"path\">/tmp/x</parameter>\n</function>\nOk.",
			want: "Writing file.\n\nOk.",
		},
		{
			name: "function_call_json_body",
			in:   "Mapping.\n<function_call name=\"bash\">\n{\"command\":\"ls\"}\n</function_call>\nDone.",
			want: "Mapping.\n\nDone.",
		},
		{
			name: "bare_function_call",
			in:   "Mapping.\n<function_call name=\"bash\">\n{\"command\":\"ls\"}\n</function_call>\nDone.",
			want: "Mapping.\n\nDone.",
		},
		{
			name: "mixed_bare_and_wrapped",
			in:   "Step 1.\n<invoke name=\"read_file\">\n<parameter name=\"path\">a.go</parameter>\n</invoke>\nStep 2.\n<function name=\"write_file\">\n<parameter name=\"path\">b.go</parameter>\n</function>\nDone.",
			want: "Step 1.\n\nStep 2.\n\nDone.",
		},
		{
			name: "react_action_lines",
			in:   "Reasoning.\nAction: bash\nAction Input: {\"command\":\"ls\"}\nResult.",
			want: "Reasoning.\nResult.",
		},
		{
			name: "boundary_check_invoke_other",
			in:   "<invoke_other>x</invoke_other>\nKeep this.",
			want: "<invoke_other>x</invoke_other>\nKeep this.",
		},
		{
			name: "boundary_check_function_call_prefix",
			in:   "<tool_call>\n{\"name\":\"bash\"}\n</tool_call>\nKeep this.",
			want: "<tool_call>\n{\"name\":\"bash\"}\n</tool_call>\nKeep this.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StripToolCalls(tc.in)
			if got != tc.want {
				t.Errorf("StripToolCalls(%q)\n  got:  %q\n  want: %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestStripThenTrim(t *testing.T) {
	// The runtime applies StripToolCalls then TrimEndOfTurnIndicator
	// in sequence. This mirrors the pipeline in runOneTurnStreaming.
	cases := []struct {
		name, in, want string
	}{
		{
			name: "invoke_with_sentinel",
			in:   "Mapping.\n<invoke name=\"bash\">\n<parameter name=\"command\">ls</parameter>\n</invoke>\nAll done.FINISHED",
			want: "Mapping.\n\nAll done",
		},
		{
			name: "function_with_sentinel",
			in:   "Writing.\n<function name=\"write_file\">\n<parameter name=\"path\">x.go</parameter>\n</function>\nDone.FINISHED",
			want: "Writing.\n\nDone",
		},
		{
			name: "function_call_with_sentinel",
			in:   "Calling.\n<function_call name=\"bash\">\n{\"command\":\"ls\"}\n</function_call>\nOk.FINISHED",
			want: "Calling.\n\nOk",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TrimEndOfTurnIndicator(StripToolCalls(tc.in))
			if got != tc.want {
				t.Errorf("Strip+Trim(%q)\n  got:  %q\n  want: %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestTrimEndOfTurnIndicator(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		// The model often splices the sentinel directly onto the
		// last word with a fake sentence-end punctuation in
		// between ("...layers.FINISHED"). We strip the sentinel
		// AND the preceding . (so "...layers" remains, not
		// "...layers."). This is the documented behaviour — see
		// the docstring on TrimEndOfTurnIndicator.
		{name: "no_sentinel", in: "Hello.", want: "Hello."},
		{name: "sentinel_after_period", in: "Hello.FINISHED", want: "Hello"},
		{name: "sentinel_after_exclaim", in: "Done!FINISHED", want: "Done"},
		{name: "sentinel_after_question", in: "Ok?FINISHED", want: "Ok"},
		{name: "just_sentinel", in: "FINISHED", want: ""},
		{name: "mid_word_preserved", in: "FINISHED is a status", want: "FINISHED is a status"},
		{name: "trailing_whitespace", in: "Done.FINISHED  \n", want: "Done"},
		{name: "sentinel_after_layers", in: "...layers.FINISHED", want: "...layers"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TrimEndOfTurnIndicator(tc.in)
			if got != tc.want {
				t.Errorf("TrimEndOfTurnIndicator(%q)\n  got:  %q\n  want: %q", tc.in, got, tc.want)
			}
		})
	}
}
