package runtime

import (
	"net/url"
	"strings"
)

// CrawlerContentProfile captures canonical crawler/content-capture shapes so
// helpers can stamp stable kinds, topics, dedupe hints, and tags before the
// main ingest pipeline runs.
type CrawlerContentProfile struct {
	Kind        string   // article, video, thread, post, content_capture
	Topic       string   // topic attached to the observation
	DefaultTags []string // tag:<value> refs for retrieval/ranking
	RefHints    []string // non-linkable refs that describe the interaction
	Summary     string   // human rationale for the mapping
}

// CrawlerContentProfiles lists the Phase 1.5 canonical profiles.
func CrawlerContentProfiles() []CrawlerContentProfile {
	return []CrawlerContentProfile{
		{
			Kind:        "article",
			Topic:       "article",
			DefaultTags: []string{"crawler", "article"},
			Summary:     "Long-form article; keep host/title for ranking and dedupe.",
		},
		{
			Kind:        "video",
			Topic:       "video",
			DefaultTags: []string{"crawler", "video"},
			Summary:     "Video content (e.g., YouTube); keep host and doc-id for reuse.",
		},
		{
			Kind:        "thread",
			Topic:       "thread",
			DefaultTags: []string{"crawler", "thread"},
			Summary:     "Multi-post thread; capture host/path to dedupe variations.",
		},
		{
			Kind:        "post",
			Topic:       "post",
			DefaultTags: []string{"crawler", "post"},
			Summary:     "Single social/blog post; treated separately from long articles.",
		},
		{
			Kind:        "content_capture",
			Topic:       "content_capture",
			DefaultTags: []string{"crawler", "content_capture"},
			Summary:     "Raw content capture or scrape when the kind is unclear.",
		},
	}
}

// FindCrawlerContentProfile resolves by kind (normalized).
func FindCrawlerContentProfile(kind string) (CrawlerContentProfile, bool) {
	kind = normalizeObservationMetaToken(kind)
	for _, profile := range CrawlerContentProfiles() {
		if normalizeObservationMetaToken(profile.Kind) == kind {
			return profile, true
		}
	}
	return CrawlerContentProfile{}, false
}

// ApplyCrawlerContentProfile stamps defaults and derives a stable doc_id if the
// caller did not provide one, stripping tracking parameters to improve dedupe.
func ApplyCrawlerContentProfile(profile CrawlerContentProfile, input ContentObservationInput) ContentObservationInput {
	if strings.TrimSpace(input.SourceChannel) == "" {
		input.SourceChannel = "crawler"
	}
	if strings.TrimSpace(input.Kind) == "" {
		input.Kind = profile.Kind
	}
	if strings.TrimSpace(input.Topic) == "" {
		input.Topic = profile.Topic
	}
	for _, tag := range profile.DefaultTags {
		input.Tags = appendUniqueString(input.Tags, tag)
	}
	for _, ref := range profile.RefHints {
		input.Refs = appendUniqueString(input.Refs, ref)
	}
	if strings.TrimSpace(input.DocID) == "" && strings.TrimSpace(input.SourceURL) != "" {
		if doc := stableCrawlerDocID(input.SourceURL); doc != "" {
			input.DocID = doc
		}
	}
	return input
}

// stableCrawlerDocID normalizes common tracking parameters before hashing so
// the same document shared with different utm/fbclid/ref queries dedupes.
func stableCrawlerDocID(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	query := parsed.Query()
	for key := range query {
		if isCrawlerTrackingParam(strings.ToLower(key)) {
			query.Del(key)
		}
	}
	parsed.RawQuery = query.Encode()
	parsed.Fragment = ""
	canonical := parsed.String()
	if canonical == "" {
		return ""
	}
	return "url_" + externalEvidenceHash(canonical)
}

func isCrawlerTrackingParam(key string) bool {
	if strings.HasPrefix(key, "utm_") {
		return true
	}
	switch key {
	case "fbclid", "ref", "ref_src", "feature", "si", "igshid", "mc_cid", "mc_eid", "t", "start":
		return true
	}
	return false
}
