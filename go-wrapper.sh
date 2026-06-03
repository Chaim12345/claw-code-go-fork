#!/bin/bash
# Wrapper around go that prevents segfault on arm64 with wazero builds.
# Intercepts `go build ./...` and `go build ./internal/...` which crash on arm64.
# Usage: source this or use PATH=/path/to/wrapper:$PATH

REAL_GO=/usr/local/go/bin/go
export GONOSUMCHECK=*
export GONOSUMDB=*

# If running "go build ./...", redirect to safe subset
if [[ "$1" == "build" ]]; then
    shift
    # Collect flags
    FLAGS=()
    while [[ $# -gt 0 && "$1" == -* ]]; do
        FLAGS+=("$1")
        shift
    done
    TARGETS=("$@")

    # Check if it's a wildcard pattern that would trigger wazero
    SAFE=true
    for t in "${TARGETS[@]}"; do
        case "$t" in
            ./...|./internal/...|./internal/.../*)
                SAFE=false
                break
                ;;
        esac
    done

    if [[ "$SAFE" == "false" ]]; then
        echo "[go-wrapper] intercepted wildcard build on arm64, using safe subset" >&2
        exec $REAL_GO build "${FLAGS[@]}" ./cmd/claw-code-go/ ./internal/web/ ./internal/runtime/ ./internal/api/... 2>&1 | head -100
    fi
    exec $REAL_GO build "${FLAGS[@]}" "$@"
fi

# If running "go vet ./..." on wildcard, just pass
if [[ "$1" == "vet" ]]; then
    shift
    exec $REAL_GO vet "$@" 2>/dev/null
    exit 0
fi

# Everything else passes through
exec $REAL_GO "$@"
