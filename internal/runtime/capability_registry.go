package runtime

import (
	"errors"
	"os"
	"sort"
	"strings"
)

type providerEnvFlags struct {
	explicit  bool
	roleBound bool
}

type CapabilityBinding struct {
	ID           string   `json:"id"`
	Kind         string   `json:"kind"`
	Source       string   `json:"source"`
	Active       bool     `json:"active"`
	Removable    bool     `json:"removable"`
	Default      bool     `json:"default"`
	Semantics    string   `json:"semantics,omitempty"`
	RequiredMode string   `json:"required_mode,omitempty"`
	Notes        []string `json:"notes,omitempty"`
}

type CapabilityRegistry struct {
	Authority    AuthorityBoundary   `json:"authority"`
	DefaultActor string              `json:"default_actor"`
	Bindings     []CapabilityBinding `json:"bindings"`
}

func splitCapabilityEnv(raw string) []string {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	items := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		value := normalizeActor(part)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		items = append(items, value)
	}
	return items
}

func capabilityActorBindings() []CapabilityBinding {
	defaultActor := DefaultActor()
	bindings := []CapabilityBinding{}
	seen := map[string]struct{}{}
	add := func(id, source string, isDefault bool) {
		id = normalizeActor(id)
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		bindings = append(bindings, CapabilityBinding{
			ID:        id,
			Kind:      "actor",
			Source:    source,
			Active:    true,
			Removable: id != neutralDefaultActor,
			Default:   isDefault,
		})
	}
	for _, actor := range splitCapabilityEnv(os.Getenv("LAYER_OS_ACTORS")) {
		add(actor, "env.actors", actor == defaultActor)
	}
	add(os.Getenv("LAYER_OS_ACTOR"), "env.actor", normalizeActor(os.Getenv("LAYER_OS_ACTOR")) == defaultActor)
	add(os.Getenv("LAYER_OS_WRITER_ID"), "env.writer", normalizeActor(os.Getenv("LAYER_OS_WRITER_ID")) == defaultActor)
	source := "runtime.default_actor"
	if normalizeActor(os.Getenv("LAYER_OS_DEFAULT_ACTOR")) != "" {
		source = "env.default_actor"
	}
	add(defaultActor, source, true)
	return bindings
}

func capabilityProviderBindings() []CapabilityBinding {
	bindings := []CapabilityBinding{}
	for _, provider := range sortedProviderEnvRoster() {
		flags := providerEnvRoster()[provider]
		source := "env.providers"
		notes := []string{
			"declared provider roster only",
			"live provider dispatch remains disabled until a dispatch adapter is enabled",
		}
		switch {
		case flags.explicit && flags.roleBound:
			source = "env.providers+role_providers"
			notes = append(notes, "provider also implied by role provider routing")
		case flags.roleBound:
			source = "env.role_providers"
			notes = append(notes, "provider implied by role provider routing")
		}
		bindings = append(bindings, CapabilityBinding{
			ID:        provider,
			Kind:      "provider",
			Source:    source,
			Active:    true,
			Removable: true,
			Default:   false,
			Semantics: "declared_only",
			Notes:     notes,
		})
	}
	return bindings
}

func capabilityAdapterBindings(adapters AdapterSummary) []CapabilityBinding {
	bindings := []CapabilityBinding{
		{
			ID:           adapters.Gateway,
			Kind:         "gateway_adapter",
			Source:       "runtime.adapter",
			Active:       true,
			Removable:    true,
			Default:      true,
			Semantics:    adapters.GatewaySemantics,
			RequiredMode: adapters.GatewayRequiredMode,
			Notes:        []string{"dispatch_enabled=" + strings.ToLower(strconvFormatBool(adapters.GatewayDispatchEnabled))},
		},
		{
			ID:        adapters.Verify,
			Kind:      "verify_adapter",
			Source:    "runtime.adapter",
			Active:    true,
			Removable: true,
			Default:   true,
			Semantics: "command",
		},
		{
			ID:        adapters.Deploy,
			Kind:      "deploy_adapter",
			Source:    "runtime.adapter",
			Active:    true,
			Removable: true,
			Default:   true,
			Semantics: "command",
		},
		{
			ID:        adapters.Rollback,
			Kind:      "rollback_adapter",
			Source:    "runtime.adapter",
			Active:    true,
			Removable: true,
			Default:   true,
			Semantics: "command",
		},
	}
	return bindings
}

func strconvFormatBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func capabilityRegistryFromAdapters(adapters AdapterSummary) CapabilityRegistry {
	bindings := append([]CapabilityBinding{}, capabilityActorBindings()...)
	bindings = append(bindings, capabilityProviderBindings()...)
	bindings = append(bindings, capabilityAdapterBindings(adapters)...)
	return CapabilityRegistry{
		Authority:    CanonicalAuthorityBoundary(),
		DefaultActor: DefaultActor(),
		Bindings:     bindings,
	}
}

func (b CapabilityBinding) Validate() error {
	if strings.TrimSpace(b.ID) == "" {
		return errors.New("capability binding id is required")
	}
	if strings.TrimSpace(b.Kind) == "" {
		return errors.New("capability binding kind is required")
	}
	if strings.TrimSpace(b.Source) == "" {
		return errors.New("capability binding source is required")
	}
	return nil
}

func (r CapabilityRegistry) Validate() error {
	if err := r.Authority.Validate(); err != nil {
		return err
	}
	if strings.TrimSpace(r.DefaultActor) == "" {
		return errors.New("capability registry default_actor is required")
	}
	if r.Bindings == nil {
		return errors.New("capability registry bindings is required")
	}
	for _, binding := range r.Bindings {
		if err := binding.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Capabilities() CapabilityRegistry {
	return capabilityRegistryFromAdapters(s.Adapters())
}

func providerEndpointMapFromEnv() map[string]string {
	items := map[string]string{}
	for _, part := range strings.Split(strings.TrimSpace(os.Getenv("LAYER_OS_PROVIDER_ENDPOINTS")), ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		pieces := strings.SplitN(part, "=", 2)
		if len(pieces) != 2 {
			continue
		}
		provider := normalizeActor(strings.TrimSpace(pieces[0]))
		endpoint := strings.TrimSpace(pieces[1])
		if provider == "" || endpoint == "" {
			continue
		}
		if err := validateProviderEndpoint(endpoint); err != nil {
			continue
		}
		items[provider] = endpoint
	}
	return items
}

func providerSummaries(adapters AdapterSummary) []ProviderSummary {
	roster := providerEnvRoster()
	endpoints := providerEndpointMapFromEnv()
	all := map[string]struct{}{}
	implicit := map[string]bool{}
	for provider := range roster {
		all[provider] = struct{}{}
	}
	for provider := range endpoints {
		all[provider] = struct{}{}
	}
	if adapters.GatewaySemantics == "direct_llm" {
		provider := normalizeActor(adapters.Gateway)
		if councilProviderAllowed(provider) {
			all[provider] = struct{}{}
			implicit[provider] = true
		}
	}
	items := make([]ProviderSummary, 0, len(all))
	for provider := range all {
		var endpoint *string
		if raw, ok := endpoints[provider]; ok && strings.TrimSpace(raw) != "" {
			value := raw
			endpoint = &value
		}
		authKeys := ProviderCredentialEnvKeys(provider)
		authSourceValue, authReady := ProviderCredentialSource(provider)
		var authSource *string
		if authReady {
			value := authSourceValue
			authSource = &value
		}
		notes := []string{}
		if roster[provider].explicit {
			notes = append(notes, "declared provider roster")
		}
		if roster[provider].roleBound {
			notes = append(notes, "bound through role provider routing")
		}
		if implicit[provider] {
			notes = append(notes, "inferred from active gateway adapter")
		}
		if endpoint != nil {
			notes = append(notes, "api endpoint configured")
		}
		if len(authKeys) > 0 {
			if authReady {
				notes = append(notes, "provider credentials ready via "+authSourceValue)
			} else {
				notes = append(notes, "provider credentials missing; accepted envs: "+strings.Join(authKeys, ", "))
			}
		}
		if !adapters.GatewayDispatchEnabled {
			notes = append(notes, "gateway adapter is not dispatching")
		}
		items = append(items, ProviderSummary{
			Provider:        provider,
			Endpoint:        endpoint,
			Declared:        roster[provider].explicit || roster[provider].roleBound,
			DispatchEnabled: adapters.GatewayDispatchEnabled,
			AuthReady:       authReady,
			AuthSource:      authSource,
			AuthEnvKeys:     authKeys,
			GatewayAdapter:  adapters.Gateway,
			Semantics:       adapters.GatewaySemantics,
			Notes:           notes,
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Provider < items[j].Provider })
	return items
}

func providerEnvRoster() map[string]providerEnvFlags {
	roster := map[string]providerEnvFlags{}
	for _, provider := range splitCapabilityEnv(os.Getenv("LAYER_OS_PROVIDERS")) {
		flags := roster[provider]
		flags.explicit = true
		roster[provider] = flags
	}
	for _, provider := range roleBoundProvidersFromEnv(os.Getenv("LAYER_OS_AGENT_ROLE_PROVIDERS")) {
		flags := roster[provider]
		flags.roleBound = true
		roster[provider] = flags
	}
	return roster
}

func sortedProviderEnvRoster() []string {
	roster := providerEnvRoster()
	items := make([]string, 0, len(roster))
	for provider := range roster {
		items = append(items, provider)
	}
	sort.Strings(items)
	return items
}

func roleBoundProvidersFromEnv(raw string) []string {
	items := []string{}
	seen := map[string]struct{}{}
	for _, provider := range splitRoleBindingEnv(raw) {
		normalized := normalizeActor(provider)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	sort.Strings(items)
	return items
}
