package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestCheckClassifiesHTTPAndTransportFailures(t *testing.T) {
	for _, status := range []int{200, 404, 500, 503} {
		client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader("[]"))}, nil
		})}
		ok, _ := check(context.Background(), client, "http://backend/api/articles")
		if ok != (status == 200) {
			t.Fatalf("status %d classified as %t", status, ok)
		}
	}
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	})}
	if ok, _ := check(context.Background(), client, "http://backend/api/articles"); ok {
		t.Fatal("transport failure classified as success")
	}
}

func TestCheckTimeoutCoversResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	client := &http.Client{Timeout: 50 * time.Millisecond}
	if ok, _ := check(context.Background(), client, server.URL); ok {
		t.Fatal("response body timeout classified as success")
	}
}

func TestCheckDoesNotAcceptRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	if ok, _ := check(context.Background(), client, server.URL); ok {
		t.Fatal("redirect classified as success")
	}
}
