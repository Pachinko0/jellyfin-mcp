package tools

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	jf "github.com/jaredtrent/jellyfin-mcp/internal/jellyfin"
)

const maxJellyseerrWatchlistPages = 1000

func RegisterWatchlistTools(server *mcp.Server, client jf.Client, enabled func(string, *mcp.ToolAnnotations) bool) {
	if !enabled("jellyfin_watchlist", AnnotReadOnly) {
		return
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:  "jellyfin_watchlist",
		Title: "Jellyseerr Watchlist",
		Description: "Read the current Jellyseerr watchlist for the active Jellyfin user. " +
			"This is the Jellyseerr watchlist, not Jellyfin favorites. Results include compact metadata, availability in Jellyfin, and image URLs.",
		Annotations: AnnotReadOnly,
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ jf.NoInput) (*mcp.CallToolResult, *jf.WatchlistOutput, error) {
		seerr, err := jf.NewJellyseerrClient()
		if err != nil {
			return jf.ErrResult("Jellyseerr watchlist is unavailable: %v", err), nil, nil
		}
		jellyfinUserID, err := client.GetUserID(ctx)
		if err != nil || strings.TrimSpace(jellyfinUserID) == "" {
			return jf.ErrResult("Unable to identify the current Jellyfin user."), nil, nil
		}
		seerrUserID, err := resolveJellyseerrUser(ctx, seerr, jellyfinUserID)
		if err != nil {
			return jf.ErrResult("Unable to resolve the Jellyseerr user. Verify the Jellyseerr API key permissions and server compatibility."), nil, nil
		}
		rawItems, page, totalPages, totalResults, err := fetchJellyseerrWatchlist(ctx, seerr, seerrUserID)
		if err != nil {
			return jf.ErrResult("Unable to retrieve the Jellyseerr watchlist: %v", err), nil, nil
		}
		items := make([]map[string]any, 0, len(rawItems))
		seen := make(map[string]struct{}, len(rawItems))
		for i, raw := range rawItems {
			item, key := buildWatchlistItem(ctx, seerr, client, raw, i)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			items = append(items, item)
		}
		output := &jf.WatchlistOutput{TotalCount: len(items), Page: page, TotalPages: totalPages, Items: items}
		message := fmt.Sprintf("Jellyseerr watchlist (%d items, %d raw results, %d pages):\n\n%s", len(items), totalResults, totalPages, jf.FormatJSON(items))
		return jf.TextResult(message), output, nil
	})
}

func resolveJellyseerrUser(ctx context.Context, seerr *jf.JellyseerrClient, jellyfinUserID string) (string, error) {
	var raw any
	if err := seerr.Get(ctx, "/api/v1/user", nil, "", &raw); err != nil {
		return "", err
	}
	for _, user := range jellyseerrUsers(raw) {
		if normalizeUserID(jf.GetString(user, "jellyfinUserId")) != normalizeUserID(jellyfinUserID) {
			continue
		}
		id := stringOrNumber(user["id"])
		if id == "" {
			return "", fmt.Errorf("matched Jellyseerr user has no internal id")
		}
		return id, nil
	}
	return "", fmt.Errorf("no Jellyseerr user matches the current Jellyfin user")
}

func jellyseerrUsers(raw any) []map[string]any {
	if users := jf.ToSlice(raw); users != nil {
		return mapSlice(users)
	}
	m := jf.ToMap(raw)
	if m == nil {
		return nil
	}
	for _, key := range []string{"users", "results", "items"} {
		if users := jf.ToSlice(m[key]); users != nil {
			return mapSlice(users)
		}
	}
	if _, ok := m["jellyfinUserId"]; ok {
		return []map[string]any{m}
	}
	return nil
}

func mapSlice(values []any) []map[string]any {
	result := make([]map[string]any, 0, len(values))
	for _, value := range values {
		if m := jf.ToMap(value); m != nil {
			result = append(result, m)
		}
	}
	return result
}

func normalizeUserID(value string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), "-", ""))
}

func fetchJellyseerrWatchlist(ctx context.Context, seerr *jf.JellyseerrClient, userID string) ([]any, int, int, int, error) {
	all := make([]any, 0)
	page := 1
	totalPages := 0
	totalResults := -1
	for {
		if page < 1 || page > maxJellyseerrWatchlistPages {
			return nil, 0, 0, 0, fmt.Errorf("invalid or excessive pagination")
		}
		var response map[string]any
		params := url.Values{"page": {strconv.Itoa(page)}}
		endpoint := fmt.Sprintf("/api/v1/user/%s/watchlist", url.PathEscape(userID))
		if err := seerr.Get(ctx, endpoint, params, userID, &response); err != nil {
			return nil, 0, 0, 0, err
		}
		responsePage, ok := intField(response, "page")
		if !ok || responsePage < 1 || responsePage != page {
			return nil, 0, 0, 0, fmt.Errorf("Jellyseerr returned an unexpected page")
		}
		responsePages, ok := intField(response, "totalPages")
		if !ok || responsePages < 1 || responsePages > maxJellyseerrWatchlistPages {
			return nil, 0, 0, 0, fmt.Errorf("Jellyseerr returned invalid totalPages")
		}
		responseTotal, ok := intField(response, "totalResults")
		if !ok || responseTotal < 0 {
			return nil, 0, 0, 0, fmt.Errorf("Jellyseerr returned invalid totalResults")
		}
		if totalPages == 0 {
			totalPages = responsePages
			totalResults = responseTotal
		} else if totalPages != responsePages || totalResults != responseTotal {
			return nil, 0, 0, 0, fmt.Errorf("Jellyseerr pagination metadata changed between pages")
		}
		results, exists := response["results"]
		if !exists {
			return nil, 0, 0, 0, fmt.Errorf("Jellyseerr response omitted results")
		}
		pageResults := jf.ToSlice(results)
		all = append(all, pageResults...)
		if page == totalPages {
			break
		}
		page++
	}
	if len(all) != totalResults {
		return nil, 0, 0, 0, fmt.Errorf("Jellyseerr pagination is incomplete")
	}
	return all, 1, totalPages, totalResults, nil
}

func buildWatchlistItem(ctx context.Context, seerr *jf.JellyseerrClient, client jf.Client, raw any, index int) (map[string]any, string) {
	entry := jf.ToMap(raw)
	media := jf.ToMap(entry["media"])
	if media == nil {
		media = entry
	}
	mediaType := normalizeMediaType(firstString(media, entry, "mediaType", "type"))
	tmdbID, hasTMDBID := firstInt(media, entry, "tmdbId", "tmdbID")
	name := mediaName(media, entry, mediaType)
	if name == "" && hasTMDBID && mediaType != "" {
		endpointType := "tv"
		if mediaType == "Movie" {
			endpointType = "movie"
		}
		var details map[string]any
		if err := seerr.Get(ctx, fmt.Sprintf("/api/v1/%s/%d", endpointType, tmdbID), nil, "", &details); err == nil {
			if name == "" {
				name = mediaName(details, nil, mediaType)
			}
			mergeMissing(media, details)
		}
	}
	created := firstString(media, entry, "watchlistAddedAt", "createdAt", "created_at", "addedAt")
	poster := firstString(media, entry, "posterPath", "poster_path")
	backdrop := firstString(media, entry, "backdropPath", "backdrop_path")
	jellyfinID := firstString(media, entry, "jellyfinMediaId")
	item := map[string]any{
		"name":               nullableString(name),
		"type":               nullableString(mediaType),
		"year":               nullableInt(yearFrom(media, entry, mediaType)),
		"tmdb_id":            nullableIntValue(tmdbID, hasTMDBID),
		"overview":           nullableString(firstString(media, entry, "overview")),
		"community_rating":   nullableNumber(media, entry, "voteAverage", "vote_average", "communityRating"),
		"watchlist_id":       nullableNumber(media, entry, "watchlistId", "id"),
		"in_jellyfin":        jellyfinID != "",
		"jellyfin_id":        nullableString(jellyfinID),
		"media_status":       nullableNumber(media, entry, "status", "mediaStatus"),
		"watchlist_added_at": nullableString(created),
		"image_url":          imageURL("https://image.tmdb.org/t/p/w500", poster),
		"backdrop_url":       imageURL("https://image.tmdb.org/t/p/w1280", backdrop),
		"jellyfin_image_url": jellyfinImageURL(client.BaseURL(), jellyfinID),
	}
	keyID := "missing"
	if hasTMDBID {
		keyID = strconv.Itoa(tmdbID)
	} else {
		keyID = fmt.Sprintf("missing-%d", index)
	}
	return item, strings.ToLower(mediaType) + ":" + keyID
}

func normalizeMediaType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "movie", "movies":
		return "Movie"
	case "tv", "series", "show", "shows":
		return "Series"
	default:
		return ""
	}
}

func mediaName(media, entry map[string]any, mediaType string) string {
	if mediaType == "Movie" {
		if value := firstString(media, entry, "title"); value != "" {
			return value
		}
	} else if mediaType == "Series" {
		if value := firstString(media, entry, "name"); value != "" {
			return value
		}
	}
	return firstString(media, entry, "title", "name")
}

func yearFrom(media, entry map[string]any, mediaType string) int {
	dateKeys := []string{"releaseDate", "release_date"}
	if mediaType == "Series" {
		dateKeys = []string{"firstAirDate", "first_air_date"}
	}
	for _, key := range dateKeys {
		if value := firstString(media, entry, key); len(value) >= 4 {
			if year, err := strconv.Atoi(value[:4]); err == nil {
				return year
			}
		}
	}
	return 0
}

func mergeMissing(target, source map[string]any) {
	for key, value := range source {
		if _, exists := target[key]; !exists || target[key] == nil || target[key] == "" {
			target[key] = value
		}
	}
}

func firstString(primary, secondary map[string]any, keys ...string) string {
	for _, source := range []map[string]any{primary, secondary} {
		for _, key := range keys {
			if value, ok := source[key].(string); ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}

func firstInt(primary, secondary map[string]any, keys ...string) (int, bool) {
	for _, source := range []map[string]any{primary, secondary} {
		for _, key := range keys {
			if value, ok := intValue(source[key]); ok {
				return value, true
			}
		}
	}
	return 0, false
}

func intField(source map[string]any, key string) (int, bool) { return intValue(source[key]) }

func intValue(value any) (int, bool) {
	switch value := value.(type) {
	case int:
		return value, true
	case int64:
		return int(value), true
	case float64:
		return int(value), value == float64(int(value))
	case string:
		parsed, err := strconv.Atoi(value)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func stringOrNumber(value any) string {
	if value, ok := value.(string); ok {
		return value
	}
	if number, ok := intValue(value); ok {
		return strconv.Itoa(number)
	}
	return ""
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableInt(value int) any {
	if value == 0 {
		return nil
	}
	return value
}

func nullableIntValue(value int, ok bool) any {
	if !ok {
		return nil
	}
	return value
}

func nullableNumber(primary, secondary map[string]any, keys ...string) any {
	for _, source := range []map[string]any{primary, secondary} {
		for _, key := range keys {
			if value, ok := source[key]; ok {
				switch value.(type) {
				case float64, float32, int, int64, int32, string:
					return value
				}
			}
		}
	}
	return nil
}

func imageURL(base, path string) any {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}

func jellyfinImageURL(base, itemID string) any {
	if itemID == "" || strings.TrimSpace(base) == "" {
		return nil
	}
	return strings.TrimRight(base, "/") + "/Items/" + url.PathEscape(itemID) + "/Images/Primary"
}
