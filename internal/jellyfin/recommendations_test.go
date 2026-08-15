package jellyfin

import "testing"

func TestRankRecommendationsUsesExplicitSignalsAndExcludesPlayed(t *testing.T) {
	history := []RecommendationCandidate{{Genres: []string{"science fiction"}, People: []string{"denis villeneuve"}, Favorite: true, Rating: 9}}
	profile := BuildPreferenceProfile(history)
	results := RankRecommendations([]RecommendationCandidate{
		{ID: "1", Name: "Dune", Genres: []string{"science fiction"}, People: []string{"denis villeneuve"}, CommunityRating: 8.5},
		{ID: "2", Name: "Already Watched", Genres: []string{"science fiction"}, Played: true},
		{ID: "3", Name: "Comedy", Genres: []string{"comedy"}, CommunityRating: 7},
	}, profile, "", 0, false, 10)
	if len(results) != 2 || results[0].Candidate.ID != "1" {
		t.Fatalf("unexpected results: %#v", results)
	}
	if len(results[0].Reasons) == 0 {
		t.Fatal("expected explainable recommendation reasons")
	}
}

func TestRankRecommendationsHonorsRuntimeAndLimit(t *testing.T) {
	results := RankRecommendations([]RecommendationCandidate{
		{ID: "1", Name: "Long", RuntimeMinutes: 140},
		{ID: "2", Name: "Short", RuntimeMinutes: 80},
		{ID: "3", Name: "Shorter", RuntimeMinutes: 70},
	}, PreferenceProfile{Genres: map[string]float64{}, People: map[string]float64{}}, "", 90, false, 1)
	if len(results) != 1 || results[0].Candidate.ID != "2" {
		t.Fatalf("unexpected constrained results: %#v", results)
	}
}

func TestSelectPlaylistItemsFillsBudgetAndDiversifiesGenres(t *testing.T) {
	ranked := []RankedRecommendation{
		{Candidate: RecommendationCandidate{ID: "1", RuntimeMinutes: 30, Genres: []string{"drama"}}},
		{Candidate: RecommendationCandidate{ID: "2", RuntimeMinutes: 30, Genres: []string{"drama"}}},
		{Candidate: RecommendationCandidate{ID: "3", RuntimeMinutes: 30, Genres: []string{"comedy"}}},
		{Candidate: RecommendationCandidate{ID: "4", RuntimeMinutes: 30, Genres: []string{"action"}}},
	}
	selected := SelectPlaylistItems(ranked, 90, 10)
	if len(selected) != 3 || selected[0].Candidate.ID != "1" || selected[2].Candidate.ID != "3" {
		t.Fatalf("unexpected playlist selection: %#v", selected)
	}
}
