package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScopedWriteTokenAuthorizesMatchingActorOnly(t *testing.T) {
	dataDir := t.TempDir()
	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if err := service.WithActor("gemini", func(s *Service) error {
		_, err := s.SetWriteToken("gemini-secret")
		return err
	}); err != nil {
		t.Fatalf("set scoped token: %v", err)
	}

	if !service.AuthStatus().WriteAuthEnabled {
		t.Fatal("expected scoped token to enable write auth")
	}
	if !service.AuthorizeWriteTokenForActor("gemini-secret", "gemini") {
		t.Fatal("expected matching actor token to authorize")
	}
	if service.AuthorizeWriteTokenForActor("gemini-secret", "codex") {
		t.Fatal("expected mismatched actor token to fail")
	}

	disk, err := newDiskStore(dataDir)
	if err != nil {
		t.Fatalf("new disk store: %v", err)
	}
	raw, err := os.ReadFile(disk.authPath())
	if err != nil {
		t.Fatalf("read auth config: %v", err)
	}
	text := string(raw)
	if strings.Contains(text, "gemini-secret") {
		t.Fatalf("expected auth config to avoid plaintext token, got %s", text)
	}
	if !strings.Contains(text, "scoped_token_hashes") {
		t.Fatalf("expected scoped token hash persistence, got %s", text)
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if !reloaded.AuthStatus().WriteAuthEnabled {
		t.Fatal("expected scoped token to keep write auth enabled after reload")
	}
	if !reloaded.AuthorizeWriteTokenForActor("gemini-secret", "gemini") {
		t.Fatal("expected reloaded scoped token to authorize matching actor")
	}
}

func TestAuthStoragePathStaysLocalForEphemeralRuntimeData(t *testing.T) {
	dataDir := t.TempDir()
	disk, err := newDiskStore(dataDir)
	if err != nil {
		t.Fatalf("new disk store: %v", err)
	}
	if got := disk.authPath(); got != filepath.Join(dataDir, "auth.json") {
		t.Fatalf("expected temp auth path to stay local, got %q", got)
	}
}

func TestAuthStoragePathMovesPersistentRuntimeOutsideDataDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	baseDir := filepath.Clean(filepath.Join(os.TempDir(), "..", "layer-os-persistent"))
	path := resolveAuthStoragePath(baseDir)
	if pathWithinDir(path, baseDir) {
		t.Fatalf("expected persistent auth path outside runtime data dir, got %q", path)
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("user config dir: %v", err)
	}
	expectedDir := filepath.Join(configDir, "layer-os", "auth")
	if !pathWithinDir(path, expectedDir) {
		t.Fatalf("expected auth path under %q, got %q", expectedDir, path)
	}
}

func TestLoadAuthConfigMigratesLegacyAuthConfigToExternalPath(t *testing.T) {
	dataDir := t.TempDir()
	externalPath := filepath.Join(t.TempDir(), "auth", "write-auth.json")
	t.Setenv("LAYER_OS_AUTH_FILE", externalPath)

	legacy := authConfig{WriteTokenHash: hashToken("migrate-secret")}
	if err := writeJSONAtomic(filepath.Join(dataDir, "auth.json"), legacy); err != nil {
		t.Fatalf("write legacy auth: %v", err)
	}

	service, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if !service.AuthorizeWriteTokenForActor("migrate-secret", "founder") {
		t.Fatal("expected migrated auth config to authorize founder token")
	}

	raw, err := os.ReadFile(externalPath)
	if err != nil {
		t.Fatalf("read external auth config: %v", err)
	}
	if !strings.Contains(string(raw), legacy.WriteTokenHash) {
		t.Fatalf("expected migrated auth hash in external auth config, got %s", string(raw))
	}

	reloaded, err := NewService(dataDir)
	if err != nil {
		t.Fatalf("reload service: %v", err)
	}
	if !reloaded.AuthorizeWriteTokenForActor("migrate-secret", "founder") {
		t.Fatal("expected reloaded service to use migrated external auth config")
	}
}

func TestClearingFounderTokenKeepsScopedAuthEnabled(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if err := service.WithActor("founder", func(s *Service) error {
		_, err := s.SetWriteToken("founder-secret")
		return err
	}); err != nil {
		t.Fatalf("set founder token: %v", err)
	}
	if err := service.WithActor("gemini", func(s *Service) error {
		_, err := s.SetWriteToken("gemini-secret")
		return err
	}); err != nil {
		t.Fatalf("set gemini token: %v", err)
	}

	var status AuthStatus
	if err := service.WithActor("founder", func(s *Service) error {
		var clearErr error
		status, clearErr = s.ClearWriteToken()
		return clearErr
	}); err != nil {
		t.Fatalf("clear founder token: %v", err)
	}

	if !status.WriteAuthEnabled {
		t.Fatal("expected scoped auth to remain enabled after founder token clear")
	}
	if service.AuthorizeWriteTokenForActor("founder-secret", "codex") {
		t.Fatal("expected cleared founder token to stop authorizing")
	}
	if !service.AuthorizeWriteTokenForActor("gemini-secret", "gemini") {
		t.Fatal("expected scoped gemini token to remain valid")
	}
}

func TestNormalizeVerificationCommandRejectsRelativeExecutable(t *testing.T) {
	if _, err := normalizeVerificationCommand([]string{"go", "test", "./..."}); err == nil {
		t.Fatal("expected relative verification executable to be rejected")
	}
}

func TestDefaultGoTestVerificationCommandUsesAbsoluteBinary(t *testing.T) {
	command, err := DefaultGoTestVerificationCommand()
	if err != nil {
		t.Fatalf("default verification command: %v", err)
	}
	if len(command) != 3 {
		t.Fatalf("expected 3 command parts, got %+v", command)
	}
	if !filepath.IsAbs(command[0]) {
		t.Fatalf("expected absolute go binary, got %q", command[0])
	}
	if command[1] != "test" || command[2] != "./..." {
		t.Fatalf("unexpected default verification command: %+v", command)
	}
}

func TestValidateProviderEndpointRejectsLocalAndPrivateHosts(t *testing.T) {
	cases := []string{
		"ftp://example.com/v1/respond",
		"http://localhost:8080/v1/respond",
		"https://api.local/v1/respond",
		"https://service.internal/v1/respond",
		"http://127.0.0.1:8080/v1/respond",
		"https://10.0.0.8/v1/respond",
		"https://192.168.1.5/v1/respond",
		"https://[::1]/v1/respond",
		"https://user:pass@provider.example/v1/respond",
		"https://provider.example/v1/respond?token=secret",
	}
	for _, endpoint := range cases {
		if err := validateProviderEndpoint(endpoint); err == nil {
			t.Fatalf("expected provider endpoint rejection for %q", endpoint)
		}
	}
}

func TestValidateProviderEndpointAllowsPublicHTTPSHost(t *testing.T) {
	if err := validateProviderEndpoint("https://provider.example/v1/respond"); err != nil {
		t.Fatalf("expected public provider endpoint to be allowed, got %v", err)
	}
}

func TestNewDiskStoreNormalizesBaseDir(t *testing.T) {
	root := t.TempDir()
	disk, err := newDiskStore(filepath.Join(root, "runtime", "..", "runtime"))
	if err != nil {
		t.Fatalf("new disk store: %v", err)
	}
	if !filepath.IsAbs(disk.baseDir) {
		t.Fatalf("expected absolute disk base dir, got %q", disk.baseDir)
	}
	if strings.Contains(disk.baseDir, "..") {
		t.Fatalf("expected normalized disk base dir, got %q", disk.baseDir)
	}
}

func TestCapitalizationRejectsTraversalEntryID(t *testing.T) {
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	if err := service.DebugAppendCapitalizationEntry(CapitalizationEntry{EntryID: "../escape"}); err == nil {
		t.Fatal("expected capitalization traversal entry id to be rejected")
	}
}
