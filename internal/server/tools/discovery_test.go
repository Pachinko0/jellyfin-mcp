package tools_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestSearch_HappyPath(t *testing.T) {
	mc := &mockClient{
		getFunc: func(_ context.Context, endpoint string, _ url.Values, dest any) error {
			if strings.HasPrefix(endpoint, "/Users/") && strings.HasSuffix(endpoint, "/Items") {
				return jsonInto(map[string]any{
					"Items": []map[string]any{
						{"Id": "item-1", "Name": "Neon Cascade", "Type": "Movie", "ProductionYear": 1999},
						{"Id": "item-2", "Name": "Neon Cascade Reloaded", "Type": "Movie", "ProductionYear": 2003},
					},
					"TotalRecordCount": 2,
				}, dest)
			}
			return nil
		},
	}

	result := callTool(t, mc, "", "jellyfin_search", map[string]any{
		"query": "Neon",
	})

	text := resultText(t, result)
	if !strings.Contains(text, "Neon Cascade") {
		t.Errorf("expected result to contain 'Neon Cascade', got: %s", text)
	}
	if !strings.Contains(text, "2 results") {
		t.Errorf("expected result to mention '2 results', got: %s", text)
	}
}

func TestSearch_EmptyResults(t *testing.T) {
	mc := &mockClient{
		getFunc: func(_ context.Context, _ string, _ url.Values, dest any) error {
			return jsonInto(map[string]any{
				"Items":            []map[string]any{},
				"TotalRecordCount": 0,
			}, dest)
		},
	}

	result := callTool(t, mc, "", "jellyfin_search", map[string]any{
		"query": "nonexistent",
	})

	text := resultText(t, result)
	if !strings.Contains(text, "0 results") {
		t.Errorf("expected '0 results', got: %s", text)
	}
}

func TestSearch_APIError(t *testing.T) {
	mc := &mockClient{
		getFunc: func(_ context.Context, endpoint string, _ url.Values, _ any) error {
			if strings.Contains(endpoint, "/Items") {
				return fmt.Errorf("API error 500: Internal Server Error")
			}
			return nil
		},
	}

	result := callTool(t, mc, "", "jellyfin_search", map[string]any{
		"query": "test",
	})

	text := resultText(t, result)
	if !strings.Contains(text, "error") && !strings.Contains(text, "Error") {
		t.Errorf("expected error message, got: %s", text)
	}
	if !result.IsError {
		t.Error("expected IsError to be true")
	}
}

func TestLibraries_HappyPath(t *testing.T) {
	mc := &mockClient{
		getFunc: func(_ context.Context, endpoint string, _ url.Values, dest any) error {
			if endpoint == "/Library/VirtualFolders" {
				return jsonInto([]map[string]any{
					{
						"Name":           "Movies",
						"ItemId":         "lib-1",
						"CollectionType": "movies",
						"Locations":      []string{"/media/movies"},
					},
					{
						"Name":           "TV Shows",
						"ItemId":         "lib-2",
						"CollectionType": "tvshows",
						"Locations":      []string{"/media/tv"},
					},
				}, dest)
			}
			return nil
		},
	}

	result := callTool(t, mc, "", "jellyfin_libraries", nil)

	text := resultText(t, result)
	if !strings.Contains(text, "Movies") {
		t.Errorf("expected 'Movies' in result, got: %s", text)
	}
	if !strings.Contains(text, "2 libraries") {
		t.Errorf("expected '2 libraries' in result, got: %s", text)
	}
}

func TestGetItem_HappyPath(t *testing.T) {
	mc := &mockClient{
		getFunc: func(_ context.Context, endpoint string, _ url.Values, dest any) error {
			if strings.Contains(endpoint, "/Items/") {
				return jsonInto(map[string]any{
					"Id":              "item-1",
					"Name":            "Lucid Horizon",
					"Type":            "Movie",
					"ProductionYear":  2010,
					"Overview":        "A rogue architect builds impossible worlds inside shared dreams.",
					"Genres":          []string{"Action", "Sci-Fi"},
					"CommunityRating": 8.8,
				}, dest)
			}
			return nil
		},
	}

	result := callTool(t, mc, "", "jellyfin_get_item", map[string]any{
		"id": "item-1",
	})

	text := resultText(t, result)
	if !strings.Contains(text, "Lucid Horizon") {
		t.Errorf("expected 'Lucid Horizon' in result, got: %s", text)
	}
}

func TestGetItem_NotFound(t *testing.T) {
	mc := &mockClient{
		getFunc: func(_ context.Context, endpoint string, _ url.Values, _ any) error {
			if strings.Contains(endpoint, "/Items/") {
				return fmt.Errorf("API error 404: Not Found")
			}
			return nil
		},
	}

	result := callTool(t, mc, "", "jellyfin_get_item", map[string]any{
		"id": "nonexistent-id",
	})

	text := resultText(t, result)
	if !strings.Contains(text, "error") && !strings.Contains(text, "Error") {
		t.Errorf("expected error message, got: %s", text)
	}
	if !result.IsError {
		t.Error("expected IsError to be true")
	}
}

func TestMediaActions_DownloadURLUsesCurrentAPIKeyQueryName(t *testing.T) {
	result := callTool(t, &mockClient{
		baseURLVal: "http://jellyfin.test",
		apiKeyVal:  "secret",
	}, "discovery", "jellyfin_item_extras", map[string]any{
		"action":  "download_url",
		"item_id": "item-1",
	})

	text := resultText(t, result)
	if !strings.Contains(text, "/Items/item-1/Download?ApiKey=secret") {
		t.Fatalf("expected current ApiKey query parameter, got: %s", text)
	}
	if strings.Contains(text, "api_key=") {
		t.Fatalf("deprecated api_key query parameter is still present: %s", text)
	}
}
