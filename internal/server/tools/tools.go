package tools

import (
	"log"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	jf "github.com/jaredtrent/jellyfin-mcp/internal/jellyfin"
)

// ToolsetMap maps toolset names to the tool names they contain.
// Used with --toolsets flag to enable groups of related tools.
var ToolsetMap = map[string][]string{
	"discovery": {
		"jellyfin_libraries",
		"jellyfin_search",
		"jellyfin_browse",
		"jellyfin_get_item",
		"jellyfin_recommendations",
		"jellyfin_item_extras",
		"jellyfin_watchlist",
	},
	"media": {
		"jellyfin_tv_shows",
		"jellyfin_music",
		"jellyfin_people",
	},
	"user": {
		"jellyfin_user_data",
		"jellyfin_playlists",
		"jellyfin_collections",
	},
	"playback": {
		"jellyfin_sessions",
		"jellyfin_playback_control",
		"jellyfin_play",
		"jellyfin_syncplay",
	},
	"admin": {
		"jellyfin_system_info",
		"jellyfin_system_control",
		"jellyfin_users",
		"jellyfin_library_manage",
		"jellyfin_tasks",
		"jellyfin_plugins",
		"jellyfin_devices",
		"jellyfin_server",
	},
	"content": {
		"jellyfin_metadata",
		"jellyfin_subtitles_lyrics",
		"jellyfin_images",
		"jellyfin_videos",
	},
	"livetv": {
		"jellyfin_live_tv",
		"jellyfin_recordings",
	},
	"analytics": {
		"jellyfin_analytics",
	},
}

// Shared tool annotation presets used by all register*Tools functions.
var (
	AnnotReadOnly    = &mcp.ToolAnnotations{ReadOnlyHint: true, IdempotentHint: true, OpenWorldHint: jf.BoolPtr(false)}
	AnnotWriteOp     = &mcp.ToolAnnotations{ReadOnlyHint: false, IdempotentHint: true, OpenWorldHint: jf.BoolPtr(false)}
	AnnotWriteCreate = &mcp.ToolAnnotations{ReadOnlyHint: false, IdempotentHint: false, OpenWorldHint: jf.BoolPtr(false)}
	AnnotDestructive = &mcp.ToolAnnotations{ReadOnlyHint: false, DestructiveHint: jf.BoolPtr(true), OpenWorldHint: jf.BoolPtr(false)}
)

// BuildToolFilter creates the enabled() callback used by all register*Tools functions.
// It supports three filtering dimensions:
//   - toolsets: comma-separated toolset names (e.g. "discovery,media,playback")
//   - readOnly: when true, only tools with ReadOnlyHint=true are registered
//   - disableDestructive: when true, tools with DestructiveHint=true are skipped
func BuildToolFilter(toolsets string, readOnly, disableDestructive bool) func(string, *mcp.ToolAnnotations) bool {
	// Build the set of allowed tool names from toolsets
	var allowed map[string]bool
	if toolsets != "" {
		allowed = make(map[string]bool)
		for _, ts := range strings.Split(toolsets, ",") {
			ts = strings.TrimSpace(ts)
			if ts == "" {
				continue
			}
			tools, ok := ToolsetMap[ts]
			if !ok {
				valid := make([]string, 0, len(ToolsetMap))
				for name := range ToolsetMap {
					valid = append(valid, name)
				}
				log.Printf("WARNING: unknown toolset %q (valid: %s)", ts, strings.Join(valid, ", "))
				continue
			}
			for _, t := range tools {
				allowed[t] = true
			}
		}
	}

	return func(name string, annotations *mcp.ToolAnnotations) bool {
		// Toolset filter: if toolsets specified, tool must be in an enabled set
		if allowed != nil && !allowed[name] {
			return false
		}

		// Annotation-based filters
		if annotations == nil {
			// Tools with no annotations are not read-only and not destructive
			return !readOnly
		}
		if readOnly && !annotations.ReadOnlyHint {
			return false
		}
		if disableDestructive && annotations.DestructiveHint != nil && *annotations.DestructiveHint {
			return false
		}

		return true
	}
}

func RegisterTools(server *mcp.Server, client jf.Client, enabled func(string, *mcp.ToolAnnotations) bool) {
	RegisterDiscoveryTools(server, client, enabled)
	RegisterWatchlistTools(server, client, enabled)
	RegisterMediaTools(server, client, enabled)
	RegisterUserTools(server, client, enabled)
	RegisterPlaybackTools(server, client, enabled)
	RegisterAdminSystemTools(server, client, enabled)
	RegisterAdminUserTools(server, client, enabled)
	RegisterAdminLibraryTools(server, client, enabled)
	RegisterAdminBackendTools(server, client, enabled)
	RegisterLiveTVTools(server, client, enabled)
	RegisterContentTools(server, client, enabled)
	RegisterAnalyticsTools(server, client, enabled)
}
