package prompts

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	jf "github.com/jaredtrent/jellyfin-mcp/internal/jellyfin"
)

func RegisterPrompts(server *mcp.Server, _ jf.Client) {

	// --- find-and-play ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "find-and-play",
		Description: "Search for media by name and start playback on a connected client",
		Arguments: []*mcp.PromptArgument{
			{Name: "query", Description: "What to search for (title, artist, album, etc.)", Required: true},
			{Name: "type", Description: "Media type filter: Movie, Series, Episode, Audio, MusicAlbum"},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		query := req.Params.Arguments["query"]
		mediaType := req.Params.Arguments["type"]

		instruction := fmt.Sprintf(
			"Help me find and play \"%s\". Follow these steps:\n"+
				"1. Use jellyfin_search with query=\"%s\"",
			query, query)
		if mediaType != "" {
			instruction += fmt.Sprintf(" and type=\"%s\"", mediaType)
		}
		instruction += "\n2. Show me the results and let me pick one" +
			"\n3. Use jellyfin_sessions action=\"list\" to find an active session" +
			"\n4. If no sessions are active, tell me I need to open a Jellyfin client (web, mobile, or TV app) first" +
			"\n5. Use jellyfin_play with the chosen item and session"

		return &mcp.GetPromptResult{
			Description: "Find and play media",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- resume-watching ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "resume-watching",
		Description: "Pick up where you left off — in-progress movies, episodes, and audio with resume positions",
	}, func(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		instruction := "Show me what I can continue watching or listening to. Follow these steps:\n" +
			"1. Gather data in parallel:\n" +
			"   - jellyfin_sessions action=\"resume\" to get items with saved playback positions\n" +
			"   - jellyfin_recommendations type=\"next_up\" to find the next episode in series I'm following\n" +
			"2. Present a combined \"pick up where you left off\" list organized by type (movies, TV, audio), " +
			"showing the title, progress percentage or time remaining, and when I last watched each item"

		return &mcp.GetPromptResult{
			Description: "Resume in-progress media",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- whats-new ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "whats-new",
		Description: "Recently added media and next episodes to watch — a personalized viewing guide",
		Arguments: []*mcp.PromptArgument{
			{Name: "library", Description: "Library name to scope results (e.g. Movies, TV Shows)"},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		library := req.Params.Arguments["library"]

		instruction := "Show me what's new and what I should watch next. Follow these steps:\n"
		if library != "" {
			instruction += fmt.Sprintf(
				"1. Use jellyfin_libraries to find the library ID for \"%s\"\n"+
					"2. Gather data in parallel:\n"+
					"   - jellyfin_recommendations type=\"latest\" with that parent_id\n"+
					"   - jellyfin_recommendations type=\"next_up\" to see next episodes to watch\n"+
					"   - jellyfin_sessions action=\"resume\" to find in-progress items\n", library)
			instruction += "3. Present a combined viewing guide: new additions, next episodes, and items to resume — " +
				"organized by category with ratings and brief descriptions"
		} else {
			instruction += "1. Gather data in parallel:\n" +
				"   - jellyfin_recommendations type=\"latest\" to see recently added items\n" +
				"   - jellyfin_recommendations type=\"next_up\" to see next episodes to watch\n" +
				"   - jellyfin_sessions action=\"resume\" to find in-progress items\n"
			instruction += "2. Present a combined viewing guide: new additions, next episodes, and items to resume — " +
				"organized by category with ratings and brief descriptions"
		}

		return &mcp.GetPromptResult{
			Description: "Latest additions and next episodes",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- movie-night ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "movie-night",
		Description: "Curated movie suggestions from your library, filtered by genre and rating",
		Arguments: []*mcp.PromptArgument{
			{Name: "genre", Description: "Preferred genre: Action, Comedy, Drama, Horror, Sci-Fi, Thriller, etc."},
			{Name: "mood", Description: "Viewing mood: relaxing, exciting, thought-provoking, funny, scary"},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		genre := req.Params.Arguments["genre"]
		mood := req.Params.Arguments["mood"]

		instruction := "Help me pick a movie for tonight. Follow these steps:\n"

		if genre != "" {
			instruction += fmt.Sprintf(
				"1. Use jellyfin_browse with type=\"Movie\", genre=\"%s\", is_played=false, "+
					"sort_by=\"CommunityRating\", sort_order=\"Descending\", limit=10\n", genre)
		} else {
			instruction += "1. Use jellyfin_browse with type=\"Movie\", is_played=false, " +
				"sort_by=\"CommunityRating\", sort_order=\"Descending\", limit=10\n"
		}

		instruction += "2. Use jellyfin_recommendations type=\"personalized\" with query=" + fmt.Sprintf("\"%s\"", mood) + " for explainable library-local picks\n" +
			"3. Present the top 5 suggestions with ratings, year, runtime, and a brief overview"

		if mood != "" {
			instruction += fmt.Sprintf(
				"\n4. When ranking results, prioritize movies that match a \"%s\" mood based on their genre and overview — "+
					"for example, \"relaxing\" favors light dramas and comedies, \"exciting\" favors action and thrillers, "+
					"\"scary\" favors horror and suspense", mood)
		}

		return &mcp.GetPromptResult{
			Description: "Movie night suggestions",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- playlist-builder ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "playlist-builder",
		Description: "Generate a playlist from a natural-language mood, activity, or time budget",
		Arguments: []*mcp.PromptArgument{
			{Name: "name", Description: "Playlist name", Required: true},
			{Name: "request", Description: "Mood or activity, such as relaxed Sunday, workout, or family movie night", Required: true},
			{Name: "duration_minutes", Description: "Target duration in minutes"},
			{Name: "media_type", Description: "Audio or Video (default: Video)"},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		name := req.Params.Arguments["name"]
		request := req.Params.Arguments["request"]
		duration := req.Params.Arguments["duration_minutes"]
		mediaType := req.Params.Arguments["media_type"]
		if duration == "" {
			duration = "90"
		}
		if mediaType == "" {
			mediaType = "Video"
		}
		instruction := fmt.Sprintf("Build a Jellyfin playlist named %q for: %s. Follow these steps:\n", name, request) +
			fmt.Sprintf("1. Use jellyfin_playlists action=\"generate\" with name=%q, query=%q, media_type=%q, duration_minutes=%s\n", name, request, mediaType, duration) +
			"2. Show me the generated preview and explain the strongest matches and exclusions\n" +
			"3. Do not create the playlist until I explicitly confirm the preview\n" +
			"4. If I confirm, call jellyfin_playlists action=\"generate\" again with confirm=true"
		return &mcp.GetPromptResult{
			Description: "Generate a smart playlist",
			Messages:    []*mcp.PromptMessage{{Role: "user", Content: &mcp.TextContent{Text: instruction}}},
		}, nil
	})

	// --- music-listen ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "music-listen",
		Description: "Find and play music — search by artist, album, or song and optionally generate a smart mix",
		Arguments: []*mcp.PromptArgument{
			{Name: "query", Description: "Artist name, album title, or song name", Required: true},
			{Name: "type", Description: "What to search for: artist, album, song, genre, playlist"},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		query := req.Params.Arguments["query"]
		searchType := req.Params.Arguments["type"]

		// Map friendly type names to Jellyfin item types
		jellyfinType := ""
		switch searchType {
		case "artist":
			jellyfinType = "MusicArtist"
		case "album":
			jellyfinType = "MusicAlbum"
		case "song":
			jellyfinType = "Audio"
		case "playlist":
			jellyfinType = "Playlist"
		}

		instruction := fmt.Sprintf("Help me find and play music: \"%s\". Follow these steps:\n", query)

		if searchType == "genre" {
			instruction += fmt.Sprintf(
				"1. Use jellyfin_browse with type=\"Audio\", genre=\"%s\", sort_by=\"Random\", limit=20\n", query)
		} else {
			instruction += fmt.Sprintf("1. Use jellyfin_search with query=\"%s\"", query)
			if jellyfinType != "" {
				instruction += fmt.Sprintf(", type=\"%s\"", jellyfinType)
			} else {
				instruction += ", type=\"MusicAlbum\" — also try type=\"MusicArtist\" if no album matches"
			}
			instruction += "\n"
		}

		instruction += "2. Show me the results and let me pick one\n" +
			"3. If I pick an artist or album, offer to generate a smart mix using " +
			"jellyfin_music action=\"instant_mix\" with that item's ID for a playlist of similar tracks\n" +
			"4. Use jellyfin_sessions action=\"list\" to find an active session\n" +
			"5. Use jellyfin_play with the chosen item(s) and session"

		return &mcp.GetPromptResult{
			Description: "Find and play music",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- binge-watch ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "binge-watch",
		Description: "Check progress on a TV series and queue up the next episodes to watch",
		Arguments: []*mcp.PromptArgument{
			{Name: "query", Description: "TV series name", Required: true},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		query := req.Params.Arguments["query"]

		instruction := fmt.Sprintf("Help me set up a binge session for \"%s\". Follow these steps:\n", query) +
			fmt.Sprintf("1. Use jellyfin_search with query=\"%s\", type=\"Series\" to find the show\n", query) +
			"2. Use jellyfin_tv_shows action=\"seasons\" with the series_id to see all seasons and watched/unwatched counts\n" +
			"3. Use jellyfin_tv_shows action=\"next_up\" with the series_id to find where I left off\n" +
			"4. Show my progress: which seasons are complete, where I am now, and how many episodes remain\n" +
			"5. Offer to queue the next 3-5 unwatched episodes — use jellyfin_tv_shows action=\"episodes\" to get their IDs, " +
			"then jellyfin_sessions action=\"list\" and jellyfin_play with play_command=\"PlayNow\" and all episode IDs"

		return &mcp.GetPromptResult{
			Description: "TV binge session setup",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- fix-subtitles ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "fix-subtitles",
		Description: "Search for and download missing subtitles for a movie or episode",
		Arguments: []*mcp.PromptArgument{
			{Name: "query", Description: "Movie or episode name to find subtitles for", Required: true},
			{Name: "language", Description: "Subtitle language code: en, es, fr, de, ja, pt, zh, ko (default: en)"},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		query := req.Params.Arguments["query"]
		language := req.Params.Arguments["language"]
		if language == "" {
			language = "en"
		}

		instruction := fmt.Sprintf("Help me find subtitles for \"%s\". Follow these steps:\n", query) +
			fmt.Sprintf("1. Use jellyfin_search with query=\"%s\" to find the item (try type=\"Movie\", then type=\"Episode\" if no match)\n", query) +
			fmt.Sprintf("2. Use jellyfin_subtitles_lyrics action=\"search_subtitles\" with the item_id and language=\"%s\"\n", language) +
			"3. Show me the available subtitles with their source, format, and rating if available\n" +
			"4. Let me pick one, then use jellyfin_subtitles_lyrics action=\"download_subtitle\" with the subtitle_id"

		return &mcp.GetPromptResult{
			Description: "Find and download subtitles",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- who-is-watching ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "who-is-watching",
		Description: "Dashboard of all active playback sessions — who's watching what, on which device",
	}, func(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		instruction := "Show me what's happening on the Jellyfin server right now. Follow these steps:\n" +
			"1. Use jellyfin_sessions action=\"list\" to get all active sessions\n" +
			"2. Present a dashboard for each session showing: user name, device/client name, " +
			"what's playing (title, type, season/episode for TV), playback state (playing/paused), " +
			"and progress (current position / total duration)\n" +
			"3. If no sessions are active, say so and suggest checking jellyfin_devices action=\"list\" " +
			"to see recently connected devices"

		return &mcp.GetPromptResult{
			Description: "Active session dashboard",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- troubleshoot ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "troubleshoot",
		Description: "Diagnose a server issue — check server logs for errors, failed tasks, plugin health, and system status",
		Arguments: []*mcp.PromptArgument{
			{Name: "issue", Description: "Description of the problem you're experiencing"},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		issue := req.Params.Arguments["issue"]

		instruction := "Diagnose the Jellyfin server. Follow these steps:\n" +
			"1. Use jellyfin_system_info action=\"ping\" to check if the server is responsive\n" +
			"2. After ping succeeds, gather diagnostic data in parallel:\n" +
			"   - jellyfin_system_info action=\"info\" for server version, pending restart, and update status\n" +
			"   - jellyfin_tasks action=\"list\" to check for failed or stuck tasks\n" +
			"   - jellyfin_plugins action=\"list\" to check plugin health (status 'Restart' or 'Superseded' indicates issues)\n" +
			"   - jellyfin_system_info action=\"logs\" to find the most recent server log (named log_*.log, not FFmpeg logs)\n" +
			"3. Use jellyfin_system_info action=\"log_file\" with the most recent server log, limit=100 — focus on [WRN] and [ERR] entries\n"

		if issue != "" {
			instruction += fmt.Sprintf(
				"4. The user reported this issue: \"%s\" — use issue-specific checks:\n", issue)
			lissue := strings.ToLower(issue)
			switch {
			case strings.Contains(lissue, "playback") || strings.Contains(lissue, "buffer") || strings.Contains(lissue, "transcode") || strings.Contains(lissue, "stream"):
				instruction += "   - jellyfin_sessions action=\"list\" to check active playback sessions\n" +
					"   - jellyfin_analytics action=\"codec_report\" to identify transcoding-heavy codecs\n" +
					"   - Reference jellyfin://guides/transcoding for hardware transcoding setup\n"
			case strings.Contains(lissue, "connect") || strings.Contains(lissue, "remote") || strings.Contains(lissue, "network") || strings.Contains(lissue, "access"):
				instruction += "   - jellyfin_system_info action=\"info\" for server address and network config\n" +
					"   - Reference jellyfin://guides/remote-access for reverse proxy and port forwarding setup\n"
			case strings.Contains(lissue, "library") || strings.Contains(lissue, "scan") || strings.Contains(lissue, "missing") || strings.Contains(lissue, "metadata"):
				instruction += "   - jellyfin_tasks action=\"list\" to check library scan status\n" +
					"   - jellyfin_libraries to verify library configuration\n" +
					"   - Reference jellyfin://guides/library-setup for library organization best practices\n"
			case strings.Contains(lissue, "plugin") || strings.Contains(lissue, "extension"):
				instruction += "   - jellyfin_plugins action=\"list\" to check installed plugins and versions\n" +
					"   - Reference jellyfin://guides/plugins for plugin troubleshooting\n"
			default:
				instruction += "   - jellyfin_system_info action=\"storage\" to check disk space\n" +
					"   - jellyfin_server action=\"list_backups\" for backup context\n" +
					"   - Correlate error messages from logs with the reported symptom\n"
			}
			instruction += "5. Summarize: root cause (if identifiable), affected components, and suggested fix\n" +
				"6. Reference jellyfin://guides/troubleshooting for common solutions if applicable\n"
		} else {
			instruction += "4. Also gather:\n" +
				"   - jellyfin_system_info action=\"storage\" to check disk space\n" +
				"   - jellyfin_server action=\"list_backups\" for backup context\n" +
				"5. Summarize: any issues found, affected components, and suggested fixes\n" +
				"6. Reference jellyfin://guides/troubleshooting for common solutions if applicable\n"
		}

		return &mcp.GetPromptResult{
			Description: "Troubleshoot server issues",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- bulk-metadata-fix ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "bulk-metadata-fix",
		Description: "Find and fix metadata issues across a library — missing overviews, wrong years, missing genres, or re-identify items",
		Arguments: []*mcp.PromptArgument{
			{Name: "library", Description: "Library name to scan (e.g. Movies, TV Shows)", Required: true},
			{Name: "issue", Description: "Issue type: missing_overview, wrong_year, missing_genres, wrong_title, re_identify", Required: true},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		library := req.Params.Arguments["library"]
		issue := req.Params.Arguments["issue"]

		instruction := fmt.Sprintf("Help me fix metadata issues in the \"%s\" library. Issue type: %s\n\n", library, issue)
		instruction += fmt.Sprintf("1. Use jellyfin_libraries to find the library ID for \"%s\"\n", library)

		switch issue {
		case "missing_overview":
			instruction += "2. Use jellyfin_browse with the library parent_id, sort_by=\"SortName\" to list items (paginate with start_index to cover the full library)\n" +
				"3. Browse results include overview when present — items WITHOUT an 'overview' field in the results genuinely lack one. No need to call jellyfin_get_item to confirm.\n" +
				"4. For items with missing overviews, use jellyfin_metadata action=\"search\" to find correct metadata\n" +
				"5. Present the items and proposed fixes, then use jellyfin_metadata action=\"batch_update\" with confirm=true to apply\n"
		case "wrong_year":
			instruction += "2. Use jellyfin_browse with the library parent_id to list items\n" +
				"3. Identify items where the year looks wrong (compare with known release years)\n" +
				"4. For each wrong item, use jellyfin_metadata action=\"search\" with the correct year to find the right match\n" +
				"5. Present corrections and use jellyfin_metadata action=\"batch_update\" with confirm=true to fix years\n"
		case "missing_genres":
			instruction += "2. Use jellyfin_browse with the library parent_id to list items\n" +
				"3. Browse results do NOT include genres — you must use jellyfin_get_item on each item to check if genres are empty\n" +
				"4. For items missing genres, use jellyfin_metadata action=\"search\" to find correct metadata\n" +
				"5. Present fixes and use jellyfin_metadata action=\"batch_update\" with confirm=true to apply genres\n"
		case "wrong_title", "re_identify":
			instruction += "2. Use jellyfin_browse with the library parent_id to list items\n" +
				"3. For each misidentified item, use jellyfin_metadata action=\"search\" with search_query set to the correct title\n" +
				"4. Show the search results and let the user pick the correct match\n" +
				"5. Use jellyfin_metadata action=\"apply\" with the correct provider_name and provider_id\n"
		default:
			instruction += "2. Use jellyfin_browse with the library parent_id to scan items\n" +
				"3. Identify any metadata problems\n" +
				"4. Present findings and offer to fix using jellyfin_metadata\n"
		}

		return &mcp.GetPromptResult{
			Description: "Bulk metadata fix",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- subtitle-audit ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "subtitle-audit",
		Description: "Audit a library for missing subtitles and batch-download them",
		Arguments: []*mcp.PromptArgument{
			{Name: "library", Description: "Library name to audit (e.g. Movies, TV Shows)", Required: true},
			{Name: "language", Description: "Subtitle language code (default: en)"},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		library := req.Params.Arguments["library"]
		language := req.Params.Arguments["language"]
		if language == "" {
			language = "en"
		}

		instruction := fmt.Sprintf("Audit the \"%s\" library for missing subtitles (language: %s). Follow these steps:\n", library, language) +
			fmt.Sprintf("1. Use jellyfin_libraries to find the library ID for \"%s\"\n", library) +
			"2. Use jellyfin_browse with parent_id and has_subtitles=false to find items without subtitles\n" +
			"3. Report the count of items missing subtitles\n" +
			"4. Ask if I want to batch-download subtitles for these items\n" +
			fmt.Sprintf("5. If confirmed, use jellyfin_subtitles_lyrics action=\"batch_download_subtitles\" with the item_ids and language=\"%s\" and confirm=true\n", language) +
			"6. Present the results showing which items got subtitles and which failed"

		return &mcp.GetPromptResult{
			Description: "Subtitle audit and download",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- library-report ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "library-report",
		Description: "Comprehensive library analytics report: stats, codecs, unplayed items, recent additions, and duplicates",
		Arguments: []*mcp.PromptArgument{
			{Name: "library", Description: "Library name to report on (optional, all libraries if omitted)"},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		library := req.Params.Arguments["library"]

		instruction := "Generate a comprehensive library analytics report. Follow these steps:\n"
		if library != "" {
			instruction += fmt.Sprintf("1. Use jellyfin_libraries to find the library ID for \"%s\"\n", library)
			instruction += "2. Gather all analytics data in parallel using the parent_id:\n"
			instruction += "   - jellyfin_analytics action=\"library_stats\" for item counts by type\n"
			instruction += "   - jellyfin_analytics action=\"codec_report\" for codec/resolution distribution\n"
			instruction += "   - jellyfin_analytics action=\"never_played\" for unplayed items\n"
			instruction += "   - jellyfin_analytics action=\"recently_added\" for recent additions\n"
			instruction += "   - jellyfin_analytics action=\"duplicate_check\" for potential duplicates\n"
		} else {
			instruction += "1. Gather all analytics data in parallel:\n"
			instruction += "   - jellyfin_analytics action=\"library_stats\" for item counts by type\n"
			instruction += "   - jellyfin_analytics action=\"codec_report\" for codec/resolution distribution\n"
			instruction += "   - jellyfin_analytics action=\"never_played\" for unplayed items\n"
			instruction += "   - jellyfin_analytics action=\"recently_added\" for recent additions\n"
			instruction += "   - jellyfin_analytics action=\"duplicate_check\" for potential duplicates\n"
		}
		instruction += "\nPresent a formatted summary report with:\n" +
			"- Library size by content type\n" +
			"- Top codecs, resolutions, and containers\n" +
			"- Number of never-played items\n" +
			"- Recent additions summary\n" +
			"- Any duplicate items found\n" +
			"- Recommendations (e.g., codec issues that may cause transcoding, items to clean up)"

		return &mcp.GetPromptResult{
			Description: "Library analytics report",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- duplicate-finder ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "duplicate-finder",
		Description: "Find duplicate media items in a library and optionally remove inferior copies",
		Arguments: []*mcp.PromptArgument{
			{Name: "library", Description: "Library name to scan (e.g. Movies)", Required: true},
			{Name: "type", Description: "Item type to check: Movie, Series, Episode, Audio (default: Movie)"},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		library := req.Params.Arguments["library"]
		itemType := req.Params.Arguments["type"]
		if itemType == "" {
			itemType = "Movie"
		}

		instruction := fmt.Sprintf("Find duplicate %s items in the \"%s\" library. Follow these steps:\n", itemType, library) +
			fmt.Sprintf("1. Use jellyfin_libraries to find the library ID for \"%s\"\n", library) +
			fmt.Sprintf("2. Use jellyfin_analytics action=\"duplicate_check\" type=\"%s\" with the parent_id\n", itemType) +
			"3. For each duplicate group, use jellyfin_get_item for each copy to compare quality (resolution, bitrate, codec, file size)\n" +
			"4. Present a comparison table for each duplicate group showing:\n" +
			"   - Resolution, codec, bitrate, file size, file path\n" +
			"   - Which copy is recommended to keep (higher quality)\n" +
			"5. Ask if I want to delete any inferior copies\n" +
			"6. If confirmed, use jellyfin_library_manage action=\"delete_item\" with confirm=true for each item to remove"

		return &mcp.GetPromptResult{
			Description: "Find and remove duplicate media",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- watch-history ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "watch-history",
		Description: "View watch history for a user over a time period — what was played, when, and for how long",
		Arguments: []*mcp.PromptArgument{
			{Name: "user", Description: "Username to check history for (optional, defaults to current user)"},
			{Name: "days", Description: "Number of days to look back (default: 30)"},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		user := req.Params.Arguments["user"]
		days := req.Params.Arguments["days"]
		if days == "" {
			days = "30"
		}

		instruction := "Show watch history. Follow these steps:\n"
		if user != "" {
			instruction += fmt.Sprintf("1. Use jellyfin_users action=\"list\" to find the user ID for \"%s\"\n", user)
			instruction += "2. Use jellyfin_system_info action=\"playback_history\" with the user_id, limit=200\n"
		} else {
			instruction += "1. Use jellyfin_recommendations type=\"recently_played\" to get recently watched items sorted by last played date\n"
		}
		step := 2
		if user != "" {
			step = 3
		}
		instruction += fmt.Sprintf("%d. Filter to the last %s days based on the last_played date\n", step, days) +
			fmt.Sprintf("%d. Present a chronological watch history showing:\n", step+1) +
			"   - Date and time\n" +
			"   - What was played (title, type, series/episode info)\n" +
			"   - User who played it (if showing all users)\n" +
			fmt.Sprintf("%d. Summarize: total items watched, most-watched genres, average per day", step+2)

		return &mcp.GetPromptResult{
			Description: "Watch history report",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- codec-optimize ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "codec-optimize",
		Description: "Analyze media codecs and optimize transcoding settings to reduce server load",
		Arguments: []*mcp.PromptArgument{
			{Name: "library", Description: "Library name to analyze (optional, all libraries if omitted)"},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		library := req.Params.Arguments["library"]

		instruction := "Analyze media codecs and optimize transcoding settings. Follow these steps:\n"
		if library != "" {
			instruction += fmt.Sprintf("1. Use jellyfin_libraries to find the library ID for \"%s\"\n", library)
			instruction += "2. Gather data in parallel using the parent_id:\n"
		} else {
			instruction += "1. Gather data in parallel:\n"
		}
		instruction += "   - jellyfin_analytics action=\"codec_report\" for codec/resolution distribution\n" +
			"   - jellyfin_server action=\"get_config_section\" key=\"encoding\" for current transcoding settings\n" +
			"   - jellyfin_system_info action=\"info\" for server hardware info\n"
		step := 3
		if library != "" {
			step = 4
		}
		instruction += fmt.Sprintf("%d. Analyze the codec distribution and identify:\n", step) +
			"   - Codecs that require transcoding on most clients (e.g. HEVC on older devices)\n" +
			"   - Resolution distribution and bandwidth implications\n" +
			"   - Audio codecs that may need transcoding\n"
		instruction += fmt.Sprintf("%d. Compare with current hardware transcoding settings and suggest optimizations:\n", step+1) +
			"   - Recommend hardware acceleration method based on server hardware\n" +
			"   - Suggest enabling/disabling specific codec support\n" +
			"   - Reference jellyfin://guides/transcoding for setup instructions\n"
		instruction += fmt.Sprintf("%d. If changes are needed, offer to apply via jellyfin_server action=\"update_config_section\" key=\"encoding\"", step+2)

		return &mcp.GetPromptResult{
			Description: "Codec analysis and transcoding optimization",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- parental-controls ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "parental-controls",
		Description: "Set up kid-safe access — create restricted user accounts with content rating limits and library access controls",
		Arguments: []*mcp.PromptArgument{
			{Name: "username", Description: "Username for the child account", Required: true},
			{Name: "max_rating", Description: "Maximum content rating to allow: G, PG, PG-13, TV-Y, TV-G, TV-PG, TV-14 (default: PG)"},
		},
	}, func(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		username := req.Params.Arguments["username"]
		maxRating := req.Params.Arguments["max_rating"]
		if maxRating == "" {
			maxRating = "PG"
		}

		instruction := fmt.Sprintf("Set up parental controls for a child account \"%s\" (max rating: %s). Follow these steps:\n", username, maxRating) +
			"1. Use jellyfin_users action=\"list\" to check if the user already exists\n" +
			fmt.Sprintf("2. If not, use jellyfin_users action=\"create\" username=\"%s\" with a password\n", username) +
			"3. Use jellyfin_libraries to list available libraries and their IDs\n" +
			"4. Ask which libraries the child should have access to (suggest excluding adult-oriented libraries)\n" +
			"5. Use jellyfin_users action=\"update_policy\" with:\n" +
			"   - is_admin=false\n" +
			"   - enable_all_folders=false\n" +
			"   - enabled_folder_ids=[selected library IDs]\n" +
			"6. Reference jellyfin://guides/users-and-access for additional parental control options:\n" +
			"   - Content rating limits (set in Dashboard > Users > select user > Access)\n" +
			"   - Tag-based blocking for specific content\n" +
			"   - Access schedules to limit viewing times\n" +
			fmt.Sprintf("7. Summarize the configured restrictions for \"%s\"", username)

		return &mcp.GetPromptResult{
			Description: "Parental controls setup",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- server-setup ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "server-setup",
		Description: "Review and optimize server configuration — transcoding, networking, and general settings",
	}, func(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		instruction := "Review and optimize the Jellyfin server configuration. Follow these steps:\n" +
			"1. Gather server state in parallel:\n" +
			"   - jellyfin_system_info action=\"info\" for server version and OS\n" +
			"   - jellyfin_server action=\"get_config\" for full server configuration\n" +
			"   - jellyfin_server action=\"get_config_section\" key=\"encoding\" for transcoding settings\n" +
			"2. Reference jellyfin://guides/transcoding for hardware transcoding best practices\n" +
			"3. Analyze the current configuration and identify:\n" +
			"   - Transcoding: hardware acceleration method, enabled codecs, bitrate limits\n" +
			"   - Networking: base URL, remote access settings\n" +
			"   - General: metadata providers, subtitle settings, scheduled tasks\n" +
			"4. Present findings and suggest optimizations based on the server's OS and hardware\n" +
			"5. For any changes, use jellyfin_server action=\"update_config_section\" with the modified section"

		return &mcp.GetPromptResult{
			Description: "Server configuration review",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})

	// --- library-health ---
	server.AddPrompt(&mcp.Prompt{
		Name:        "library-health",
		Description: "Comprehensive server health check: status, storage, tasks, plugins, logs, and backups",
	}, func(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		instruction := "Run a comprehensive health check on the Jellyfin server. Follow these steps:\n" +
			"1. Use jellyfin_system_info action=\"health_check\" — this checks server status, storage, tasks, plugins, logs, and backups in one call\n" +
			"2. If the health_check returns warnings or errors, drill into specific areas:\n" +
			"   - For storage warnings: jellyfin_analytics action=\"size_report\" to find what's using space\n" +
			"   - For task failures: jellyfin_tasks action=\"get\" with the failed task_id for details\n" +
			"   - For log errors: jellyfin_system_info action=\"log_file\" with severity=\"warn+error\" for full context\n" +
			"   - For plugin issues: jellyfin_plugins action=\"list\" for version details\n" +
			"3. Summarize: overall status, any action items, and their urgency\n" +
			"4. If codec or transcoding issues are found, suggest reading jellyfin://guides/transcoding"

		return &mcp.GetPromptResult{
			Description: "Server health check",
			Messages: []*mcp.PromptMessage{{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			}},
		}, nil
	})
}
