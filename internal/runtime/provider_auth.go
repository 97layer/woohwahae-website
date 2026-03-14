package runtime

import (
	"os"
	"strings"
)

func providerCredentialEnvKeys(provider string) []string {
	switch normalizeActor(provider) {
	case "claude", "anthropic":
		return []string{"ANTHROPIC_API_KEY"}
	case "openai":
		return []string{"OPENAI_API_KEY"}
	case "gemini", "google":
		return []string{"GEMINI_API_KEY", "GOOGLE_API_KEY"}
	default:
		return nil
	}
}

func ProviderCredentialEnvKeys(provider string) []string {
	keys := providerCredentialEnvKeys(provider)
	if len(keys) == 0 {
		return []string{}
	}
	return append([]string{}, keys...)
}

func ProviderCredentialSource(provider string) (string, bool) {
	for _, key := range providerCredentialEnvKeys(provider) {
		if strings.TrimSpace(os.Getenv(key)) != "" {
			return key, true
		}
	}
	return "", false
}

func ProviderCredentialValue(provider string) (string, bool) {
	source, ok := ProviderCredentialSource(provider)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(os.Getenv(source)), true
}

func ProviderCredentialReady(provider string) bool {
	_, ok := ProviderCredentialSource(provider)
	return ok
}
