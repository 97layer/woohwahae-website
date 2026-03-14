package main

import (
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type feedIngestOutput struct {
	Adapter                string                      `json:"adapter"`
	Feed                   runtime.FeedScan            `json:"feed"`
	DryRun                 bool                        `json:"dry_run"`
	CreatedObservations    []runtime.ObservationRecord `json:"created_observations"`
	CreatedSourceIntakes   []runtime.ObservationRecord `json:"created_source_intakes"`
	SkippedObservationRefs []string                    `json:"skipped_observation_refs"`
}

var fetchRSSFeedBody = fetchFeedBody

func runRSSIngest(service ingestService, args []string) {
	cmd := flag.NewFlagSet("ingest rss", flag.ExitOnError)
	feedURL := cmd.String("feed", "", "rss or atom feed url")
	limit := cmd.Int("limit", 5, "max feed items to ingest")
	channel := cmd.String("channel", "crawler", "source channel for created observations")
	kind := cmd.String("kind", "article", "content kind for created observations")
	topic := cmd.String("topic", "", "optional observation topic override")
	actor := cmd.String("actor", "rss", "actor name")
	author := cmd.String("author", "", "author override when feed items omit author")
	tags := cmd.String("tags", "", "comma-separated tags added to every item")
	refs := cmd.String("refs", "", "comma-separated refs added to every item")
	dryRun := cmd.Bool("dry-run", false, "preview feed items without writing observations")
	parseArgs(cmd, args)

	if strings.TrimSpace(*feedURL) == "" {
		log.Fatal("feed is required")
	}
	if *limit < 0 {
		log.Fatal("limit must not be negative")
	}
	if !*dryRun {
		requireCLIWriteAuth(service)
	}

	raw, err := fetchRSSFeedBody(strings.TrimSpace(*feedURL))
	if err != nil {
		log.Fatal(err)
	}
	feed, err := runtime.ParseFeedDocument(raw, strings.TrimSpace(*feedURL), *limit)
	if err != nil {
		log.Fatal(err)
	}

	output := feedIngestOutput{
		Adapter:                "rss",
		Feed:                   feed,
		DryRun:                 *dryRun,
		CreatedObservations:    []runtime.ObservationRecord{},
		CreatedSourceIntakes:   []runtime.ObservationRecord{},
		SkippedObservationRefs: []string{},
	}
	if *dryRun {
		writeJSON(output)
		return
	}

	baseRefs := splitCSV(*refs)
	baseTags := splitCSV(*tags)
	for _, item := range feed.Items {
		docID := strings.TrimSpace(item.ID)
		if docID == "" {
			docID = rssSyntheticDocID(feed.SourceURL, item)
		}
		itemTags := append([]string{}, baseTags...)
		itemTags = append(itemTags, item.Tags...)
		itemRefs := append([]string{}, baseRefs...)
		itemRefs = appendUniqueCLIString(itemRefs, "feed_kind:"+feed.Kind)
		if feed.SourceURL != "" {
			itemRefs = appendUniqueCLIString(itemRefs, "feed_source:"+feed.SourceURL)
		}

		obs := runtime.BuildContentObservation(runtime.ContentObservationInput{
			SourceChannel: strings.TrimSpace(*channel),
			Actor:         strings.TrimSpace(*actor),
			Topic:         strings.TrimSpace(*topic),
			Title:         item.Title,
			Summary:       item.Summary,
			Excerpt:       item.Excerpt,
			SourceURL:     item.Link,
			Author:        firstCLIValue(strings.TrimSpace(*author), item.Author),
			Kind:          strings.TrimSpace(*kind),
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
				output.SkippedObservationRefs = append(output.SkippedObservationRefs, dedupeRef)
				sourceObservation = existing[0]
			}
		}
		if strings.TrimSpace(sourceObservation.ObservationID) == strings.TrimSpace(obs.ObservationID) {
			created, err := service.CreateObservation(obs)
			if err != nil {
				log.Fatal(err)
			}
			sourceObservation = created
			output.CreatedObservations = append(output.CreatedObservations, created)
		}

		record := runtime.BuildSourceIntakeRecordFromContent(runtime.SourceIntakeContentInput{
			SourceObservation: sourceObservation,
			IntakeClass:       "public_collector",
			Title:             item.Title,
			URL:               item.Link,
			Excerpt:           firstCLIValue(item.Excerpt, item.Summary),
		})
		sourceIntake, created, err := runtime.EnsureSourceIntakeObservation(service, sourceObservation, record)
		if err != nil {
			log.Fatal(err)
		}
		if created {
			output.CreatedSourceIntakes = append(output.CreatedSourceIntakes, sourceIntake)
		}
	}

	writeJSON(output)
}

func fetchFeedBody(feedURL string) ([]byte, error) {
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
		return nil, logOutputError("feed fetch failed", resp.StatusCode, raw)
	}
	return raw, nil
}

func rssSyntheticDocID(feedURL string, item runtime.FeedItem) string {
	parts := []string{
		strings.TrimSpace(feedURL),
		strings.TrimSpace(item.Title),
		strings.TrimSpace(item.Link),
		item.PublishedAt.UTC().Format(time.RFC3339),
	}
	joined := strings.Join(parts, "\x00")
	sum := sha1.Sum([]byte(joined))
	return "feed_" + hex.EncodeToString(sum[:])[:12]
}

func firstCLIValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func logOutputError(message string, statusCode int, raw []byte) error {
	body := strings.TrimSpace(string(raw))
	if body == "" {
		return &feedHTTPError{Message: message, StatusCode: statusCode}
	}
	return &feedHTTPError{Message: message, StatusCode: statusCode, Body: body}
}

type feedHTTPError struct {
	Message    string
	StatusCode int
	Body       string
}

func (e *feedHTTPError) Error() string {
	if strings.TrimSpace(e.Body) == "" {
		return e.Message + " status=" + strconv.Itoa(e.StatusCode)
	}
	return e.Message + " status=" + strconv.Itoa(e.StatusCode) + " body=" + strconv.Quote(truncateFeedBody(e.Body, 180))
}

func appendUniqueCLIString(items []string, value string) []string {
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

func truncateFeedBody(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len([]rune(value)) <= max {
		return value
	}
	runes := []rune(value)
	return string(runes[:max]) + "…"
}
