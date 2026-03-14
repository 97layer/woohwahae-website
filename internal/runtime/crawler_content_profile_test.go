package runtime

import (
	"strings"
	"testing"
)

func TestCrawlerContentProfilesCanonicalKinds(t *testing.T) {
	profiles := CrawlerContentProfiles()
	if len(profiles) != 5 {
		t.Fatalf("expected 5 crawler profiles, got %d", len(profiles))
	}
	seen := map[string]bool{}
	for _, p := range profiles {
		if strings.TrimSpace(p.Kind) == "" || strings.TrimSpace(p.Topic) == "" {
			t.Fatalf("profile missing kind/topic: %+v", p)
		}
		if seen[p.Kind] {
			t.Fatalf("duplicate kind %q", p.Kind)
		}
		seen[p.Kind] = true
	}
}

func TestFindCrawlerContentProfileNormalizesKind(t *testing.T) {
	profile, ok := FindCrawlerContentProfile("Video")
	if !ok {
		t.Fatalf("expected to find video profile")
	}
	if profile.Kind != "video" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
}

func TestApplyCrawlerContentProfileSetsDefaultsAndDocID(t *testing.T) {
	profile := CrawlerContentProfiles()[0] // article
	input := ContentObservationInput{
		SourceURL: "https://example.com/story?id=123&utm_source=twitter",
	}
	updated := ApplyCrawlerContentProfile(profile, input)
	if updated.SourceChannel != "crawler" {
		t.Fatalf("expected crawler source, got %q", updated.SourceChannel)
	}
	if updated.Kind != "article" || updated.Topic != "article" {
		t.Fatalf("expected article defaults, got kind=%q topic=%q", updated.Kind, updated.Topic)
	}
	if updated.DocID == "" {
		t.Fatalf("expected docID derived from URL")
	}
}

func TestStableCrawlerDocIDStripsTrackingParams(t *testing.T) {
	urls := []string{
		"https://example.com/path/to/story?utm_source=twitter&ref=abc",
		"https://example.com/path/to/story?ref_src=google&fbclid=123",
		"https://example.com/path/to/story?si=abcd&t=45s",
		"https://example.com/path/to/story?start=45&feature=share&igshid=xyz",
		"https://example.com/path/to/story?mc_cid=foo&mc_eid=bar",
	}
	var docID string
	for i, u := range urls {
		id := stableCrawlerDocID(u)
		if id == "" {
			t.Fatalf("expected doc id for %q", u)
		}
		if i == 0 {
			docID = id
			continue
		}
		if id != docID {
			t.Fatalf("expected tracking params removed for dedupe, got %q vs %q for url %q", docID, id, u)
		}
	}
}
