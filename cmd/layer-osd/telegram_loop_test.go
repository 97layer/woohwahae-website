package main

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestFetchUpdatesIncludesTelegramErrorDetail(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusConflict,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"ok":false,"error_code":409,"description":"Conflict: terminated by other getUpdates request"}`)),
		}, nil
	})}

	_, err := fetchUpdates(client, "https://example.test/bot", 0)
	if err == nil {
		t.Fatal("expected fetchUpdates error")
	}
	message := err.Error()
	for _, needle := range []string{"getUpdates failed", "status=409", "telegram_code=409", "other getUpdates request"} {
		if !strings.Contains(message, needle) {
			t.Fatalf("expected %q in %q", needle, message)
		}
	}
}

func TestSendMessageIncludesTelegramErrorDetail(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if !strings.Contains(string(body), "\"chat_id\":1234") {
			t.Fatalf("unexpected payload: %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"ok":false,"error_code":400,"description":"Bad Request: chat not found"}`)),
		}, nil
	})}

	err := sendMessage(client, "https://example.test/bot", 1234, "hello")
	if err == nil {
		t.Fatal("expected sendMessage error")
	}
	message := err.Error()
	for _, needle := range []string{"sendMessage failed", "status=400", "telegram_code=400", "chat not found"} {
		if !strings.Contains(message, needle) {
			t.Fatalf("expected %q in %q", needle, message)
		}
	}
}
