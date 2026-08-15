package jellyfin

import (
	"fmt"
	"sort"
	"strings"
)

// RecommendationCandidate is the small, provider-neutral view used by the
// local recommendation and playlist engines.
type RecommendationCandidate struct {
	ID              string
	Name            string
	Type            string
	Genres          []string
	People          []string
	Played          bool
	Favorite        bool
	Rating          float64
	CommunityRating float64
	RuntimeMinutes  int
	Raw             map[string]any
}

// SelectPlaylistItems greedily fills a duration budget while avoiding a run of
// items with the same leading genre. It preserves the recommendation order.
func SelectPlaylistItems(ranked []RankedRecommendation, durationMinutes, limit int) []RankedRecommendation {
	if durationMinutes <= 0 {
		durationMinutes = 90
	}
	if limit <= 0 {
		limit = 25
	}
	selected := make([]RankedRecommendation, 0, limit)
	usedMinutes := 0
	genreCounts := make(map[string]int)
	for _, item := range ranked {
		if len(selected) >= limit {
			break
		}
		minutes := item.Candidate.RuntimeMinutes
		if minutes <= 0 {
			minutes = 5
		}
		if minutes > durationMinutes || (usedMinutes > 0 && usedMinutes+minutes > durationMinutes) {
			continue
		}
		primaryGenre := ""
		if len(item.Candidate.Genres) > 0 {
			primaryGenre = item.Candidate.Genres[0]
		}
		if primaryGenre != "" && genreCounts[primaryGenre] >= 3 {
			continue
		}
		selected = append(selected, item)
		usedMinutes += minutes
		if primaryGenre != "" {
			genreCounts[primaryGenre]++
		}
	}
	return selected
}

type PreferenceProfile struct {
	Genres map[string]float64
	People map[string]float64
}

type RankedRecommendation struct {
	Candidate RecommendationCandidate `json:"-"`
	Score     float64                 `json:"score"`
	Reasons   []string                `json:"reasons"`
}

// ParseRecommendationCandidate converts a Jellyfin item response into the
// fields needed for local ranking.
func ParseRecommendationCandidate(raw map[string]any) RecommendationCandidate {
	userData := ToMap(raw["UserData"])
	people := make([]string, 0)
	for _, person := range ToSlice(raw["People"]) {
		if p := ToMap(person); p != nil {
			if name := GetString(p, "Name"); name != "" {
				people = append(people, strings.ToLower(name))
			}
		}
	}
	return RecommendationCandidate{
		ID:              GetString(raw, "Id"),
		Name:            GetString(raw, "Name"),
		Type:            GetString(raw, "Type"),
		Genres:          lowerStrings(ToStringSlice(raw["Genres"])),
		People:          people,
		Played:          GetBool(userData, "Played"),
		Favorite:        GetBool(userData, "IsFavorite"),
		Rating:          GetFloat(userData, "Rating"),
		CommunityRating: GetFloat(raw, "CommunityRating"),
		RuntimeMinutes:  int(GetInt64(raw, "RunTimeTicks") / TicksPerMinute),
		Raw:             raw,
	}
}

func lowerStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			out = append(out, strings.ToLower(value))
		}
	}
	return out
}

// BuildPreferenceProfile learns only from strong positive signals. Watching a
// title is intentionally not treated as liking it; favorites and ratings are
// stronger evidence than passive viewing history.
func BuildPreferenceProfile(history []RecommendationCandidate) PreferenceProfile {
	profile := PreferenceProfile{Genres: make(map[string]float64), People: make(map[string]float64)}
	for _, item := range history {
		weight := 0.0
		if item.Played {
			weight += 0.25 // weak implicit signal; explicit ratings remain stronger
		}
		if item.Favorite {
			weight += 3
		}
		if item.Rating >= 7 {
			weight += 2 + (item.Rating-7)/3
		}
		if item.Rating > 0 && item.Rating < 5 {
			weight -= 2
		}
		if weight == 0 {
			continue
		}
		for _, genre := range item.Genres {
			profile.Genres[genre] += weight
		}
		for _, person := range item.People {
			profile.People[person] += weight
		}
	}
	return profile
}

// RankRecommendations ranks candidates using the learned profile and request
// constraints. The result is deterministic, bounded, and explainable.
func RankRecommendations(candidates []RecommendationCandidate, profile PreferenceProfile, query string, maxRuntime int, includePlayed bool, limit int) []RankedRecommendation {
	if limit <= 0 {
		limit = 25
	}
	queryTerms := strings.Fields(strings.ToLower(query))
	ranked := make([]RankedRecommendation, 0, len(candidates))
	for _, item := range candidates {
		if item.ID == "" || item.Name == "" || (!includePlayed && item.Played) {
			continue
		}
		if maxRuntime > 0 && item.RuntimeMinutes > maxRuntime {
			continue
		}
		score := item.CommunityRating * 1.5
		reasons := make([]string, 0, 4)
		if item.Favorite {
			score += 20
			reasons = append(reasons, "you marked it as a favorite")
		}
		if item.Rating >= 7 {
			score += 15 + item.Rating
			reasons = append(reasons, fmt.Sprintf("your rating is %.1f/10", item.Rating))
		} else if item.Rating > 0 && item.Rating < 5 {
			score -= 20
		}
		for _, genre := range item.Genres {
			if value := profile.Genres[genre]; value > 0 {
				score += value * 4
				if len(reasons) < 3 {
					reasons = append(reasons, "matches genres you rate highly")
				}
			}
		}
		for _, person := range item.People {
			if value := profile.People[person]; value > 0 {
				score += value * 2
				if len(reasons) < 3 {
					reasons = append(reasons, "matches creators you enjoy")
				}
			}
		}
		for _, term := range queryTerms {
			if strings.Contains(strings.ToLower(item.Name), term) {
				score += 10
				reasons = append(reasons, "matches your request")
				break
			}
		}
		if item.CommunityRating >= 8 {
			reasons = append(reasons, "well-rated in the library")
		}
		if len(reasons) == 0 {
			reasons = append(reasons, "unwatched library item matching your profile")
		}
		ranked = append(ranked, RankedRecommendation{Candidate: item, Score: score, Reasons: reasons})
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return ranked[i].Candidate.Name < ranked[j].Candidate.Name
		}
		return ranked[i].Score > ranked[j].Score
	})
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	return ranked
}
