package tools

import (
	"context"
	"fmt"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	jf "github.com/jaredtrent/jellyfin-mcp/internal/jellyfin"
)

func RegisterUserTools(server *mcp.Server, client jf.Client, enabled func(string, *mcp.ToolAnnotations) bool) {

	// --- jellyfin_user_data ---
	if enabled("jellyfin_user_data", AnnotWriteOp) {
		mcp.AddTool(server, &mcp.Tool{
			Name:  "jellyfin_user_data",
			Title: "User Interactions",
			InputSchema: jf.WithEnums[jf.UserDataInput](map[string][]any{
				"action": {"favorite", "unfavorite", "like", "dislike", "clear_rating", "mark_played", "mark_unplayed", "rate", "get_user_data", "set_user_data"},
			}),
			Description: "Manage user interactions with media items: toggle favorites, set ratings (like/dislike), and mark items as played or unplayed. " +
				"Use 'favorite' or 'unfavorite' to manage the favorites list. Use 'like' or 'dislike' to rate items, or 'clear_rating' to remove a rating. " +
				"Use 'mark_played' to set the played flag on a movie or episode, or 'mark_unplayed' to clear it. " +
				"Use 'rate' with a numeric rating (0-10) for precise scoring. " +
				"Use 'get_user_data' to read play count, position, rating, and played status. " +
				"Use 'set_user_data' to write play state fields (requires confirm=true). " +
				"The item_id must come from a previous search, browse, or recommendations result.",
			Annotations: AnnotWriteOp,
		}, func(ctx context.Context, req *mcp.CallToolRequest, args jf.UserDataInput) (*mcp.CallToolResult, any, error) {
			userID, err := client.GetUserID(ctx)
			if err != nil {
				return jf.ErrResult("Jellyfin error: %v", err), nil, nil
			}
			itemID := jf.SanitizeID(args.ItemID)

			type simpleAction struct {
				pathFmt    string
				isDelete   bool
				params     url.Values
				successMsg string
				errorVerb  string
			}
			simpleActions := map[string]simpleAction{
				"favorite":      {"/Users/%s/FavoriteItems/%s", false, nil, "Item added to favorites.", "add favorite"},
				"unfavorite":    {"/Users/%s/FavoriteItems/%s", true, nil, "Item removed from favorites.", "remove favorite"},
				"like":          {"/Users/%s/Items/%s/Rating", false, url.Values{"likes": {"true"}}, "Item rated: liked.", "set rating"},
				"dislike":       {"/Users/%s/Items/%s/Rating", false, url.Values{"likes": {"false"}}, "Item rated: disliked.", "set rating"},
				"clear_rating":  {"/Users/%s/Items/%s/Rating", true, nil, "Rating cleared.", "clear rating"},
				"mark_played":   {"/Users/%s/PlayedItems/%s", false, nil, "Item marked as played.", "mark as played"},
				"mark_unplayed": {"/Users/%s/PlayedItems/%s", true, nil, "Item marked as unplayed.", "mark as unplayed"},
			}
			if act, ok := simpleActions[args.Action]; ok {
				endpoint := fmt.Sprintf(act.pathFmt, jf.SanitizeID(userID), itemID)
				var err error
				if act.isDelete {
					err = client.Del(ctx, endpoint, act.params)
				} else {
					err = client.PostNoContent(ctx, endpoint, act.params, nil)
				}
				if err != nil {
					return jf.ErrResult("Failed to %s: %v", act.errorVerb, err), nil, nil
				}
				return jf.TextResult(act.successMsg), nil, nil
			}

			switch args.Action {
			case "rate":
				if args.Rating == nil {
					return jf.ErrResult("rating is required (0.0-10.0)."), nil, nil
				}
				body := map[string]any{"Rating": *args.Rating}
				endpoint := fmt.Sprintf("/Users/%s/Items/%s/UserData", jf.SanitizeID(userID), itemID)
				if err := client.PostNoContent(ctx, endpoint, nil, body); err != nil {
					return jf.ErrResult("Failed to set rating: %v", err), nil, nil
				}
				return jf.TextResult(fmt.Sprintf("Rating set to %.1f.", *args.Rating)), nil, nil

			case "get_user_data":
				endpoint := fmt.Sprintf("/Users/%s/Items/%s/UserData", jf.SanitizeID(userID), itemID)
				var data map[string]any
				if err := client.Get(ctx, endpoint, nil, &data); err != nil {
					return jf.ErrResult("Jellyfin API error: %v", err), nil, nil
				}
				result := map[string]any{
					"played":     jf.GetBool(data, "Played"),
					"play_count": jf.GetInt(data, "PlayCount"),
					"favorite":   jf.GetBool(data, "IsFavorite"),
				}
				if pos := jf.GetInt64(data, "PlaybackPositionTicks"); pos > 0 {
					result["position_ticks"] = pos
					result["position_seconds"] = pos / jf.TicksPerSecond
				}
				if rating := jf.GetFloat(data, "Rating"); rating > 0 {
					result["rating"] = rating
				}
				if lp := jf.GetString(data, "LastPlayedDate"); lp != "" {
					result["last_played"] = jf.Truncate(lp, jf.DateOnlyLen)
				}
				if pct := jf.GetFloat(data, "PlayedPercentage"); pct > 0 {
					result["played_percentage"] = pct
				}
				return jf.TextResult(fmt.Sprintf("User data:\n\n%s", jf.FormatJSON(result))), nil, nil

			case "set_user_data":
				if result := jf.ConfirmationGate(ctx, req, args.Confirm, "This will overwrite user data fields for this item."); result != nil {
					return result, nil, nil
				}
				body := make(map[string]any)
				if args.Played != nil {
					body["Played"] = *args.Played
				}
				if args.PlayCount != nil {
					body["PlayCount"] = *args.PlayCount
				}
				if args.PositionTicks != nil {
					body["PlaybackPositionTicks"] = *args.PositionTicks
				}
				if args.Rating != nil {
					body["Rating"] = *args.Rating
				}
				if len(body) == 0 {
					return jf.ErrResult("Provide at least one field to update: played, play_count, position_ticks, rating."), nil, nil
				}
				endpoint := fmt.Sprintf("/Users/%s/Items/%s/UserData", jf.SanitizeID(userID), itemID)
				if err := client.PostNoContent(ctx, endpoint, nil, body); err != nil {
					return jf.ErrResult("Failed to set user data: %v", err), nil, nil
				}
				return jf.TextResult("User data updated."), nil, nil

			default:
				return jf.ErrResult("Invalid action '%s'. Valid actions: favorite, unfavorite, like, dislike, clear_rating, mark_played, mark_unplayed, rate, get_user_data, set_user_data", args.Action), nil, nil
			}
		})
	}

	// --- jellyfin_playlists ---
	if enabled("jellyfin_playlists", AnnotWriteCreate) {
		mcp.AddTool(server, &mcp.Tool{
			Name:  "jellyfin_playlists",
			Title: "Playlists",
			InputSchema: jf.WithEnums[jf.PlaylistsInput](map[string][]any{
				"action": {"list", "create", "generate", "get", "add_items", "remove_items", "move_item", "deduplicate"},
			}),
			Description: "Create and manage playlists. Use 'list' to see all playlists, 'create' to make a new one, or 'generate' to rank library items into a previewed playlist from a natural-language request. Use 'get' to view playlist items, " +
				"'add_items' or 'remove_items' to modify contents, and 'move_item' to reorder. " +
				"Use 'deduplicate' to find and remove duplicate entries (dry_run=true by default for preview, set dry_run=false and confirm=true to remove). " +
				"When creating, provide a name and optional media_type (Audio or Video). For add_items and remove_items, provide the playlist_id and item_ids array. " +
				"For move_item, provide playlist_id, item_id, and new_index (0-based position).",
			Annotations: AnnotWriteCreate,
		}, func(ctx context.Context, req *mcp.CallToolRequest, args jf.PlaylistsInput) (*mcp.CallToolResult, any, error) {
			userID, err := client.GetUserID(ctx)
			if err != nil {
				return jf.ErrResult("Jellyfin error: %v", err), nil, nil
			}

			switch args.Action {
			case "generate":
				if args.Name == "" {
					return jf.ErrResult("name is required for a generated playlist."), nil, nil
				}
				mediaTypes := "Movie,Series"
				if args.MediaType == "Audio" {
					mediaTypes = "Audio,MusicAlbum"
				}
				candidates, err := fetchLocalCandidates(ctx, client, userID, mediaTypes, args.MaxRuntimeMinutes)
				if err != nil {
					return jf.ErrResult("Jellyfin API error: %v", err), nil, nil
				}
				profile := jf.BuildPreferenceProfile(candidates)
				ranked := jf.RankRecommendations(candidates, profile, args.Query, args.MaxRuntimeMinutes, false, jf.ClampInt(args.Limit, 50, jf.MaxLimitCap))
				selected := jf.SelectPlaylistItems(ranked, args.DurationMinutes, args.Limit)
				if len(selected) == 0 {
					return jf.ErrResult("No suitable unwatched items matched the playlist request."), nil, nil
				}
				previewItems := rankedResultItems(selected)
				preview := fmt.Sprintf("Generated playlist preview for %q (%d items):\n\n%s", args.Name, len(previewItems), jf.FormatJSON(previewItems))
				if args.Confirm == nil || !*args.Confirm {
					return jf.TextResult(preview + "\n\nCall again with confirm=true to create it."), nil, nil
				}
				itemIDs := make([]string, 0, len(selected))
				for _, item := range selected {
					itemIDs = append(itemIDs, item.Candidate.ID)
				}
				body := map[string]any{"Name": args.Name, "UserId": userID, "Ids": itemIDs}
				if args.MediaType != "" {
					body["MediaType"] = args.MediaType
				}
				var result map[string]any
				if err := client.Post(ctx, "/Playlists", nil, body, &result); err != nil {
					return jf.ErrResult("Failed to create generated playlist: %v", err), nil, nil
				}
				return jf.TextResult(preview + fmt.Sprintf("\n\nPlaylist created (ID: %s).", jf.GetString(result, "Id"))), nil, nil

			case "list":
				params := url.Values{
					"IncludeItemTypes": {"Playlist"},
					"Recursive":        {"true"},
					"Fields":           {"ChildCount"},
				}
				endpoint := fmt.Sprintf("/Users/%s/Items", jf.SanitizeID(userID))
				rawItems, _, err := jf.FetchAllPages(ctx, client, endpoint, params, 200)
				if err != nil {
					return jf.ErrResult("Jellyfin API error: %v", err), nil, nil
				}
				items := make([]map[string]any, 0, len(rawItems))
				for _, raw := range rawItems {
					m := jf.ToMap(raw)
					items = append(items, map[string]any{
						"id":         jf.GetString(m, "Id"),
						"name":       jf.GetString(m, "Name"),
						"item_count": jf.GetInt(m, "ChildCount"),
					})
				}
				return jf.TextResult(fmt.Sprintf("Found %d playlists:\n\n%s", len(items), jf.FormatJSON(items))), nil, nil

			case "create":
				if args.Name == "" {
					return jf.ErrResult("name is required to create a playlist."), nil, nil
				}
				body := map[string]any{
					"Name":   args.Name,
					"UserId": userID,
				}
				if len(args.ItemIDs) > 0 {
					body["Ids"] = args.ItemIDs
				}
				if args.MediaType != "" {
					body["MediaType"] = args.MediaType
				}
				var result map[string]any
				if err := client.Post(ctx, "/Playlists", nil, body, &result); err != nil {
					return jf.ErrResult("Failed to create playlist: %v", err), nil, nil
				}
				return jf.TextResult(fmt.Sprintf("Playlist created: %s (ID: %s)", args.Name, jf.GetString(result, "Id"))), nil, nil

			case "get":
				if args.PlaylistID == "" {
					return jf.ErrResult("playlist_id is required for 'get' action."), nil, nil
				}
				params := url.Values{
					"UserId": {userID},
					"Fields": {"ProductionYear,RunTimeTicks"},
				}
				endpoint := fmt.Sprintf("/Playlists/%s/Items", jf.SanitizeID(args.PlaylistID))
				rawItems, _, err := jf.FetchAllPages(ctx, client, endpoint, params, 1000)
				if err != nil {
					return jf.ErrResult("Jellyfin API error: %v", err), nil, nil
				}
				items := jf.MapExtract(rawItems, jf.ExtractMediaItem)
				return jf.TextResult(fmt.Sprintf("Playlist items (%d):\n\n%s", len(items), jf.FormatJSON(items))), nil, nil

			case "add_items":
				if args.PlaylistID == "" || len(args.ItemIDs) == 0 {
					return jf.ErrResult("playlist_id and item_ids are required for 'add_items' action."), nil, nil
				}
				params := url.Values{"ids": {jf.JoinIDs(args.ItemIDs)}}
				endpoint := fmt.Sprintf("/Playlists/%s/Items", jf.SanitizeID(args.PlaylistID))
				if err := client.PostNoContent(ctx, endpoint, params, nil); err != nil {
					return jf.ErrResult("Failed to add items: %v", err), nil, nil
				}
				return jf.TextResult(fmt.Sprintf("Added %d items to playlist.", len(args.ItemIDs))), nil, nil

			case "remove_items":
				if args.PlaylistID == "" || len(args.ItemIDs) == 0 {
					return jf.ErrResult("playlist_id and item_ids are required for 'remove_items' action."), nil, nil
				}
				if result := jf.ConfirmationGate(ctx, req, args.Confirm, fmt.Sprintf("This will REMOVE %d items from playlist '%s'.", len(args.ItemIDs), args.PlaylistID)); result != nil {
					return result, nil, nil
				}
				params := url.Values{"entryIds": {jf.JoinIDs(args.ItemIDs)}}
				endpoint := fmt.Sprintf("/Playlists/%s/Items", jf.SanitizeID(args.PlaylistID))
				if err := client.Del(ctx, endpoint, params); err != nil {
					return jf.ErrResult("Failed to remove items: %v", err), nil, nil
				}
				return jf.TextResult(fmt.Sprintf("Removed %d items from playlist.", len(args.ItemIDs))), nil, nil

			case "move_item":
				if args.PlaylistID == "" || args.ItemID == "" || args.NewIndex == nil {
					return jf.ErrResult("playlist_id, item_id, and new_index are required for 'move_item' action."), nil, nil
				}
				endpoint := fmt.Sprintf("/Playlists/%s/Items/%s/Move/%d", jf.SanitizeID(args.PlaylistID), jf.SanitizeID(args.ItemID), *args.NewIndex)
				if err := client.PostNoContent(ctx, endpoint, nil, nil); err != nil {
					return jf.ErrResult("Failed to move item: %v", err), nil, nil
				}
				return jf.TextResult(fmt.Sprintf("Item moved to position %d.", *args.NewIndex)), nil, nil

			case "deduplicate":
				if args.PlaylistID == "" {
					return jf.ErrResult("playlist_id is required for 'deduplicate' action."), nil, nil
				}
				dryRun := args.DryRun == nil || *args.DryRun // default true

				jf.ReportProgress(ctx, req, 0, 2, "Fetching playlist items...")

				params := url.Values{
					"UserId": {userID},
					"Fields": {"ProductionYear"},
				}
				endpoint := fmt.Sprintf("/Playlists/%s/Items", jf.SanitizeID(args.PlaylistID))
				var result map[string]any
				if err := client.Get(ctx, endpoint, params, &result); err != nil {
					return jf.ErrResult("Jellyfin API error: %v", err), nil, nil
				}
				rawItems := jf.ToSlice(result["Items"])

				// Find duplicates by item ID (PlaylistItemId is the entry ID, Id is the media item ID)
				type entry struct {
					playlistItemID string
					name           string
				}
				seen := make(map[string][]entry)
				for _, raw := range rawItems {
					m := jf.ToMap(raw)
					itemID := jf.GetString(m, "Id")
					plItemID := jf.GetString(m, "PlaylistItemId")
					seen[itemID] = append(seen[itemID], entry{playlistItemID: plItemID, name: jf.GetString(m, "Name")})
				}

				var duplicateEntryIDs []string
				var dupReport []map[string]any
				for itemID, entries := range seen {
					if len(entries) > 1 {
						// Keep the first, mark the rest as duplicates
						for _, e := range entries[1:] {
							duplicateEntryIDs = append(duplicateEntryIDs, e.playlistItemID)
						}
						dupReport = append(dupReport, map[string]any{
							"item_id":     itemID,
							"name":        entries[0].name,
							"occurrences": len(entries),
							"removing":    len(entries) - 1,
						})
					}
				}

				if len(duplicateEntryIDs) == 0 {
					return jf.TextResult("No duplicate entries found in playlist."), nil, nil
				}

				if dryRun {
					return jf.TextResult(fmt.Sprintf("Found %d duplicate entries to remove (dry run — no changes made):\n\n%s\n\nTo remove duplicates, call again with dry_run=false and confirm=true.", len(duplicateEntryIDs), jf.FormatJSON(dupReport))), nil, nil
				}

				if result := jf.ConfirmationGate(ctx, req, args.Confirm, fmt.Sprintf("This will REMOVE %d duplicate entries from playlist '%s'.", len(duplicateEntryIDs), args.PlaylistID)); result != nil {
					return result, nil, nil
				}

				jf.ReportProgress(ctx, req, 1, 2, "Removing duplicates...")

				delParams := url.Values{"entryIds": {jf.JoinIDs(duplicateEntryIDs)}}
				if err := client.Del(ctx, endpoint, delParams); err != nil {
					return jf.ErrResult("Failed to remove duplicates: %v", err), nil, nil
				}
				return jf.TextResult(fmt.Sprintf("Removed %d duplicate entries from playlist:\n\n%s", len(duplicateEntryIDs), jf.FormatJSON(dupReport))), nil, nil

			default:
				return jf.ErrResult("Invalid action '%s'. Valid actions: list, create, generate, get, add_items, remove_items, move_item, deduplicate", args.Action), nil, nil
			}
		})
	}

	// --- jellyfin_collections ---
	if enabled("jellyfin_collections", AnnotWriteCreate) {
		mcp.AddTool(server, &mcp.Tool{
			Name:  "jellyfin_collections",
			Title: "Collections",
			InputSchema: jf.WithEnums[jf.CollectionsInput](map[string][]any{
				"action": {"create", "add_items", "remove_items"},
			}),
			Description: "Create and manage box set collections. Collections group related movies together (e.g. a film trilogy). " +
				"Use 'create' with a name and optional item_ids to create a new collection. " +
				"Use 'add_items' or 'remove_items' with a collection_id and item_ids to modify collection contents. " +
				"Collections appear as BoxSet items in browse and search results.",
			Annotations: AnnotWriteCreate,
		}, func(ctx context.Context, req *mcp.CallToolRequest, args jf.CollectionsInput) (*mcp.CallToolResult, any, error) {
			switch args.Action {
			case "create":
				if args.Name == "" {
					return jf.ErrResult("name is required to create a collection."), nil, nil
				}
				params := url.Values{"name": {args.Name}}
				if len(args.ItemIDs) > 0 {
					params.Set("ids", jf.JoinIDs(args.ItemIDs))
				}
				if args.ParentID != "" {
					params.Set("ParentId", args.ParentID)
				}
				var result map[string]any
				if err := client.Post(ctx, "/Collections", params, nil, &result); err != nil {
					return jf.ErrResult("Failed to create collection: %v", err), nil, nil
				}
				return jf.TextResult(fmt.Sprintf("Collection '%s' created (ID: %s).", args.Name, jf.GetString(result, "Id"))), nil, nil

			case "add_items":
				if args.CollectionID == "" || len(args.ItemIDs) == 0 {
					return jf.ErrResult("collection_id and item_ids are required for 'add_items' action."), nil, nil
				}
				params := url.Values{"ids": {jf.JoinIDs(args.ItemIDs)}}
				endpoint := fmt.Sprintf("/Collections/%s/Items", jf.SanitizeID(args.CollectionID))
				if err := client.PostNoContent(ctx, endpoint, params, nil); err != nil {
					return jf.ErrResult("Failed to add items to collection: %v", err), nil, nil
				}
				return jf.TextResult(fmt.Sprintf("Added %d items to collection.", len(args.ItemIDs))), nil, nil

			case "remove_items":
				if args.CollectionID == "" || len(args.ItemIDs) == 0 {
					return jf.ErrResult("collection_id and item_ids are required for 'remove_items' action."), nil, nil
				}
				if result := jf.ConfirmationGate(ctx, req, args.Confirm, fmt.Sprintf("This will REMOVE %d items from collection '%s'.", len(args.ItemIDs), args.CollectionID)); result != nil {
					return result, nil, nil
				}
				params := url.Values{"ids": {jf.JoinIDs(args.ItemIDs)}}
				endpoint := fmt.Sprintf("/Collections/%s/Items", jf.SanitizeID(args.CollectionID))
				if err := client.Del(ctx, endpoint, params); err != nil {
					return jf.ErrResult("Failed to remove items from collection: %v", err), nil, nil
				}
				return jf.TextResult(fmt.Sprintf("Removed %d items from collection.", len(args.ItemIDs))), nil, nil

			default:
				return jf.ErrResult("Invalid action '%s'. Valid actions: create, add_items, remove_items", args.Action), nil, nil
			}
		})
	}
}
