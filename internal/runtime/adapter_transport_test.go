package runtime

import (
	"net/http"
	"net/http/httptest"
)

type adapterRoundTripFunc func(*http.Request) (*http.Response, error)

func (fn adapterRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func newInMemoryTestClient(handler http.Handler) *http.Client {
	return &http.Client{
		Transport: adapterRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			return rec.Result(), nil
		}),
	}
}
