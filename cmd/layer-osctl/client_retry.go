package main

import (
	"net/http"
	"time"
)

var daemonReadRetryAttempts = 3
var daemonReadRetryDelay = 150 * time.Millisecond
var daemonRetrySleep = time.Sleep
var daemonStreamReconnectAttempts = 2

func (c *daemonClient) doRequest(method string, path string, payload any) (*http.Response, error) {
	attempts := 1
	if method == http.MethodGet {
		attempts = daemonReadRetryAttempts
	}
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		request, err := c.newRequest(method, path, payload)
		if err != nil {
			return nil, err
		}
		httpClient := c.httpClient
		timeout := daemonRequestTimeout(method, path)
		if httpClient == nil {
			httpClient = &http.Client{Timeout: timeout}
		} else if timeout > 0 && httpClient.Timeout != timeout {
			cloned := *httpClient
			cloned.Timeout = timeout
			httpClient = &cloned
		}
		response, err := httpClient.Do(request)
		if err == nil {
			return response, nil
		}
		lastErr = wrapDaemonError(err)
		if method != http.MethodGet || !isDaemonUnavailable(lastErr) || attempt == attempts {
			return nil, lastErr
		}
		daemonRetrySleep(daemonReadRetryDelay)
	}
	return nil, lastErr
}

func streamConnectAttempts() int {
	attempts := daemonStreamReconnectAttempts
	if attempts < 1 {
		return 1
	}
	return attempts
}
