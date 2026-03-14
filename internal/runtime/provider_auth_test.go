package runtime

import "testing"

func TestProviderCredentialReadySupportsGeminiAlias(t *testing.T) {
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "alias_key")
	if !ProviderCredentialReady("gemini") {
		t.Fatal("expected gemini provider credential readiness via GEMINI_API_KEY")
	}
	source, ok := ProviderCredentialSource("gemini")
	if !ok || source != "GEMINI_API_KEY" {
		t.Fatalf("expected GEMINI_API_KEY auth source, got source=%q ok=%v", source, ok)
	}
}

func TestProviderCredentialEnvKeysExposeGeminiAliasesInOrder(t *testing.T) {
	keys := ProviderCredentialEnvKeys("gemini")
	if len(keys) != 2 || keys[0] != "GEMINI_API_KEY" || keys[1] != "GOOGLE_API_KEY" {
		t.Fatalf("unexpected gemini credential env keys: %+v", keys)
	}
}
