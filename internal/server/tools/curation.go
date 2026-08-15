package tools

import (
	"context"
	"fmt"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	jf "github.com/jaredtrent/jellyfin-mcp/internal/jellyfin"
)

func fetchLocalCandidates(ctx context.Context, client jf.Client, userID, mediaTypes string, maxRuntime int) ([]jf.RecommendationCandidate, error) {
	params := url.Values{
		"UserId":           {userID},
		"IncludeItemTypes": {mediaTypes},
		"Recursive":        {"true"},
		"Fields":           {"Genres,People,UserData,CommunityRating,RunTimeTicks,ProductionYear"},
	}
	rawItems, _, err := jf.FetchAllPages(ctx, client, fmt.Sprintf("/Users/%s/Items", jf.SanitizeID(userID)), params, 5000)
	if err != nil {
		return nil, err
	}
	candidates := make([]jf.RecommendationCandidate, 0, len(rawItems))
	for _, raw := range rawItems {
		item := jf.ParseRecommendationCandidate(jf.ToMap(raw))
		if maxRuntime > 0 && item.RuntimeMinutes > maxRuntime {
			continue
		}
		candidates = append(candidates, item)
	}
	return candidates, nil
}

func rankedResultItems(ranked []jf.RankedRecommendation) []map[string]any {
	items := make([]map[string]any, 0, len(ranked))
	for _, result := range ranked {
		item := jf.ExtractMediaItem(result.Candidate.Raw)
		item["score"] = result.Score
		item["reasons"] = result.Reasons
		items = append(items, item)
	}
	return items
}

func registerPersonalizedRecommendation(args jf.RecommendationsInput, ctx context.Context, client jf.Client, userID string) (*mcp.CallToolResult, *jf.RecommendationsOutput, error) {
	includePlayed := args.IncludePlayed != nil && *args.IncludePlayed
	candidates, err := fetchLocalCandidates(ctx, client, userID, "Movie,Series", args.MaxRuntime)
	if err != nil {
		return jf.ErrResult("Jellyfin API error: %v", err), nil, nil
	}
	profile := jf.BuildPreferenceProfile(candidates)
	ranked := jf.RankRecommendations(candidates, profile, args.Query, args.MaxRuntime, includePlayed, jf.ClampInt(args.Limit, 25, jf.MaxLimitCap))
	items := rankedResultItems(ranked)
	return jf.TextResult(fmt.Sprintf("Personalized recommendations (%d):\n\n%s", len(items), jf.FormatJSON(items))), &jf.RecommendationsOutput{Items: jf.ToMediaItems(items), Ranked: items}, nil
}
