package runtime

import (
	"errors"
	"net"
	"net/netip"
	"net/url"
	"strings"
)

var errProviderEndpointRejected = errors.New("provider endpoint rejected by egress policy")

func validateProviderEndpoint(raw string) error {
	value := strings.TrimSpace(raw)
	if value == "" {
		return errors.New("provider endpoint is required")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return errProviderEndpointRejected
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errProviderEndpointRejected
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errProviderEndpointRejected
	}
	host := strings.TrimSpace(parsed.Hostname())
	if host == "" {
		return errProviderEndpointRejected
	}
	lowerHost := strings.ToLower(host)
	if lowerHost == "localhost" || strings.HasSuffix(lowerHost, ".localhost") || strings.HasSuffix(lowerHost, ".local") || strings.HasSuffix(lowerHost, ".internal") {
		return errProviderEndpointRejected
	}
	if addr, err := netip.ParseAddr(lowerHost); err == nil {
		if addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsMulticast() || addr.IsUnspecified() {
			return errProviderEndpointRejected
		}
		return nil
	}
	if ip := net.ParseIP(lowerHost); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified() {
			return errProviderEndpointRejected
		}
	}
	return nil
}
