package main

import (
	"errors"
	"testing"

	"layer-os/internal/runtime"
)

func TestRunFeedSensorTickCreatesObservationAndSkipsDuplicate(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	oldFetch := fetchDaemonFeedBody
	fetchDaemonFeedBody = func(feedURL string) ([]byte, error) {
		return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Layer Feed</title>
    <link>https://example.com/feed</link>
    <item>
      <title>First article</title>
      <link>https://example.com/post-1</link>
      <guid>post-1</guid>
      <description>First excerpt.</description>
    </item>
  </channel>
</rss>`), nil
	}
	defer func() { fetchDaemonFeedBody = oldFetch }()

	config := feedSensorConfig{
		Enabled: true,
		Limit:   3,
		Feeds:   []string{"https://example.com/feed.xml"},
		Channel: "crawler",
		Kind:    "article",
		Actor:   "rss_sensor",
		Tags:    []string{"signals"},
	}

	first, err := runFeedSensorTick(service, config)
	if err != nil {
		t.Fatalf("first feed sensor tick: %v", err)
	}
	if first.Created != 1 || first.Intakes != 1 || first.Skipped != 0 || first.Failed != 0 {
		t.Fatalf("unexpected first result: %+v", first)
	}
	items := service.ListObservations(runtime.ObservationQuery{Limit: 8})
	if len(items) != 2 {
		t.Fatalf("expected content + source intake observations, got %+v", items)
	}
	var contentObservation runtime.ObservationRecord
	var sourceIntakeObservation runtime.ObservationRecord
	for _, item := range items {
		switch item.Topic {
		case runtime.SourceIntakeTopic:
			sourceIntakeObservation = item
		default:
			contentObservation = item
		}
	}
	if contentObservation.ObservationID == "" || sourceIntakeObservation.ObservationID == "" {
		t.Fatalf("expected both observation kinds, got %+v", items)
	}
	if !containsFeedSensorRef(contentObservation.Refs, "content_doc:post_1") || !containsFeedSensorRef(contentObservation.Refs, "feed_kind:rss") {
		t.Fatalf("unexpected content observation refs: %+v", contentObservation)
	}
	if !containsFeedSensorRef(sourceIntakeObservation.Refs, "source_observation:"+contentObservation.ObservationID) || !containsFeedSensorRef(sourceIntakeObservation.Refs, "content_doc:post_1") {
		t.Fatalf("unexpected source intake refs: %+v", sourceIntakeObservation)
	}

	second, err := runFeedSensorTick(service, config)
	if err != nil {
		t.Fatalf("second feed sensor tick: %v", err)
	}
	if second.Created != 0 || second.Intakes != 0 || second.Skipped != 1 || second.Failed != 0 {
		t.Fatalf("unexpected second result: %+v", second)
	}
}

func TestRunFeedSensorTickReportsFetchFailure(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	oldFetch := fetchDaemonFeedBody
	fetchDaemonFeedBody = func(feedURL string) ([]byte, error) {
		return nil, errors.New("boom")
	}
	defer func() { fetchDaemonFeedBody = oldFetch }()

	result, err := runFeedSensorTick(service, feedSensorConfig{
		Enabled: true,
		Limit:   3,
		Feeds:   []string{"https://example.com/feed.xml"},
		Channel: "crawler",
		Kind:    "article",
		Actor:   "rss_sensor",
	})
	if err == nil {
		t.Fatal("expected feed sensor warning")
	}
	if result.Failed != 1 || result.Created != 0 {
		t.Fatalf("unexpected failure result: %+v", result)
	}
}

func TestRunFeedSensorTickAutoRoutesSingleHintedSource(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	oldFetch := fetchDaemonFeedBody
	fetchDaemonFeedBody = func(feedURL string) ([]byte, error) {
		return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Layer Feed</title>
    <link>https://example.com/feed</link>
    <item>
      <title>Quiet beauty note</title>
      <link>https://example.com/post-2</link>
      <guid>post-2</guid>
      <description>실무 기준이 먼저 남는다는 메모.</description>
    </item>
  </channel>
</rss>`), nil
	}
	defer func() { fetchDaemonFeedBody = oldFetch }()

	config := feedSensorConfig{
		Enabled:   true,
		Limit:     3,
		AutoRoute: true,
		Feeds:     []string{"https://example.com/feed.xml"},
		Channel:   "crawler",
		Kind:      "article",
		Actor:     "rss_sensor",
		Tags:      []string{"signals"},
		Refs:      []string{"route:woosunhokr"},
	}

	result, err := runFeedSensorTick(service, config)
	if err != nil {
		t.Fatalf("auto-route feed sensor tick: %v", err)
	}
	if result.Created != 1 || result.Intakes != 1 || result.Routed != 1 || result.Failed != 0 {
		t.Fatalf("unexpected auto-route result: %+v", result)
	}
	if got := len(service.ListProposals()); got != 1 {
		t.Fatalf("expected one proposal after auto-route, got %d", got)
	}
	routeDecisions := service.ListObservations(runtime.ObservationQuery{Topic: runtime.IntakeRouteDecisionTopic, Limit: 4})
	if len(routeDecisions) != 1 {
		t.Fatalf("expected one route decision observation, got %+v", routeDecisions)
	}
	draftSeeds := service.ListObservations(runtime.ObservationQuery{Topic: runtime.SourceDraftSeedTopic, Limit: 4})
	if len(draftSeeds) != 1 {
		t.Fatalf("expected one draft seed observation, got %+v", draftSeeds)
	}
}

func TestRunFeedSensorTickAutoPrepsSingleHintedSource(t *testing.T) {
	service, err := runtime.NewService(t.TempDir())
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	oldFetch := fetchDaemonFeedBody
	fetchDaemonFeedBody = func(feedURL string) ([]byte, error) {
		return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Layer Feed</title>
    <link>https://example.com/feed</link>
    <item>
      <title>Slow living note</title>
      <link>https://example.com/post-3</link>
      <guid>post-3</guid>
      <description>조용한 생활 리듬으로 옮길 만한 메모.</description>
    </item>
  </channel>
</rss>`), nil
	}
	defer func() { fetchDaemonFeedBody = oldFetch }()

	config := feedSensorConfig{
		Enabled:   true,
		Limit:     3,
		AutoRoute: true,
		AutoPrep:  true,
		Feeds:     []string{"https://example.com/feed.xml"},
		Channel:   "crawler",
		Kind:      "article",
		Actor:     "rss_sensor",
		Tags:      []string{"signals"},
		Refs:      []string{"route:woohwahae"},
	}

	result, err := runFeedSensorTick(service, config)
	if err != nil {
		t.Fatalf("auto-prep feed sensor tick: %v", err)
	}
	if result.Created != 1 || result.Intakes != 1 || result.Routed != 1 || result.Prepped != 1 || result.Failed != 0 {
		t.Fatalf("unexpected auto-prep result: %+v", result)
	}
	prepItems := service.ListObservations(runtime.ObservationQuery{Topic: "brand_publish_prep", Limit: 4})
	if len(prepItems) != 1 {
		t.Fatalf("expected one prep observation, got %+v", prepItems)
	}
	if len(service.ListApprovals()) != 1 {
		t.Fatalf("expected one approval after auto-prep, got %+v", service.ListApprovals())
	}
}

func containsFeedSensorRef(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
