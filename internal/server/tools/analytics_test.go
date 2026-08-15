package tools_test

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestRecentlyAdded_UsesCurrentRouteAndClientSideDateCreatedFilter(t *testing.T) {
	var requestedEndpoint string
	var requestedParams url.Values
	now := time.Now().UTC()
	mc := &mockClient{
		getFunc: func(_ context.Context, endpoint string, params url.Values, dest any) error {
			requestedEndpoint = endpoint
			requestedParams = params
			return jsonInto(map[string]any{
				"Items": []map[string]any{
					{"Id": "new-movie", "Name": "New Movie", "Type": "Movie", "DateCreated": now.Add(-2 * time.Hour).Format(time.RFC3339Nano), "PremiereDate": now.Add(-48 * time.Hour).Format(time.RFC3339Nano)},
					{"Id": "new-series", "Name": "New Series", "Type": "Series", "DateCreated": now.Add(-23 * time.Hour).Format(time.RFC3339Nano)},
					{"Id": "old-movie", "Name": "Old Movie", "Type": "Movie", "DateCreated": now.Add(-25 * time.Hour).Format(time.RFC3339Nano)},
					{"Id": "missing-date", "Name": "Missing Date", "Type": "Movie", "PremiereDate": now.Format(time.RFC3339Nano)},
				},
				"TotalRecordCount": 4,
			}, dest)
		},
	}

	result := callTool(t, mc, "analytics", "jellyfin_analytics", map[string]any{
		"action": "recently_added",
		"days":   1,
	})

	if requestedEndpoint != "/Users/test-user-id/Items" {
		t.Fatalf("expected current user items route, got %q", requestedEndpoint)
	}
	if strings.Contains(requestedEndpoint, "/Users/test-user-id/Items/") {
		t.Fatalf("unexpected item-detail route used for collection query: %q", requestedEndpoint)
	}
	if requestedParams.Get("MinDateCreated") != "" {
		t.Fatal("recently_added must not rely on MinDateCreated")
	}
	if requestedParams.Get("SortBy") != "DateCreated" || requestedParams.Get("SortOrder") != "Descending" {
		t.Fatalf("expected DateCreated descending sort, got %v", requestedParams)
	}
	if requestedParams.Get("Fields") == "" || !strings.Contains(requestedParams.Get("Fields"), "DateCreated") {
		t.Fatalf("DateCreated was not explicitly requested: %v", requestedParams)
	}
	if requestedParams.Get("IncludeItemTypes") != "Movie,Series" {
		t.Fatalf("expected Movie,Series default filter, got %q", requestedParams.Get("IncludeItemTypes"))
	}

	text := resultText(t, result)
	if !strings.Contains(text, "New Movie") || !strings.Contains(text, "New Series") {
		t.Fatalf("expected recent Movie and Series items, got: %s", text)
	}
	if strings.Contains(text, "Old Movie") || strings.Contains(text, "Missing Date") {
		t.Fatalf("old or timestamp-less item was returned: %s", text)
	}
}
