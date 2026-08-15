package tools_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWatchlist_FetchesMapsDeduplicatesAndResolvesMetadata(t *testing.T) {
	const apiKey = "watchlist-secret"
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.String())
		if r.Header.Get("X-Api-Key") != apiKey {
			t.Fatalf("wrong X-Api-Key header: %q", r.Header.Get("X-Api-Key"))
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/user":
			if r.URL.RawQuery != "" || r.Header.Get("X-Api-User") != "" {
				t.Fatalf("user lookup must not include pagination or X-Api-User: %s", r.URL.String())
			}
			_, _ = w.Write([]byte(`[{"id":7,"jellyfinUserId":"AB-CD-123"}]`))
		case "/api/v1/user/7/watchlist":
			if r.Header.Get("X-Api-User") != "7" {
				t.Fatalf("wrong X-Api-User header: %q", r.Header.Get("X-Api-User"))
			}
			switch r.URL.Query().Get("page") {
			case "1":
				_, _ = w.Write([]byte(`{"page":1,"totalPages":2,"totalResults":3,"results":[{"id":3,"media":{"mediaType":"movie","tmdbId":10,"title":"Movie One","overview":"A movie","voteAverage":7.4,"releaseDate":"2021-04-05","posterPath":"/poster.jpg","backdropPath":"/backdrop.jpg","jellyfinMediaId":"jf-10"}},{"id":4,"media":{"mediaType":"movie","tmdbId":20}}]}`))
			case "2":
				_, _ = w.Write([]byte(`{"page":2,"totalPages":2,"totalResults":3,"results":[{"id":5,"media":{"mediaType":"movie","tmdbId":10,"title":"Movie One"}}]}`))
			default:
				http.Error(w, "bad page", http.StatusBadRequest)
			}
		case "/api/v1/movie/20":
			_, _ = w.Write([]byte(`{"title":"Resolved Movie","releaseDate":"2020-01-02","posterPath":"poster2.jpg"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	t.Setenv("JELLYSEERR_URL", server.URL)
	t.Setenv("JELLYSEERR_API_KEY", apiKey)

	result := callTool(t, &mockClient{getUserIDFunc: func(context.Context) (string, error) {
		return "ab-cd-123", nil
	}}, "discovery", "jellyfin_watchlist", nil)

	text := resultText(t, result)
	if !strings.Contains(text, "Resolved Movie") || !strings.Contains(text, "Movie One") {
		t.Fatalf("expected resolved and existing items, got: %s", text)
	}
	if strings.Contains(text, "watchlist-secret") {
		t.Fatal("API key leaked into tool output")
	}
	if !strings.Contains(text, "Jellyseerr watchlist (2 items, 3 raw results, 2 pages)") {
		t.Fatalf("expected deduplicated count of 2, got: %s", text)
	}
	if !strings.Contains(text, "https://image.tmdb.org/t/p/w500/poster.jpg") || !strings.Contains(text, "/Items/jf-10/Images/Primary") {
		t.Fatalf("expected normalized image URLs, got: %s", text)
	}
	if len(requests) != 4 {
		t.Fatalf("expected user lookup, two pages, and detail lookup; got %v", requests)
	}
}

func TestWatchlist_UserLookupAuthErrorIsRedacted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "authorization watchlist-secret", http.StatusForbidden)
	}))
	defer server.Close()
	t.Setenv("JELLYSEERR_URL", server.URL)
	t.Setenv("JELLYSEERR_API_KEY", "watchlist-secret")

	result := callTool(t, &mockClient{}, "discovery", "jellyfin_watchlist", nil)
	text := resultText(t, result)
	if !result.IsError {
		t.Fatal("expected an error result")
	}
	if !strings.Contains(text, "Unable to resolve the Jellyseerr user") {
		t.Fatalf("expected clear mapping error, got: %s", text)
	}
	if strings.Contains(text, "watchlist-secret") || strings.Contains(text, "authorization") {
		t.Fatalf("authorization details leaked: %s", text)
	}
}

func TestWatchlistPaginationRejectsRawCountMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/v1/user" {
			_, _ = w.Write([]byte(`[{"id":7,"jellyfinUserId":"test-user-id"}]`))
			return
		}
		_, _ = w.Write([]byte(`{"page":1,"totalPages":1,"totalResults":2,"results":[{"id":1,"media":{"mediaType":"movie","tmdbId":1}}]}`))
	}))
	defer server.Close()
	t.Setenv("JELLYSEERR_URL", server.URL)
	t.Setenv("JELLYSEERR_API_KEY", "secret")

	result := callTool(t, &mockClient{}, "discovery", "jellyfin_watchlist", nil)
	text := resultText(t, result)
	if !result.IsError || !strings.Contains(text, "pagination is incomplete") {
		t.Fatalf("expected incomplete pagination error, got: %s", text)
	}
}

func TestWatchlistClientRejectsMissingConfiguration(t *testing.T) {
	t.Setenv("JELLYSEERR_URL", "")
	t.Setenv("JELLYSEERR_API_KEY", "")
	result := callTool(t, &mockClient{}, "discovery", "jellyfin_watchlist", nil)
	if !strings.Contains(resultText(t, result), "JELLYSEERR_URL") {
		t.Fatalf("expected missing configuration error, got: %s", resultText(t, result))
	}
}
