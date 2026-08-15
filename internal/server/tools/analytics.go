package tools

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	jf "github.com/jaredtrent/jellyfin-mcp/internal/jellyfin"
)

func RegisterAnalyticsTools(server *mcp.Server, client jf.Client, enabled func(string, *mcp.ToolAnnotations) bool) {

	// --- jellyfin_analytics ---
	if enabled("jellyfin_analytics", AnnotReadOnly) {
		mcp.AddTool(server, &mcp.Tool{
			Name:  "jellyfin_analytics",
			Title: "Analytics",
			InputSchema: jf.WithEnums[jf.AnalyticsInput](map[string][]any{
				"action": {"library_stats", "library_size", "size_report", "codec_report", "never_played", "recently_added", "duplicate_check", "played_status"},
			}),
			Description: "Library analytics and reports. All actions are read-only. Use parent_id to scope to a specific library.",
			Annotations: AnnotReadOnly,
		}, func(ctx context.Context, req *mcp.CallToolRequest, args jf.AnalyticsInput) (*mcp.CallToolResult, *jf.AnalyticsOutput, error) {
			userID, err := client.GetUserID(ctx)
			if err != nil {
				return jf.ErrResult("Jellyfin error: %v", err), nil, nil
			}

			switch args.Action {
			case "library_stats":
				types := []string{"Movie", "Series", "Episode", "Audio", "MusicAlbum", "MusicArtist", "MusicVideo", "Book", "BoxSet"}
				if args.Type != "" {
					types = []string{args.Type}
				}
				stats := make([]map[string]any, 0, len(types))
				for _, t := range types {
					params := url.Values{
						"IncludeItemTypes": {t},
						"Recursive":        {"true"},
						"Limit":            {"0"},
					}
					if args.ParentID != "" {
						params.Set("ParentId", args.ParentID)
					}
					var result map[string]any
					endpoint := fmt.Sprintf("/Users/%s/Items", jf.SanitizeID(userID))
					if err := client.Get(ctx, endpoint, params, &result); err != nil {
						continue
					}
					count := jf.GetInt(result, "TotalRecordCount")
					if count > 0 {
						stats = append(stats, map[string]any{
							"type":  t,
							"count": count,
						})
					}
				}
				typeCounts := make([]jf.TypeCount, 0, len(stats))
				for _, s := range stats {
					tc := jf.TypeCount{Type: jf.GetString(s, "type")}
					if v, ok := s["count"].(int); ok {
						tc.Count = v
					}
					typeCounts = append(typeCounts, tc)
				}
				return jf.TextResult(fmt.Sprintf("Library statistics:\n\n%s", jf.FormatJSON(stats))), &jf.AnalyticsOutput{Stats: typeCounts}, nil

			case "codec_report":
				itemType := args.Type
				if itemType == "" {
					itemType = "Movie"
				}
				params := url.Values{
					"IncludeItemTypes": {itemType},
					"Recursive":        {"true"},
					"Fields":           {"MediaSources"},
				}
				if args.ParentID != "" {
					params.Set("ParentId", args.ParentID)
				}
				endpoint := fmt.Sprintf("/Users/%s/Items", jf.SanitizeID(userID))
				rawItems, _, err := jf.FetchAllPages(ctx, client, endpoint, params, 0)
				if err != nil {
					return jf.ErrResult("Jellyfin API error: %v", err), nil, nil
				}

				videoCodecs := make(map[string]int)
				audioCodecs := make(map[string]int)
				containers := make(map[string]int)
				resolutions := make(map[string]int)
				videoRanges := make(map[string]int)
				bitDepths := make(map[string]int)
				total := 0

				for _, raw := range rawItems {
					m := jf.ToMap(raw)
					sources := jf.ToSlice(m["MediaSources"])
					for _, src := range sources {
						sm := jf.ToMap(src)
						if sm == nil {
							continue
						}
						total++
						if c := jf.GetString(sm, "Container"); c != "" {
							containers[strings.ToLower(c)]++
						}
						for _, st := range jf.ToSlice(sm["MediaStreams"]) {
							stm := jf.ToMap(st)
							if stm == nil {
								continue
							}
							switch jf.GetString(stm, "Type") {
							case "Video":
								if codec := jf.GetString(stm, "Codec"); codec != "" {
									videoCodecs[strings.ToLower(codec)]++
								}
								if w := jf.GetInt(stm, "Width"); w > 0 {
									h := jf.GetInt(stm, "Height")
									var res string
									switch {
									case w >= 3840:
										res = "4K (2160p)"
									case w >= 1920:
										res = "1080p"
									case w >= 1280:
										res = "720p"
									case w >= 720:
										res = "480p"
									default:
										res = fmt.Sprintf("%dx%d", w, h)
									}
									resolutions[res]++
								}
								// HDR and bit depth
								if vr := jf.GetString(stm, "VideoRangeType"); vr != "" {
									videoRanges[vr]++
								} else if vr := jf.GetString(stm, "VideoRange"); vr != "" {
									videoRanges[vr]++
								}
								if bd := jf.GetInt(stm, "BitDepth"); bd > 0 {
									bitDepths[fmt.Sprintf("%d-bit", bd)]++
								}
							case "Audio":
								if codec := jf.GetString(stm, "Codec"); codec != "" {
									audioCodecs[strings.ToLower(codec)]++
								}
							}
						}
					}
				}

				report := map[string]any{
					"total_media_sources": total,
					"video_codecs":        videoCodecs,
					"audio_codecs":        audioCodecs,
					"containers":          containers,
					"resolutions":         resolutions,
					"video_ranges":        videoRanges,
					"bit_depths":          bitDepths,
				}
				return jf.TextResult(fmt.Sprintf("Codec report (%d media sources analyzed):\n\n%s", total, jf.FormatJSON(report))), &jf.AnalyticsOutput{CodecReport: &jf.CodecDistribution{
					TotalMediaSources: total,
					VideoCodecs:       videoCodecs,
					AudioCodecs:       audioCodecs,
					Containers:        containers,
					Resolutions:       resolutions,
					VideoRanges:       videoRanges,
					BitDepths:         bitDepths,
				}}, nil

			case "never_played":
				maxItems := jf.ClampInt(args.Limit, 500, jf.MaxLimitCap)
				params := url.Values{
					"IsPlayed":  {"false"},
					"Recursive": {"true"},
					"SortBy":    {"DateCreated"},
					"SortOrder": {"Descending"},
					"Fields":    {"Overview,ProductionYear,CommunityRating,DateCreated"},
				}
				if args.Type != "" {
					params.Set("IncludeItemTypes", args.Type)
				} else {
					params.Set("IncludeItemTypes", "Movie,Episode")
				}
				if args.ParentID != "" {
					params.Set("ParentId", args.ParentID)
				}
				endpoint := fmt.Sprintf("/Users/%s/Items", jf.SanitizeID(userID))
				rawItems, total, err := jf.FetchAllPages(ctx, client, endpoint, params, maxItems)
				if err != nil {
					return jf.ErrResult("Jellyfin API error: %v", err), nil, nil
				}
				items := jf.MapExtract(rawItems, jf.ExtractMediaItem)
				msg := fmt.Sprintf("Never played (%d total, showing %d):\n\n%s", total, len(items), jf.FormatJSON(items))
				if len(items) < total {
					msg += fmt.Sprintf("\n\nMore results available. Increase limit (currently %d) to see more.", maxItems)
				}
				return jf.TextResult(msg), &jf.AnalyticsOutput{Items: jf.ToMediaItems(items), TotalCount: total}, nil

			case "recently_added":
				days := args.Days
				if days <= 0 {
					days = 30
				}
				maxItems := jf.ClampInt(args.Limit, 500, jf.MaxLimitCap)
				params := url.Values{
					"Recursive": {"true"},
					"SortBy":    {"DateCreated"},
					"SortOrder": {"Descending"},
					"Fields":    {"Overview,ProductionYear,CommunityRating,DateCreated"},
				}
				if args.Type != "" {
					params.Set("IncludeItemTypes", args.Type)
				} else {
					params.Set("IncludeItemTypes", "Movie,Series")
				}
				if args.ParentID != "" {
					params.Set("ParentId", args.ParentID)
				}
				endpoint := fmt.Sprintf("/Users/%s/Items", jf.SanitizeID(userID))
				rawItems, total, err := jf.FetchAllPages(ctx, client, endpoint, params, maxItems)
				if err != nil {
					return jf.ErrResult("Jellyfin API error: %v", err), nil, nil
				}
				cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
				items := filterRecentlyAdded(rawItems, cutoff)
				total = len(items)
				msg := fmt.Sprintf("Recently added in last %d days (%d total, showing %d):\n\n%s", days, total, len(items), jf.FormatJSON(items))
				if len(items) >= maxItems {
					msg += fmt.Sprintf("\n\nMore results available. Increase limit (currently %d) to see more.", maxItems)
				}
				return jf.TextResult(msg), &jf.AnalyticsOutput{Items: jf.ToMediaItems(items), TotalCount: total}, nil

			case "duplicate_check":
				itemType := args.Type
				if itemType == "" {
					itemType = "Movie"
				}
				params := url.Values{
					"IncludeItemTypes": {itemType},
					"Recursive":        {"true"},
					"Fields":           {"ProductionYear,Path"},
				}
				if args.ParentID != "" {
					params.Set("ParentId", args.ParentID)
				}
				endpoint := fmt.Sprintf("/Users/%s/Items", jf.SanitizeID(userID))
				rawItems, _, err := jf.FetchAllPages(ctx, client, endpoint, params, 0)
				if err != nil {
					return jf.ErrResult("Jellyfin API error: %v", err), nil, nil
				}

				// Group by normalized name + year
				type itemInfo struct {
					ID   string
					Name string
					Year int
					Path string
				}
				groups := make(map[string][]itemInfo)
				for _, raw := range rawItems {
					m := jf.ToMap(raw)
					name := jf.GetString(m, "Name")
					year := jf.GetInt(m, "ProductionYear")
					key := strings.ToLower(strings.TrimSpace(name))
					if year > 0 {
						key = fmt.Sprintf("%s (%d)", key, year)
					}
					groups[key] = append(groups[key], itemInfo{
						ID:   jf.GetString(m, "Id"),
						Name: name,
						Year: year,
						Path: jf.GetString(m, "Path"),
					})
				}

				// Filter to groups with 2+ items
				duplicates := make([]map[string]any, 0)
				for _, items := range groups {
					if len(items) < 2 {
						continue
					}
					copies := make([]map[string]any, 0, len(items))
					for _, it := range items {
						copy := map[string]any{
							"id":   it.ID,
							"name": it.Name,
						}
						if it.Year > 0 {
							copy["year"] = it.Year
						}
						if it.Path != "" {
							copy["path"] = it.Path
						}
						copies = append(copies, copy)
					}
					duplicates = append(duplicates, map[string]any{
						"name":   items[0].Name,
						"year":   items[0].Year,
						"copies": copies,
						"count":  len(items),
					})
				}
				dupGroups := make([]jf.DuplicateGroup, 0, len(duplicates))
				for _, d := range duplicates {
					group := jf.DuplicateGroup{
						Name: jf.GetString(d, "name"),
					}
					if v, ok := d["year"].(int); ok {
						group.Year = v
					}
					if v, ok := d["count"].(int); ok {
						group.Count = v
					}
					if copies, ok := d["copies"].([]map[string]any); ok {
						for _, c := range copies {
							dc := jf.DuplicateCopy{
								ID:   jf.GetString(c, "id"),
								Name: jf.GetString(c, "name"),
								Path: jf.GetString(c, "path"),
							}
							if v, ok := c["year"].(int); ok {
								dc.Year = v
							}
							group.Copies = append(group.Copies, dc)
						}
					}
					dupGroups = append(dupGroups, group)
				}
				return jf.TextResult(fmt.Sprintf("Potential duplicates (%d groups found):\n\n%s", len(duplicates), jf.FormatJSON(duplicates))), &jf.AnalyticsOutput{Duplicates: dupGroups}, nil

			case "library_size":
				types := []string{"Movie", "Episode", "Audio", "MusicVideo"}
				if args.Type != "" {
					types = []string{args.Type}
				}
				var totalBytes int64
				typeSizes := make([]map[string]any, 0)
				totalSteps := float64(len(types))
				for i, t := range types {
					jf.ReportProgress(ctx, req, float64(i), totalSteps, fmt.Sprintf("Calculating %s sizes...", t))
					// Fetch items in pages to aggregate sizes
					pageSize := 500
					startIndex := 0
					var typeBytes int64
					typeCount := 0
					for {
						params := url.Values{
							"IncludeItemTypes": {t},
							"Recursive":        {"true"},
							"Limit":            {fmt.Sprintf("%d", pageSize)},
							"StartIndex":       {fmt.Sprintf("%d", startIndex)},
							"Fields":           {"MediaSources"},
						}
						if args.ParentID != "" {
							params.Set("ParentId", args.ParentID)
						}
						endpoint := fmt.Sprintf("/Users/%s/Items", jf.SanitizeID(userID))
						var result map[string]any
						if err := client.Get(ctx, endpoint, params, &result); err != nil {
							break
						}
						rawItems := jf.ToSlice(result["Items"])
						if len(rawItems) == 0 {
							break
						}
						for _, raw := range rawItems {
							m := jf.ToMap(raw)
							for _, src := range jf.ToSlice(m["MediaSources"]) {
								sm := jf.ToMap(src)
								if sz := jf.GetInt64(sm, "Size"); sz > 0 {
									typeBytes += sz
									typeCount++
								}
							}
						}
						totalRecords := jf.GetInt(result, "TotalRecordCount")
						startIndex += len(rawItems)
						if startIndex >= totalRecords {
							break
						}
					}
					if typeBytes > 0 {
						totalBytes += typeBytes
						typeSizes = append(typeSizes, map[string]any{
							"type":    t,
							"files":   typeCount,
							"size_gb": fmt.Sprintf("%.2f", float64(typeBytes)/float64(jf.BytesPerGB)),
							"size_mb": typeBytes / jf.BytesPerMB,
						})
					}
				}
				jf.ReportProgress(ctx, req, totalSteps, totalSteps, "Size calculation complete")
				summary := map[string]any{
					"total_size_gb": fmt.Sprintf("%.2f", float64(totalBytes)/float64(jf.BytesPerGB)),
					"total_size_mb": totalBytes / jf.BytesPerMB,
					"by_type":       typeSizes,
				}
				outSizes := make([]jf.TypeSize, 0, len(typeSizes))
				for _, ts := range typeSizes {
					outSizes = append(outSizes, jf.TypeSize{
						Type:   jf.GetString(ts, "type"),
						Files:  func() int { v, _ := ts["files"].(int); return v }(),
						SizeGB: jf.GetString(ts, "size_gb"),
						SizeMB: func() int64 { v, _ := ts["size_mb"].(int64); return v }(),
					})
				}
				return jf.TextResult(fmt.Sprintf("Library size:\n\n%s", jf.FormatJSON(summary))), &jf.AnalyticsOutput{
					TotalSizeGB: fmt.Sprintf("%.2f", float64(totalBytes)/float64(jf.BytesPerGB)),
					TotalSizeMB: totalBytes / jf.BytesPerMB,
					ByType:      outSizes,
				}, nil

			case "played_status":
				// 1. Validate item_id
				if args.ItemID == "" {
					return jf.ErrResult("item_id is required for played_status"), nil, nil
				}

				// 2. Fetch all users
				var users []map[string]any
				if err := client.Get(ctx, "/Users", nil, &users); err != nil {
					return jf.ErrResult("Jellyfin API error: %v", err), nil, nil
				}

				// 3. Fetch item metadata using the authenticated user
				var item map[string]any
				itemEndpoint := fmt.Sprintf("/Users/%s/Items/%s", jf.SanitizeID(userID), jf.SanitizeID(args.ItemID))
				if err := client.Get(ctx, itemEndpoint, nil, &item); err != nil {
					return jf.ErrResult("Jellyfin API error: %v", err), nil, nil
				}
				itemName := jf.GetString(item, "Name")
				itemType := jf.GetString(item, "Type")

				// 4. If Series, get total episode count via Limit=0 query
				totalEpisodes := 0
				if itemType == "Series" {
					epParams := url.Values{
						"ParentId":         {jf.SanitizeID(args.ItemID)},
						"IncludeItemTypes": {"Episode"},
						"Recursive":        {"true"},
						"Limit":            {"0"},
					}
					var epResult map[string]any
					epEndpoint := fmt.Sprintf("/Users/%s/Items", jf.SanitizeID(userID))
					if err := client.Get(ctx, epEndpoint, epParams, &epResult); err == nil {
						totalEpisodes = jf.GetInt(epResult, "TotalRecordCount")
					}
				}

				// 5. Build item summary
				itemSummary := map[string]any{
					"id":   args.ItemID,
					"name": itemName,
					"type": itemType,
				}
				if totalEpisodes > 0 {
					itemSummary["total_episodes"] = totalEpisodes
				}

				// 6. Query each non-disabled user's watch data
				userResults := make([]map[string]any, 0, len(users))
				for _, u := range users {
					if policy := jf.ToMap(u["Policy"]); policy != nil && jf.GetBool(policy, "IsDisabled") {
						continue
					}
					uid := jf.GetString(u, "Id")
					name := jf.GetString(u, "Name")

					var userItem map[string]any
					userEndpoint := fmt.Sprintf("/Users/%s/Items/%s", jf.SanitizeID(uid), jf.SanitizeID(args.ItemID))
					if err := client.Get(ctx, userEndpoint, url.Values{"enableUserData": {"true"}}, &userItem); err != nil {
						continue // skip users we can't query (permissions)
					}

					entry := map[string]any{"name": name}
					if ud := jf.ToMap(userItem["UserData"]); ud != nil {
						entry["played"] = jf.GetBool(ud, "Played")
						if pc := jf.GetInt(ud, "PlayCount"); pc > 0 {
							entry["play_count"] = pc
						}
						if lp := jf.GetString(ud, "LastPlayedDate"); lp != "" {
							entry["last_played"] = jf.Truncate(lp, jf.DateOnlyLen)
						}
						if itemType == "Series" && totalEpisodes > 0 {
							unplayed := jf.GetInt(ud, "UnplayedItemCount")
							entry["episodes_played"] = totalEpisodes - unplayed
							entry["total_episodes"] = totalEpisodes
						}
					}
					userResults = append(userResults, entry)
				}

				result := map[string]any{
					"item":  itemSummary,
					"users": userResults,
				}
				outUsers := make([]jf.PlayedStatusUser, 0, len(userResults))
				for _, ur := range userResults {
					pu := jf.PlayedStatusUser{Name: jf.GetString(ur, "name")}
					if v, ok := ur["played"].(bool); ok {
						pu.Played = v
					}
					if v, ok := ur["play_count"].(int); ok {
						pu.PlayCount = v
					}
					pu.LastPlayed = jf.GetString(ur, "last_played")
					if v, ok := ur["episodes_played"].(int); ok {
						pu.EpisodesPlayed = v
					}
					if v, ok := ur["total_episodes"].(int); ok {
						pu.TotalEpisodes = v
					}
					outUsers = append(outUsers, pu)
				}
				return jf.TextResult(fmt.Sprintf("Played status for %q (%d users):\n\n%s",
						itemName, len(userResults), jf.FormatJSON(result))), &jf.AnalyticsOutput{
						ItemSummary: &jf.PlayedStatusItem{
							ID:            args.ItemID,
							Name:          itemName,
							Type:          itemType,
							TotalEpisodes: totalEpisodes,
						},
						Users: outUsers,
					}, nil

			case "size_report":
				itemType := args.Type
				if itemType == "" {
					itemType = "Series"
				}
				limit := jf.ClampInt(args.Limit, 25, jf.MaxLimitCap)

				// For Series, fetch all Episodes and group by SeriesId
				fetchType := itemType
				if itemType == "Series" {
					fetchType = "Episode"
				}

				type sizeEntry struct {
					Name  string
					ID    string
					Type  string
					Bytes int64
					Files int
				}

				entries := make(map[string]*sizeEntry)
				pageSize := 500
				startIndex := 0
				pageNum := 0
				for {
					pageNum++
					jf.ReportProgress(ctx, req, float64(pageNum-1), 0, fmt.Sprintf("Scanning %s page %d...", fetchType, pageNum))
					params := url.Values{
						"IncludeItemTypes": {fetchType},
						"Recursive":        {"true"},
						"Limit":            {fmt.Sprintf("%d", pageSize)},
						"StartIndex":       {fmt.Sprintf("%d", startIndex)},
						"Fields":           {"MediaSources"},
					}
					if args.ParentID != "" {
						params.Set("ParentId", args.ParentID)
					}
					endpoint := fmt.Sprintf("/Users/%s/Items", jf.SanitizeID(userID))
					var result map[string]any
					if err := client.Get(ctx, endpoint, params, &result); err != nil {
						return jf.ErrResult("Jellyfin API error: %v", err), nil, nil
					}
					rawItems := jf.ToSlice(result["Items"])
					if len(rawItems) == 0 {
						break
					}
					for _, raw := range rawItems {
						m := jf.ToMap(raw)
						var entryID, entryName, entryType string
						if itemType == "Series" {
							entryID = jf.GetString(m, "SeriesId")
							entryName = jf.GetString(m, "SeriesName")
							entryType = "Series"
							if entryID == "" {
								continue
							}
						} else {
							entryID = jf.GetString(m, "Id")
							entryName = jf.GetString(m, "Name")
							entryType = jf.GetString(m, "Type")
						}
						for _, src := range jf.ToSlice(m["MediaSources"]) {
							sm := jf.ToMap(src)
							if sz := jf.GetInt64(sm, "Size"); sz > 0 {
								e, ok := entries[entryID]
								if !ok {
									e = &sizeEntry{Name: entryName, ID: entryID, Type: entryType}
									entries[entryID] = e
								}
								e.Bytes += sz
								e.Files++
							}
						}
					}
					totalRecords := jf.GetInt(result, "TotalRecordCount")
					startIndex += len(rawItems)
					if startIndex >= totalRecords {
						break
					}
				}

				// Sort descending by size
				sorted := make([]*sizeEntry, 0, len(entries))
				for _, e := range entries {
					sorted = append(sorted, e)
				}
				sort.Slice(sorted, func(i, j int) bool {
					return sorted[i].Bytes > sorted[j].Bytes
				})

				// Take top N
				if len(sorted) > limit {
					sorted = sorted[:limit]
				}

				results := make([]map[string]any, 0, len(sorted))
				for _, e := range sorted {
					results = append(results, map[string]any{
						"name":    e.Name,
						"id":      e.ID,
						"type":    e.Type,
						"files":   e.Files,
						"size_gb": fmt.Sprintf("%.2f", float64(e.Bytes)/float64(jf.BytesPerGB)),
						"size_mb": e.Bytes / jf.BytesPerMB,
					})
				}

				label := itemType
				if itemType == "Series" {
					label = "Series (by episode sizes)"
				}
				sizeEntries := make([]jf.SizeEntry, 0, len(results))
				for _, r := range results {
					sizeEntries = append(sizeEntries, jf.SizeEntry{
						Name:   jf.GetString(r, "name"),
						ID:     jf.GetString(r, "id"),
						Type:   jf.GetString(r, "type"),
						Files:  func() int { v, _ := r["files"].(int); return v }(),
						SizeGB: jf.GetString(r, "size_gb"),
						SizeMB: func() int64 { v, _ := r["size_mb"].(int64); return v }(),
					})
				}
				return jf.TextResult(fmt.Sprintf("Size report — top %d %s:\n\n%s", len(results), label, jf.FormatJSON(results))), &jf.AnalyticsOutput{SizeReport: sizeEntries}, nil

			default:
				return jf.ErrResult("Invalid action '%s'. Valid actions: library_stats, codec_report, never_played, recently_added, duplicate_check, library_size, size_report, played_status", args.Action), nil, nil
			}
		})
	}
}

// filterRecentlyAdded uses DateCreated, not PremiereDate, because this report
// is about when Jellyfin added an item to the library. Items with missing or
// invalid DateCreated values are excluded rather than being treated as recent.
func filterRecentlyAdded(rawItems []any, cutoff time.Time) []map[string]any {
	items := make([]map[string]any, 0, len(rawItems))
	for _, raw := range rawItems {
		m := jf.ToMap(raw)
		created := jf.GetString(m, "DateCreated")
		if created == "" {
			continue
		}
		createdAt, err := time.Parse(time.RFC3339Nano, created)
		if err != nil || createdAt.Before(cutoff) {
			continue
		}
		item := jf.ExtractMediaItem(m)
		item["date_added"] = createdAt.UTC().Format(jf.DateOnlyFormat)
		items = append(items, item)
	}
	return items
}
