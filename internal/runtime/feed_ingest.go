package runtime

import (
	"bytes"
	"encoding/xml"
	"errors"
	"strings"
	"time"
)

type FeedScan struct {
	Kind      string     `json:"kind"`
	Title     string     `json:"title"`
	Link      string     `json:"link,omitempty"`
	SourceURL string     `json:"source_url"`
	Items     []FeedItem `json:"items"`
}

type FeedItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Link        string    `json:"link,omitempty"`
	Summary     string    `json:"summary,omitempty"`
	Excerpt     string    `json:"excerpt,omitempty"`
	Author      string    `json:"author,omitempty"`
	PublishedAt time.Time `json:"published_at,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
}

type rssDocument struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title string    `xml:"title"`
	Link  string    `xml:"link"`
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title          string   `xml:"title"`
	Link           string   `xml:"link"`
	Description    string   `xml:"description"`
	GUID           string   `xml:"guid"`
	PubDate        string   `xml:"pubDate"`
	Author         string   `xml:"author"`
	Categories     []string `xml:"category"`
	ContentEncoded string   `xml:"encoded"`
}

type atomDocument struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	Links   []atomLink  `xml:"link"`
	Entries []atomEntry `xml:"entry"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type atomEntry struct {
	ID         string          `xml:"id"`
	Title      string          `xml:"title"`
	Summary    string          `xml:"summary"`
	Content    string          `xml:"content"`
	Updated    string          `xml:"updated"`
	Published  string          `xml:"published"`
	Links      []atomLink      `xml:"link"`
	Author     atomEntryAuthor `xml:"author"`
	Categories []atomCategory  `xml:"category"`
}

type atomEntryAuthor struct {
	Name string `xml:"name"`
}

type atomCategory struct {
	Term string `xml:"term,attr"`
}

func ParseFeedDocument(raw []byte, sourceURL string, limit int) (FeedScan, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return FeedScan{}, errors.New("feed body is empty")
	}
	if limit < 0 {
		return FeedScan{}, errors.New("limit must not be negative")
	}

	var root struct {
		XMLName xml.Name
	}
	if err := xml.Unmarshal(raw, &root); err != nil {
		return FeedScan{}, err
	}

	switch strings.ToLower(strings.TrimSpace(root.XMLName.Local)) {
	case "rss":
		return parseRSSFeed(raw, sourceURL, limit)
	case "feed":
		return parseAtomFeed(raw, sourceURL, limit)
	default:
		return FeedScan{}, errors.New("unsupported feed format")
	}
}

func parseRSSFeed(raw []byte, sourceURL string, limit int) (FeedScan, error) {
	var doc rssDocument
	if err := xml.Unmarshal(raw, &doc); err != nil {
		return FeedScan{}, err
	}
	items := make([]FeedItem, 0, len(doc.Channel.Items))
	for _, item := range doc.Channel.Items {
		excerpt := strings.TrimSpace(item.Description)
		if excerpt == "" {
			excerpt = strings.TrimSpace(item.ContentEncoded)
		}
		feedItem := FeedItem{
			ID:          strings.TrimSpace(item.GUID),
			Title:       strings.TrimSpace(item.Title),
			Link:        strings.TrimSpace(item.Link),
			Summary:     limitText(strings.TrimSpace(item.Title), 180),
			Excerpt:     strings.TrimSpace(excerpt),
			Author:      strings.TrimSpace(item.Author),
			PublishedAt: parseFeedTime(item.PubDate),
			Tags:        normalizeFeedTags(item.Categories),
		}
		if feedItem.Summary == "" {
			feedItem.Summary = externalObservationSummary(feedItem.Excerpt, "rss item")
		}
		items = append(items, feedItem)
		if limit > 0 && len(items) >= limit {
			break
		}
	}
	return FeedScan{
		Kind:      "rss",
		Title:     strings.TrimSpace(doc.Channel.Title),
		Link:      strings.TrimSpace(doc.Channel.Link),
		SourceURL: strings.TrimSpace(sourceURL),
		Items:     items,
	}, nil
}

func parseAtomFeed(raw []byte, sourceURL string, limit int) (FeedScan, error) {
	var doc atomDocument
	if err := xml.Unmarshal(raw, &doc); err != nil {
		return FeedScan{}, err
	}
	items := make([]FeedItem, 0, len(doc.Entries))
	for _, entry := range doc.Entries {
		excerpt := strings.TrimSpace(entry.Summary)
		if excerpt == "" {
			excerpt = strings.TrimSpace(entry.Content)
		}
		feedItem := FeedItem{
			ID:          strings.TrimSpace(entry.ID),
			Title:       strings.TrimSpace(entry.Title),
			Link:        firstAtomLink(entry.Links),
			Summary:     limitText(strings.TrimSpace(entry.Title), 180),
			Excerpt:     strings.TrimSpace(excerpt),
			Author:      strings.TrimSpace(entry.Author.Name),
			PublishedAt: parseFeedTime(firstNonEmpty(entry.Published, entry.Updated)),
			Tags:        atomCategoryTerms(entry.Categories),
		}
		if feedItem.Summary == "" {
			feedItem.Summary = externalObservationSummary(feedItem.Excerpt, "atom entry")
		}
		items = append(items, feedItem)
		if limit > 0 && len(items) >= limit {
			break
		}
	}
	return FeedScan{
		Kind:      "atom",
		Title:     strings.TrimSpace(doc.Title),
		Link:      firstAtomLink(doc.Links),
		SourceURL: strings.TrimSpace(sourceURL),
		Items:     items,
	}, nil
}

func parseFeedTime(raw string) time.Time {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}
	}
	for _, layout := range []string{
		time.RFC3339,
		time.RFC3339Nano,
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC850,
		time.ANSIC,
	} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}

func normalizeFeedTags(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		value := normalizeObservationMetaToken(item)
		if value == "" {
			continue
		}
		out = appendUniqueString(out, value)
	}
	return out
}

func atomCategoryTerms(items []atomCategory) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		value := normalizeObservationMetaToken(item.Term)
		if value == "" {
			continue
		}
		out = appendUniqueString(out, value)
	}
	return out
}

func firstAtomLink(items []atomLink) string {
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.Rel), "alternate") && strings.TrimSpace(item.Href) != "" {
			return strings.TrimSpace(item.Href)
		}
	}
	for _, item := range items {
		if strings.TrimSpace(item.Href) != "" {
			return strings.TrimSpace(item.Href)
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
