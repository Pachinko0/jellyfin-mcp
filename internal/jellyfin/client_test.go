package jellyfin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestDoRequestUsesCurrentAuthorizationSchema(t *testing.T) {
	const apiKey = "0123456789abcdef0123456789abcdef"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != `MediaBrowser Token="`+apiKey+`"` {
			t.Errorf("Authorization header = %q, want MediaBrowser token schema", got)
		}
		if got := r.Header.Get("X-MediaBrowser-Token"); got != "" {
			t.Errorf("legacy X-MediaBrowser-Token header was sent: %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &JellyfinClient{
		baseURL:    server.URL,
		apiKey:     apiKey,
		httpClient: server.Client(),
	}
	if _, err := client.DoRequest(context.Background(), http.MethodGet, "/System/Info", url.Values{}, nil); err != nil {
		t.Fatalf("DoRequest() error = %v", err)
	}
}

func TestPostRawUsesCurrentAuthorizationSchema(t *testing.T) {
	const apiKey = "0123456789abcdef0123456789abcdef"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != `MediaBrowser Token="`+apiKey+`"` {
			t.Errorf("Authorization header = %q, want MediaBrowser token schema", got)
		}
		if got := r.Header.Get("X-MediaBrowser-Token"); got != "" {
			t.Errorf("legacy X-MediaBrowser-Token header was sent: %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &JellyfinClient{
		baseURL:    server.URL,
		apiKey:     apiKey,
		httpClient: server.Client(),
	}
	if err := client.PostRaw(context.Background(), "/Items/1/Images/Primary", nil, []byte("image"), "image/jpeg"); err != nil {
		t.Fatalf("PostRaw() error = %v", err)
	}
}
