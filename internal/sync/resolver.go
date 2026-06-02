package sync

import (
	"context"
	"fmt"
)

// DefaultResolver implements basic conflict resolution strategies
type DefaultResolver struct {
	strategy ConflictResolutionStrategy
}

// NewDefaultResolver creates a new default resolver
func NewDefaultResolver(strategy ConflictResolutionStrategy) *DefaultResolver {
	return &DefaultResolver{
		strategy: strategy,
	}
}

// Resolve resolves a conflict between local and remote sessions
func (r *DefaultResolver) Resolve(ctx context.Context, local, remote SessionData) (SessionData, error) {
	switch r.strategy {
	case StrategyNewest:
		return r.resolveNewest(local, remote)

	case StrategyLocal:
		return local, nil

	case StrategyRemote:
		return remote, nil

	case StrategyMerge:
		return SessionData{}, fmt.Errorf("merge strategy not implemented yet")

	case StrategyAsk:
		return SessionData{}, fmt.Errorf("ask strategy requires user interaction")

	default:
		return SessionData{}, fmt.Errorf("unknown conflict resolution strategy: %v", r.strategy)
	}
}

// resolveNewest keeps the most recently modified version
func (r *DefaultResolver) resolveNewest(local, remote SessionData) (SessionData, error) {
	if local.Meta.LastModified.After(remote.Meta.LastModified) {
		return local, nil
	}
	return remote, nil
}

// InteractiveResolver prompts the user to choose which version to keep
type InteractiveResolver struct {
	promptFunc func(local, remote SessionData) (SessionData, error)
}

// NewInteractiveResolver creates a new interactive resolver
func NewInteractiveResolver(promptFunc func(local, remote SessionData) (SessionData, error)) *InteractiveResolver {
	return &InteractiveResolver{
		promptFunc: promptFunc,
	}
}

// Resolve prompts the user to resolve the conflict
func (r *InteractiveResolver) Resolve(ctx context.Context, local, remote SessionData) (SessionData, error) {
	if r.promptFunc == nil {
		return SessionData{}, fmt.Errorf("no prompt function configured")
	}

	return r.promptFunc(local, remote)
}

// MergeResolver attempts to merge both versions (future implementation)
type MergeResolver struct {
	// Future: implement intelligent merging of session messages
}

// NewMergeResolver creates a new merge resolver
func NewMergeResolver() *MergeResolver {
	return &MergeResolver{}
}

// Resolve attempts to merge local and remote sessions
func (r *MergeResolver) Resolve(ctx context.Context, local, remote SessionData) (SessionData, error) {
	// TODO: Implement intelligent merging
	// For now, fall back to newest strategy
	if local.Meta.LastModified.After(remote.Meta.LastModified) {
		return local, nil
	}
	return remote, nil
}
