package runtime

import "testing"

func TestCapabilityRegistryUsesNeutralDefaultAndAdapterBindings(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	t.Setenv("LAYER_OS_PROVIDER_ENDPOINTS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	registry := service.Capabilities()
	if err := registry.Validate(); err != nil {
		t.Fatalf("validate registry: %v", err)
	}
	if registry.Authority.LegacyReferenceMode != authorityLegacyReferenceMode || registry.Authority.ExternalJobReportSurface != "/api/layer-os/jobs/report" {
		t.Fatalf("unexpected authority boundary: %+v", registry.Authority)
	}
	if registry.DefaultActor != "system" {
		t.Fatalf("expected neutral default actor system, got %q", registry.DefaultActor)
	}
	var actorFound bool
	var gatewayFound bool
	for _, binding := range registry.Bindings {
		switch binding.Kind {
		case "actor":
			if binding.ID == "system" {
				actorFound = true
				if !binding.Default || binding.Removable {
					t.Fatalf("unexpected system actor binding: %+v", binding)
				}
			}
		case "gateway_adapter":
			gatewayFound = true
			if binding.ID != "record" || binding.Semantics != "record_only" || binding.RequiredMode != "single" {
				t.Fatalf("unexpected gateway binding: %+v", binding)
			}
		}
	}
	if !actorFound {
		t.Fatalf("expected neutral actor binding in %+v", registry.Bindings)
	}
	if !gatewayFound {
		t.Fatalf("expected gateway adapter binding in %+v", registry.Bindings)
	}
}

func TestCapabilityRegistrySurfacesDeclaredActorsAndProviders(t *testing.T) {
	t.Setenv("LAYER_OS_ACTORS", "claude,gemini")
	t.Setenv("LAYER_OS_DEFAULT_ACTOR", "claude")
	t.Setenv("LAYER_OS_PROVIDERS", "openai,anthropic")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	registry := service.Capabilities()
	if registry.DefaultActor != "claude" {
		t.Fatalf("expected default actor claude, got %q", registry.DefaultActor)
	}
	found := map[string]bool{}
	for _, binding := range registry.Bindings {
		found[binding.Kind+":"+binding.ID] = true
		if binding.Kind == "provider" && binding.Semantics != "declared_only" {
			t.Fatalf("unexpected provider semantics: %+v", binding)
		}
	}
	for _, key := range []string{"actor:claude", "actor:gemini", "provider:openai", "provider:anthropic"} {
		if !found[key] {
			t.Fatalf("expected binding %s in %+v", key, registry.Bindings)
		}
	}
}

func TestCapabilityRegistrySurfacesRoleBoundProviders(t *testing.T) {
	t.Setenv("LAYER_OS_PROVIDERS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "planner=claude,implementer=openai,verifier=claude")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	registry := service.Capabilities()
	found := map[string]CapabilityBinding{}
	for _, binding := range registry.Bindings {
		if binding.Kind == "provider" {
			found[binding.ID] = binding
		}
	}
	if len(found) != 2 {
		t.Fatalf("expected 2 role-bound provider bindings, got %+v", found)
	}
	for _, provider := range []string{"claude", "openai"} {
		binding, ok := found[provider]
		if !ok {
			t.Fatalf("expected provider binding %q in %+v", provider, registry.Bindings)
		}
		if binding.Source != "env.role_providers" || binding.Semantics != "declared_only" {
			t.Fatalf("unexpected role-bound provider binding: %+v", binding)
		}
	}
}

func TestProviderSummariesInferActiveDirectLLMAdapter(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "claude")
	t.Setenv("LAYER_OS_PROVIDERS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	providers := service.Providers()
	if len(providers) != 1 || providers[0].Provider != "claude" {
		t.Fatalf("expected implicit claude provider summary, got %+v", providers)
	}
	found := false
	for _, note := range providers[0].Notes {
		if note == "inferred from active gateway adapter" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected inferred-provider note, got %+v", providers[0].Notes)
	}
}

func TestProviderSummariesIncludeRoleBoundProviders(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "")
	t.Setenv("LAYER_OS_PROVIDERS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "planner=claude,designer=openai")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	providers := service.Providers()
	if len(providers) != 2 || providers[0].Provider != "claude" || providers[1].Provider != "openai" {
		t.Fatalf("expected sorted role-bound providers, got %+v", providers)
	}
	for _, provider := range providers {
		if !provider.Declared {
			t.Fatalf("expected role-bound provider to be treated as declared, got %+v", provider)
		}
		found := false
		for _, note := range provider.Notes {
			if note == "bound through role provider routing" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected role-bound note on provider summary, got %+v", provider)
		}
	}
}

func TestProviderSummariesSurfaceAuthAliasesAndSource(t *testing.T) {
	t.Setenv("LAYER_OS_GATEWAY_ADAPTER", "")
	t.Setenv("LAYER_OS_PROVIDERS", "")
	t.Setenv("LAYER_OS_AGENT_ROLE_PROVIDERS", "designer=gemini")
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("GEMINI_API_KEY", "test_key")
	service, err := NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	providers := service.Providers()
	if len(providers) != 1 || providers[0].Provider != "gemini" {
		t.Fatalf("expected gemini provider summary, got %+v", providers)
	}
	if !providers[0].AuthReady {
		t.Fatalf("expected gemini auth to be ready, got %+v", providers[0])
	}
	if providers[0].AuthSource == nil || *providers[0].AuthSource != "GEMINI_API_KEY" {
		t.Fatalf("expected GEMINI_API_KEY auth source, got %+v", providers[0])
	}
	if len(providers[0].AuthEnvKeys) != 2 || providers[0].AuthEnvKeys[0] != "GEMINI_API_KEY" || providers[0].AuthEnvKeys[1] != "GOOGLE_API_KEY" {
		t.Fatalf("expected gemini auth aliases, got %+v", providers[0].AuthEnvKeys)
	}
}
