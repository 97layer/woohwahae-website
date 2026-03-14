package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

const (
	defaultFeedSensorInterval = 15 * time.Minute
	defaultFeedSensorLimit    = 3
	defaultFeedPrepChannel    = "threads"
)

type feedSensorConfig struct {
	Enabled   bool
	Interval  time.Duration
	Limit     int
	AutoRoute bool
	AutoPrep  bool
	Feeds     []string
	Channel   string
	Kind      string
	Actor     string
	Author    string
	Tags      []string
	Refs      []string
}

type feedSensorTickResult struct {
	Feeds   int
	Items   int
	Created int
	Intakes int
	Routed  int
	Prepped int
	Skipped int
	Failed  int
}

var fetchDaemonFeedBody = fetchDaemonFeedHTTPBody

func loadFeedSensorConfig() feedSensorConfig {
	feeds := splitSensorValues(os.Getenv("LAYER_OS_SENSOR_RSS_FEEDS"))
	interval := parseFeedSensorInterval(os.Getenv("LAYER_OS_SENSOR_INTERVAL"))
	limit := parseFeedSensorLimit(os.Getenv("LAYER_OS_SENSOR_LIMIT"))
	channel := strings.TrimSpace(os.Getenv("LAYER_OS_SENSOR_CHANNEL"))
	if channel == "" {
		channel = "crawler"
	}
	kind := strings.TrimSpace(os.Getenv("LAYER_OS_SENSOR_KIND"))
	if kind == "" {
		kind = "article"
	}
	actor := strings.TrimSpace(os.Getenv("LAYER_OS_SENSOR_ACTOR"))
	if actor == "" {
		actor = "rss_sensor"
	}
	return feedSensorConfig{
		Enabled:   len(feeds) > 0,
		Interval:  interval,
		Limit:     limit,
		AutoRoute: parseFeedSensorBool(os.Getenv("LAYER_OS_SENSOR_AUTO_ROUTE")),
		AutoPrep:  parseFeedSensorBool(os.Getenv("LAYER_OS_SENSOR_AUTO_PREP")),
		Feeds:     feeds,
		Channel:   channel,
		Kind:      kind,
		Actor:     actor,
		Author:    strings.TrimSpace(os.Getenv("LAYER_OS_SENSOR_AUTHOR")),
		Tags:      splitSensorValues(os.Getenv("LAYER_OS_SENSOR_TAGS")),
		Refs:      splitSensorValues(os.Getenv("LAYER_OS_SENSOR_REFS")),
	}
}

func startFeedSensorLoop(ctx context.Context, service *runtime.Service) {
	config := loadFeedSensorConfig()
	if !config.Enabled {
		log.Println("feed sensor: disabled")
		return
	}
	log.Printf("feed sensor: started interval=%s limit=%d feeds=%d channel=%s kind=%s auto_route=%t auto_prep=%t", config.Interval, config.Limit, len(config.Feeds), config.Channel, config.Kind, config.AutoRoute, config.AutoPrep)
	runFeedSensorTickWithLogging(service, config)
	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("feed sensor: stopped")
			return
		case <-ticker.C:
			runFeedSensorTickWithLogging(service, config)
		}
	}
}

func runFeedSensorTickWithLogging(service *runtime.Service, config feedSensorConfig) {
	result, err := runFeedSensorTick(service, config)
	if err != nil {
		log.Printf("feed sensor: tick warning: %v", err)
	}
	if result.Created > 0 || result.Intakes > 0 || result.Routed > 0 || result.Prepped > 0 || result.Skipped > 0 || result.Failed > 0 {
		log.Printf("feed sensor: feeds=%d items=%d created=%d intakes=%d routed=%d prepped=%d skipped=%d failed=%d", result.Feeds, result.Items, result.Created, result.Intakes, result.Routed, result.Prepped, result.Skipped, result.Failed)
	}
}

func runFeedSensorTick(service *runtime.Service, config feedSensorConfig) (feedSensorTickResult, error) {
	result := feedSensorTickResult{Feeds: len(config.Feeds)}
	warnings := make([]string, 0)
	for _, feedURL := range config.Feeds {
		raw, err := fetchDaemonFeedBody(feedURL)
		if err != nil {
			result.Failed++
			warnings = append(warnings, "fetch "+feedURL+": "+err.Error())
			continue
		}
		feed, err := runtime.ParseFeedDocument(raw, feedURL, config.Limit)
		if err != nil {
			result.Failed++
			warnings = append(warnings, "parse "+feedURL+": "+err.Error())
			continue
		}
		result.Items += len(feed.Items)
		for _, item := range feed.Items {
			docID := strings.TrimSpace(item.ID)
			if docID == "" {
				docID = daemonFeedSyntheticDocID(feed.SourceURL, item)
			}
			itemTags := append([]string{}, config.Tags...)
			itemTags = append(itemTags, item.Tags...)
			itemRefs := append([]string{}, config.Refs...)
			itemRefs = appendUniqueFeedSensorString(itemRefs, "feed_kind:"+feed.Kind)
			if feed.SourceURL != "" {
				itemRefs = appendUniqueFeedSensorString(itemRefs, "feed_source:"+feed.SourceURL)
			}
			obs := runtime.BuildContentObservation(runtime.ContentObservationInput{
				SourceChannel: config.Channel,
				Actor:         config.Actor,
				Title:         item.Title,
				Summary:       item.Summary,
				Excerpt:       item.Excerpt,
				SourceURL:     item.Link,
				Author:        firstFeedSensorValue(config.Author, item.Author),
				Kind:          config.Kind,
				DocID:         docID,
				Tags:          itemTags,
				Refs:          itemRefs,
				CapturedAt:    item.PublishedAt,
			})
			sourceObservation := obs
			dedupeRef := runtime.ContentDedupeRef(docID, item.Link)
			if dedupeRef != "" {
				existing := service.ListObservations(runtime.ObservationQuery{Ref: dedupeRef, Limit: 1})
				if len(existing) > 0 {
					result.Skipped++
					sourceObservation = existing[0]
				}
			}
			if strings.TrimSpace(sourceObservation.ObservationID) == strings.TrimSpace(obs.ObservationID) {
				created, err := service.CreateObservation(obs)
				if err != nil {
					result.Failed++
					warnings = append(warnings, "create "+feedURL+": "+err.Error())
					continue
				}
				sourceObservation = created
				result.Created++
			}

			record := runtime.BuildSourceIntakeRecordFromContent(runtime.SourceIntakeContentInput{
				SourceObservation: sourceObservation,
				IntakeClass:       "public_collector",
				Title:             item.Title,
				URL:               item.Link,
				Excerpt:           firstFeedSensorValue(item.Excerpt, item.Summary),
			})
			sourceIntakeObservation, created, err := runtime.EnsureSourceIntakeObservation(service, sourceObservation, record)
			if err != nil {
				result.Failed++
				warnings = append(warnings, "source intake "+feedURL+": "+err.Error())
				continue
			} else if created {
				result.Intakes++
			}
			if config.AutoRoute {
				if target, ok := runtime.SourceIntakeAutoRouteTarget(record, sourceObservation.Refs); ok {
					lane, err := runtime.OpenSourceIntakeDraftLane(service, sourceIntakeObservation, target, config.Actor, "feed_sensor")
					if err != nil {
						result.Failed++
						warnings = append(warnings, "route "+feedURL+": "+err.Error())
						continue
					}
					result.Routed++
					if config.AutoPrep && strings.TrimSpace(lane.DraftSeed.ObservationID) != "" {
						if _, err := service.OpenSourceDraftSeedPublishPrep(lane.DraftSeed.ObservationID, defaultFeedPrepChannel); err != nil {
							result.Failed++
							warnings = append(warnings, "prep "+feedURL+": "+err.Error())
							continue
						}
						result.Prepped++
					}
				}
			}
		}
	}
	if len(warnings) == 0 {
		return result, nil
	}
	return result, &feedSensorError{Messages: warnings}
}

func parseFeedSensorInterval(raw string) time.Duration {
	if value := strings.TrimSpace(raw); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultFeedSensorInterval
}

func parseFeedSensorLimit(raw string) int {
	if value := strings.TrimSpace(raw); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			return parsed
		}
	}
	return defaultFeedSensorLimit
}

func parseFeedSensorBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func splitSensorValues(raw string) []string {
	raw = strings.ReplaceAll(raw, "\n", ",")
	raw = strings.ReplaceAll(raw, ";", ",")
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		out = appendUniqueFeedSensorString(out, value)
	}
	return out
}

func appendUniqueFeedSensorString(items []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return items
	}
	for _, item := range items {
		if strings.TrimSpace(item) == value {
			return items
		}
	}
	return append(items, value)
}

func fetchDaemonFeedHTTPBody(feedURL string) ([]byte, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(feedURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &feedSensorHTTPError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(raw))}
	}
	return raw, nil
}

func daemonFeedSyntheticDocID(feedURL string, item runtime.FeedItem) string {
	parts := []string{
		strings.TrimSpace(feedURL),
		strings.TrimSpace(item.Title),
		strings.TrimSpace(item.Link),
		item.PublishedAt.UTC().Format(time.RFC3339),
	}
	sum := sha1.Sum([]byte(strings.Join(parts, "\x00")))
	return "feed_" + hex.EncodeToString(sum[:])[:12]
}

func firstFeedSensorValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

type feedSensorHTTPError struct {
	StatusCode int
	Body       string
}

func (e *feedSensorHTTPError) Error() string {
	if strings.TrimSpace(e.Body) == "" {
		return "feed fetch failed status=" + strconv.Itoa(e.StatusCode)
	}
	return "feed fetch failed status=" + strconv.Itoa(e.StatusCode) + " body=" + strconv.Quote(truncateFeedSensorBody(e.Body, 180))
}

type feedSensorError struct {
	Messages []string
}

func (e *feedSensorError) Error() string {
	return strings.Join(e.Messages, "; ")
}

func truncateFeedSensorBody(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len([]rune(value)) <= max {
		return value
	}
	runes := []rune(value)
	return string(runes[:max]) + "…"
}
