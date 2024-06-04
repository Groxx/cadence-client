package debug

import "go.uber.org/cadence/internal"

// deprecated: unstable API
func CachedWIDs() []string {
	return internal.GetCachedWorkflows()
}
