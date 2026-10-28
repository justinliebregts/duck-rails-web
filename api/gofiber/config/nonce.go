package config

import (
	"os"
)

var (
	EnvNonceSkip           = "NONCE_SKIP_PHRASE"
	defaultNonceSkipPhrase = "sample-phrase"
	nonceSkipPhrase        = defaultNonceSkipPhrase

	EnvNonceSecret     = "NONCE_SECRET"
	defaultNonceSecret = "GunRaiders@2026-05-18"
	nonceSecret        = defaultNonceSecret
)

func RetrieveNonceSkipPhrase() string {
	if nonceSkipPhrase == defaultNonceSkipPhrase {
		envVal := os.Getenv(EnvNonceSkip)
		if envVal != "" {
			nonceSkipPhrase = envVal
		}
	}
	return nonceSkipPhrase
}

func RetrieveNonceSecret() string {
	if nonceSecret == defaultNonceSecret {
		envVal := os.Getenv(EnvNonceSecret)
		if envVal != "" {
			nonceSecret = envVal
		}
	}
	return nonceSecret
}
