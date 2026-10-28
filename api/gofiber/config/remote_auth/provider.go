package remote_auth

import (
	"os"
	"sync"
)

var (
	remoteAuthSkipKey   = "REMOTE_AUTHORIZATION_SKIP"
	remoteAuthSkipPhase = ""

	once sync.Once
)

func init() {
	_ = loadRemoteAuthSkip(false)
}

func IsRemoteAuthSkipped(phase string) bool {

	// Immediately reject empty input
	if phase == "" {
		return false
	}

	// Reject if the server's secret isn't configured/is empty
	if remoteAuthSkipPhase == "" {
		return false
	}

	// Perform the comparison
	return phase == remoteAuthSkipPhase
}

func loadRemoteAuthSkip(forceReload bool) string {
	if forceReload {
		once = sync.Once{}
	}
	once.Do(func() {
		remoteAuthSkipPhase = os.Getenv(remoteAuthSkipKey)
	})
	return remoteAuthSkipPhase
}
