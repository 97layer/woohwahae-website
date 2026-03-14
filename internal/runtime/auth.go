package runtime

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

type authConfig struct {
	WriteTokenHash    string            `json:"write_token_hash"`
	ScopedTokenHashes map[string]string `json:"scoped_token_hashes,omitempty"`
	UpdatedAt         *time.Time        `json:"updated_at,omitempty"`
}

type authStore struct {
	mu     sync.RWMutex
	config authConfig
}

func newAuthStore(config authConfig) *authStore {
	return &authStore{config: config}
}

func authEnabled(config authConfig) bool {
	if strings.TrimSpace(config.WriteTokenHash) != "" {
		return true
	}
	for _, hash := range config.ScopedTokenHashes {
		if strings.TrimSpace(hash) != "" {
			return true
		}
	}
	return false
}

func (s *authStore) current() authConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *authStore) status() AuthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return AuthStatus{
		WriteAuthEnabled: authEnabled(s.config),
		UpdatedAt:        s.config.UpdatedAt,
	}
}

func (s *authStore) setToken(token string, actor string, updatedAt time.Time) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("write token is required")
	}
	actor = ResolveActor(actor)
	s.mu.Lock()
	defer s.mu.Unlock()
	if actor == "founder" {
		s.config.WriteTokenHash = hashToken(token)
	} else {
		if s.config.ScopedTokenHashes == nil {
			s.config.ScopedTokenHashes = map[string]string{}
		}
		s.config.ScopedTokenHashes[actor] = hashToken(token)
	}
	s.config.UpdatedAt = &updatedAt
	return nil
}

func (s *authStore) clear(actor string, updatedAt time.Time) {
	actor = ResolveActor(actor)
	s.mu.Lock()
	defer s.mu.Unlock()
	if actor == "founder" {
		s.config.WriteTokenHash = ""
	} else if s.config.ScopedTokenHashes != nil {
		delete(s.config.ScopedTokenHashes, actor)
		if len(s.config.ScopedTokenHashes) == 0 {
			s.config.ScopedTokenHashes = nil
		}
	}
	s.config.UpdatedAt = &updatedAt
}

func (s *authStore) authorize(token string, actor string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !authEnabled(s.config) {
		return true
	}
	sum := hashToken(strings.TrimSpace(token))
	actor = ResolveActor(actor)
	if scoped := strings.TrimSpace(s.config.ScopedTokenHashes[actor]); scoped != "" {
		if subtle.ConstantTimeCompare([]byte(sum), []byte(scoped)) == 1 {
			return true
		}
	}
	if strings.TrimSpace(s.config.WriteTokenHash) == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(sum), []byte(s.config.WriteTokenHash)) == 1
}

func defaultAuthConfig() authConfig {
	return authConfig{}
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func GenerateToken() (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
