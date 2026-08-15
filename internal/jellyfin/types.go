package jellyfin

type NoInput struct{}

// --- Discovery & Browsing ---

type SearchInput struct {
	Query string `json:"query" jsonschema:"Search query text to find matching media items"`
	Type  string `json:"type,omitempty" jsonschema:"Filter by item type: Movie, Series, Episode, Audio, MusicAlbum, MusicArtist, Person, Genre"`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum number of results (default 30)"`
}

type BrowseInput struct {
	ParentID        string   `json:"parent_id,omitempty" jsonschema:"Library or folder ID to browse within. Get IDs from jellyfin_libraries"`
	Type            string   `json:"type,omitempty" jsonschema:"Filter by item type: Movie, Series, Episode, Audio, MusicAlbum, BoxSet, Playlist"`
	Genre           string   `json:"genre,omitempty" jsonschema:"Filter by genre name, e.g. Action, Comedy, Drama, Sci-Fi"`
	Year            *int     `json:"year,omitempty" jsonschema:"Filter by production year"`
	Studio          string   `json:"studio,omitempty" jsonschema:"Filter by studio name"`
	Person          string   `json:"person,omitempty" jsonschema:"Filter by person name (actor, director, etc.)"`
	Tags            string   `json:"tags,omitempty" jsonschema:"Filter by tag name"`
	SortBy          string   `json:"sort_by,omitempty" jsonschema:"Sort field: SortName, DateCreated, PremiereDate, CommunityRating, ProductionYear, Runtime, Random, DatePlayed, PlayCount (default SortName)"`
	SortOrder       string   `json:"sort_order,omitempty" jsonschema:"Sort direction: Ascending or Descending (default Ascending)"`
	IsFavorite      *bool    `json:"is_favorite,omitempty" jsonschema:"Filter to favorites only (true) or non-favorites only (false)"`
	IsPlayed        *bool    `json:"is_played,omitempty" jsonschema:"Filter by Jellyfin played flag: true = marked played, false = unplayed. Auto-set when playback reaches the server threshold, or toggled manually"`
	MinRating       *float64 `json:"min_community_rating,omitempty" jsonschema:"Minimum community rating filter (0.0 to 10.0)"`
	OfficialRating  string   `json:"official_rating,omitempty" jsonschema:"Filter by content rating (e.g. G, PG, PG-13, R, TV-MA)"`
	HasSubtitles    *bool    `json:"has_subtitles,omitempty" jsonschema:"Filter items with subtitles (true) or without (false)"`
	MinDateCreated  string   `json:"min_date_created,omitempty" jsonschema:"Items added after this date (ISO 8601, e.g. 2024-01-01)"`
	MaxDateCreated  string   `json:"max_date_created,omitempty" jsonschema:"Items added before this date (ISO 8601)"`
	MinPremiereDate string   `json:"min_premiere_date,omitempty" jsonschema:"Items released after this date (ISO 8601)"`
	MaxPremiereDate string   `json:"max_premiere_date,omitempty" jsonschema:"Items released before this date (ISO 8601)"`
	Limit           int      `json:"limit,omitempty" jsonschema:"Maximum results per page (default 50)"`
	StartIndex      int      `json:"start_index,omitempty" jsonschema:"Starting index for pagination (default 0)"`
}

type GetItemInput struct {
	ID string `json:"id" jsonschema:"Jellyfin item ID (UUID from search or browse results)"`
}

type RecommendationsInput struct {
	Type          string `json:"type" jsonschema:"Recommendation type: personalized (local explainable ranking), next_up (next episodes in progress series), suggestions (Jellyfin suggestions), latest (recently added to library), similar (items like a given item - requires item_id), movie_recs (movie recommendations), upcoming (upcoming TV episodes), recently_played (recently watched items)"`
	ItemID        string `json:"item_id,omitempty" jsonschema:"Item ID required for 'similar' type. Get from search or browse results"`
	ParentID      string `json:"parent_id,omitempty" jsonschema:"Library ID to scope 'latest' results. Get from jellyfin_libraries"`
	Limit         int    `json:"limit,omitempty" jsonschema:"Maximum results (default 25)"`
	Query         string `json:"query,omitempty" jsonschema:"Optional words or mood to match against titles"`
	MaxRuntime    int    `json:"max_runtime_minutes,omitempty" jsonschema:"Optional maximum runtime in minutes"`
	IncludePlayed *bool  `json:"include_played,omitempty" jsonschema:"Include already played items (default false)"`
}

// --- Media Navigation ---

type TVShowsInput struct {
	Action       string `json:"action" jsonschema:"Action to perform: seasons (list all seasons of a series), episodes (list episodes in a season), next_up (next unplayed episode for a series or all series)"`
	SeriesID     string `json:"series_id,omitempty" jsonschema:"Series ID (required for seasons and episodes, optional for next_up to scope to one series)"`
	SeasonNumber *int   `json:"season_number,omitempty" jsonschema:"Season number (required for episodes action, e.g. 1 for Season 1)"`
	Limit        int    `json:"limit,omitempty" jsonschema:"Maximum results for next_up (default 25)"`
}

type MusicInput struct {
	Action string `json:"action" jsonschema:"Action: artists (browse all artists), album_artists (browse album artists only), genres (list music genres), instant_mix (generate auto-playlist from a seed item)"`
	Query  string `json:"query,omitempty" jsonschema:"Filter artists or genres by name"`
	ItemID string `json:"item_id,omitempty" jsonschema:"Seed item ID for instant_mix (album, artist, song, or playlist ID)"`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum results (default 25)"`
}

type PeopleInput struct {
	Action string `json:"action" jsonschema:"Action: persons (list actors, directors, writers) or studios (list production studios)"`
	Query  string `json:"query,omitempty" jsonschema:"Filter by name"`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum results (default 25)"`
}

// --- User Media Management ---

type UserDataInput struct {
	Action        string   `json:"action" jsonschema:"Action: favorite (add to favorites), unfavorite (remove), like (thumbs up), dislike (thumbs down), clear_rating (remove rating), mark_played (set played flag), mark_unplayed (clear played flag), rate (set numeric rating 0-10), get_user_data (read play state), set_user_data (write play state — requires confirm)"`
	ItemID        string   `json:"item_id" jsonschema:"Jellyfin item ID to act on"`
	Rating        *float64 `json:"rating,omitempty" jsonschema:"Numeric rating 0.0-10.0 for rate action"`
	PlayCount     *int     `json:"play_count,omitempty" jsonschema:"Play count for set_user_data"`
	PositionTicks *int64   `json:"position_ticks,omitempty" jsonschema:"Playback position in ticks for set_user_data (10,000,000 ticks = 1 second)"`
	Played        *bool    `json:"played,omitempty" jsonschema:"Played status for set_user_data"`
	Confirm       *bool    `json:"confirm,omitempty" jsonschema:"Set to true to confirm set_user_data"`
}

type PlaylistsInput struct {
	Action            string   `json:"action" jsonschema:"Action: list (all playlists), create (new playlist), generate (rank candidates and preview/create a playlist), get (playlist items), add_items (add to playlist), remove_items (remove from playlist), move_item (reorder), deduplicate (find/remove duplicate entries)"`
	PlaylistID        string   `json:"playlist_id,omitempty" jsonschema:"Playlist ID (required for get, add_items, remove_items, move_item, deduplicate)"`
	Name              string   `json:"name,omitempty" jsonschema:"Playlist name (required for create)"`
	ItemIDs           []string `json:"item_ids,omitempty" jsonschema:"Item IDs to add or remove"`
	ItemID            string   `json:"item_id,omitempty" jsonschema:"Single item ID for move_item"`
	NewIndex          *int     `json:"new_index,omitempty" jsonschema:"New position index for move_item (0-based)"`
	MediaType         string   `json:"media_type,omitempty" jsonschema:"Media type for create: Audio or Video"`
	DryRun            *bool    `json:"dry_run,omitempty" jsonschema:"For deduplicate: report-only mode (default true)"`
	Confirm           *bool    `json:"confirm,omitempty" jsonschema:"Set to true to confirm destructive operations (required for remove_items, deduplicate with dry_run=false)"`
	Query             string   `json:"query,omitempty" jsonschema:"Natural-language playlist intent such as relaxed Sunday or high-energy workout"`
	DurationMinutes   int      `json:"duration_minutes,omitempty" jsonschema:"Target playlist duration in minutes (default 90)"`
	MaxRuntimeMinutes int      `json:"max_runtime_minutes,omitempty" jsonschema:"Maximum runtime per video item"`
	Limit             int      `json:"limit,omitempty" jsonschema:"Maximum generated items (default 25)"`
}

type CollectionsInput struct {
	Action       string   `json:"action" jsonschema:"Action: create (new box set collection), add_items (add items to collection), remove_items (remove items from collection)"`
	CollectionID string   `json:"collection_id,omitempty" jsonschema:"Collection ID (required for add_items, remove_items)"`
	Name         string   `json:"name,omitempty" jsonschema:"Collection name (required for create)"`
	ItemIDs      []string `json:"item_ids,omitempty" jsonschema:"Item IDs to add or remove"`
	ParentID     string   `json:"parent_id,omitempty" jsonschema:"Parent folder ID for create"`
	Confirm      *bool    `json:"confirm,omitempty" jsonschema:"Set to true to confirm remove_items operation"`
}

// --- Playback & Sessions ---

type SessionsInput struct {
	Action string `json:"action" jsonschema:"Action: list (all active sessions with now-playing info and session IDs), resume (items with in-progress playback)"`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum resume items (default 15)"`
}

type PlaybackControlInput struct {
	SessionID string `json:"session_id" jsonschema:"Session ID to control (get from jellyfin_sessions list action)"`
	Command   string `json:"command" jsonschema:"Command: Pause, Unpause, Stop, NextTrack, PreviousTrack, Seek, Mute, Unmute, ToggleMute, SetVolume, SendMessage, GoHome, GoToSettings, ChannelUp, ChannelDown, DisplayContent"`
	SeekTicks *int64 `json:"seek_position_ticks,omitempty" jsonschema:"Position in ticks for Seek (10,000,000 ticks = 1 second)"`
	Volume    *int   `json:"volume,omitempty" jsonschema:"Volume level 0-100 for SetVolume"`
	Message   string `json:"message,omitempty" jsonschema:"Text message for SendMessage"`
	ItemID    string `json:"item_id,omitempty" jsonschema:"Item ID for DisplayContent (navigate client to show an item)"`
}

type PlayInput struct {
	SessionID   string   `json:"session_id" jsonschema:"Target session ID (get from jellyfin_sessions list action)"`
	ItemIDs     []string `json:"item_ids" jsonschema:"Ordered list of item IDs to play"`
	PlayCommand string   `json:"play_command,omitempty" jsonschema:"Queue mode: PlayNow (replace queue and start, default), PlayNext (insert after current), PlayLast (append to end)"`
	StartIndex  *int     `json:"start_index,omitempty" jsonschema:"Index in item_ids to start from (default 0)"`
}

type SyncPlayInput struct {
	Action    string   `json:"action" jsonschema:"Action: list (active groups), new (create group), join (join group), leave (leave group), play, pause, seek, stop, queue (add to queue), next, previous, set_repeat, set_shuffle"`
	GroupID   string   `json:"group_id,omitempty" jsonschema:"SyncPlay group ID (required for join)"`
	ItemIDs   []string `json:"item_ids,omitempty" jsonschema:"Item IDs for queue action"`
	SeekTicks *int64   `json:"seek_position_ticks,omitempty" jsonschema:"Seek position in ticks for seek action"`
	Mode      string   `json:"mode,omitempty" jsonschema:"Mode for set_repeat (RepeatNone, RepeatAll, RepeatOne) or set_shuffle (Sorted, Shuffle)"`
}

// --- Administration ---

type SystemInfoInput struct {
	Action       string `json:"action" jsonschema:"Action: whoami (current user identity and permissions), info (server info with pending restart and update status), storage (disk space), activity_log (user-level events like logins and playback — NOT server errors), ping (check responsiveness), logs (list log files), log_file (read log file content — check for [WRN] and [ERR] tags), playback_history (recently played items with play counts — not an event log), health_check (comprehensive server health report — checks server status, storage, tasks, plugins, logs, and backups in one call)"`
	Limit        int    `json:"limit,omitempty" jsonschema:"Number of activity log entries or log file tail lines (default 25 for activity_log, 200 for log_file)"`
	Name         string `json:"name,omitempty" jsonschema:"Log file name for log_file action (get names from logs action)"`
	UserID       string `json:"user_id,omitempty" jsonschema:"User ID filter for playback_history and activity_log"`
	VerifiedOnly *bool  `json:"verified_only,omitempty" jsonschema:"For playback_history: only return items with verified playback sessions (default true). Set to false to include items manually marked as watched."`
	LogType      string `json:"log_type,omitempty" jsonschema:"Filter for logs action: main (server logs only, default), ffmpeg (transcode logs only), all (everything)"`
	Severity     string `json:"severity,omitempty" jsonschema:"Filter for log_file action: all (default), warn (lines with [WRN]), error (lines with [ERR]), warn+error (both warnings and errors)"`
	MinDate      string `json:"min_date,omitempty" jsonschema:"ISO 8601 date filter for activity_log (e.g. 2024-01-01)"`
	Type         string `json:"type,omitempty" jsonschema:"Type filter for activity_log (e.g. SessionStarted, VideoPlaybackStopped, TaskCompleted, Error)"`
}

type SystemControlInput struct {
	Action  string `json:"action" jsonschema:"Action: restart (restart the Jellyfin server), shutdown (stop the server completely)"`
	Confirm *bool  `json:"confirm,omitempty" jsonschema:"Set to true to confirm destructive operation"`
}

type UsersInput struct {
	Action           string   `json:"action" jsonschema:"Action: list, get, create, delete, update_policy (permissions), update_password, update_config (per-user preferences), qc_status (Quick Connect enabled?), qc_initiate (start Quick Connect), qc_authorize (authorize Quick Connect code)"`
	UserID           string   `json:"user_id,omitempty" jsonschema:"User ID (required for get, delete, update_policy, update_password, update_config)"`
	Username         string   `json:"username,omitempty" jsonschema:"Username (required for create)"`
	Password         string   `json:"password,omitempty" jsonschema:"Password (for create or update_password)"`
	IsAdmin          *bool    `json:"is_admin,omitempty" jsonschema:"Grant admin privileges (for update_policy)"`
	IsDisabled       *bool    `json:"is_disabled,omitempty" jsonschema:"Disable user account (for update_policy)"`
	EnableAllFolders *bool    `json:"enable_all_folders,omitempty" jsonschema:"Access all libraries (for update_policy)"`
	EnabledFolderIDs []string `json:"enabled_folder_ids,omitempty" jsonschema:"Library IDs user can access (for update_policy when enable_all_folders is false)"`
	Confirm          *bool    `json:"confirm,omitempty" jsonschema:"Set to true to confirm delete operation"`
	Config           any      `json:"config,omitempty" jsonschema:"User configuration object for update_config — fetch user first to see current config, modify, then POST"`
	Code             string   `json:"code,omitempty" jsonschema:"Quick Connect code for qc_authorize"`
	SubtitleLanguage string   `json:"subtitle_language,omitempty" jsonschema:"Preferred subtitle language for update_config (ISO 639 code)"`
	AudioLanguage    string   `json:"audio_language,omitempty" jsonschema:"Preferred audio language for update_config (ISO 639 code)"`
	PlayDefaultAudio *bool    `json:"play_default_audio,omitempty" jsonschema:"Play default audio track for update_config"`
}

type LibraryManageInput struct {
	Action          string `json:"action" jsonschema:"Action: scan (scan all libraries), refresh_item (refresh metadata for one item), delete_item (permanently delete), list_folders (list libraries), add_folder (create library), remove_folder (delete library), add_path (add media path), remove_path (remove media path), rename_folder (rename library), update_options (set library options), browse_drives (list server drives), browse_directory (browse server filesystem)"`
	ItemID          string `json:"item_id,omitempty" jsonschema:"Item ID for refresh_item or delete_item"`
	FolderName      string `json:"folder_name,omitempty" jsonschema:"Library name for add_folder, remove_folder, or rename_folder (current name)"`
	NewName         string `json:"new_name,omitempty" jsonschema:"New library name for rename_folder"`
	CollectionType  string `json:"collection_type,omitempty" jsonschema:"Type for add_folder: movies, tvshows, music, musicvideos, homevideos, boxsets, books, mixed"`
	Path            string `json:"path,omitempty" jsonschema:"Filesystem path for add_path, remove_path, or browse_directory"`
	LibraryOptions  any    `json:"library_options,omitempty" jsonschema:"Library options object for update_options — use list_folders to see current options, modify, then POST"`
	ReplaceMetadata *bool  `json:"replace_all_metadata,omitempty" jsonschema:"Replace all metadata on refresh (default false)"`
	ReplaceImages   *bool  `json:"replace_all_images,omitempty" jsonschema:"Replace all images on refresh (default false)"`
	Confirm         *bool  `json:"confirm,omitempty" jsonschema:"Set to true to confirm destructive operations (delete_item, remove_folder, remove_path)"`
}

type TasksInput struct {
	Action   string `json:"action" jsonschema:"Action: list (all scheduled tasks with status), get (task details), start (run a task now), stop (cancel running task), set_triggers (replace task triggers — requires confirm)"`
	TaskID   string `json:"task_id,omitempty" jsonschema:"Task ID (required for get, start, stop, set_triggers)"`
	Triggers any    `json:"triggers,omitempty" jsonschema:"Array of trigger objects for set_triggers. Types: DailyTrigger, WeeklyTrigger, IntervalTrigger, StartupTrigger. Use 'get' first to see current format."`
	Confirm  *bool  `json:"confirm,omitempty" jsonschema:"Set to true to confirm set_triggers"`
}

type PluginsInput struct {
	Action      string `json:"action" jsonschema:"Action: list (installed plugins), enable, disable, uninstall, get_config (plugin configuration), update_config (set plugin config — requires confirm), list_packages (available packages from repositories), install, list_repos (plugin repositories), set_repos (replace all repositories — requires confirm)"`
	PluginID    string `json:"plugin_id,omitempty" jsonschema:"Plugin ID (for enable, disable, uninstall, get_config, update_config)"`
	Version     string `json:"version,omitempty" jsonschema:"Plugin version (for enable, disable, uninstall)"`
	PackageName string `json:"package_name,omitempty" jsonschema:"Package name for install"`
	RepoURL     string `json:"repository_url,omitempty" jsonschema:"Repository URL for install"`
	Confirm     *bool  `json:"confirm,omitempty" jsonschema:"Set to true to confirm uninstall, update_config, or set_repos"`
	Config      any    `json:"config,omitempty" jsonschema:"Plugin configuration object for update_config — use get_config first to see current format"`
	Repos       any    `json:"repos,omitempty" jsonschema:"Array of {Name, Url} objects for set_repos — replaces all repositories. Use list_repos first, modify, then POST back."`
}

type LiveTVInput struct {
	Action    string `json:"action" jsonschema:"Action: channels (list channels), channel (channel details), programs (current/upcoming), recommended (recommended programs), guide_info (TV guide metadata), tuners (available tuner devices)"`
	ChannelID string `json:"channel_id,omitempty" jsonschema:"Channel ID (required for channel action, optional for programs)"`
	Limit     int    `json:"limit,omitempty" jsonschema:"Maximum results (default 25)"`
}

type RecordingsInput struct {
	Action      string `json:"action" jsonschema:"Action: list (recordings), delete, timers (scheduled recordings), create_timer (schedule from program), cancel_timer, series_timers (series rules), create_series_timer, cancel_series_timer"`
	RecordingID string `json:"recording_id,omitempty" jsonschema:"Recording ID for delete"`
	TimerID     string `json:"timer_id,omitempty" jsonschema:"Timer ID for cancel_timer or cancel_series_timer"`
	ProgramID   string `json:"program_id,omitempty" jsonschema:"Program ID for create_timer or create_series_timer"`
	Confirm     *bool  `json:"confirm,omitempty" jsonschema:"Set to true to confirm destructive operations (delete, cancel_timer, cancel_series_timer)"`
}

// --- Metadata & Content ---

type MetadataInput struct {
	Action          string   `json:"action" jsonschema:"Action: search (find metadata online), apply (apply search result), update (edit item fields), batch_update (update multiple items), editor_info (available metadata fields), external_ids (provider ID types)"`
	ItemID          string   `json:"item_id,omitempty" jsonschema:"Item ID (required for apply, update, editor_info, external_ids)"`
	ItemIDs         []string `json:"item_ids,omitempty" jsonschema:"Multiple item IDs for batch_update (max 50)"`
	SearchType      string   `json:"search_type,omitempty" jsonschema:"Type for search: Movie, Series, Person, Book, BoxSet, MusicAlbum, MusicArtist"`
	SearchQuery     string   `json:"search_query,omitempty" jsonschema:"Search text (required for search)"`
	SearchYear      *int     `json:"search_year,omitempty" jsonschema:"Year to narrow search results"`
	ProviderName    string   `json:"provider_name,omitempty" jsonschema:"Provider for apply (e.g. TheMovieDb, Tvdb)"`
	ProviderID      string   `json:"provider_id,omitempty" jsonschema:"Provider-specific ID for apply"`
	Name            string   `json:"name,omitempty" jsonschema:"New name for update"`
	Overview        string   `json:"overview,omitempty" jsonschema:"New overview text for update"`
	Genres          []string `json:"genres,omitempty" jsonschema:"Genre list for update"`
	Tags            []string `json:"tags,omitempty" jsonschema:"Tag list for update"`
	Studios         []string `json:"studios,omitempty" jsonschema:"Studio names for update"`
	Year            *int     `json:"production_year,omitempty" jsonschema:"Production year for update"`
	CommunityRating *float64 `json:"community_rating,omitempty" jsonschema:"Community rating for update/batch_update (0.0 to 10.0)"`
	OfficialRating  string   `json:"official_rating,omitempty" jsonschema:"Content rating for update/batch_update (e.g. G, PG, PG-13, R, TV-MA)"`
	SortName        string   `json:"sort_name,omitempty" jsonschema:"Custom sort name for update/batch_update"`
	LockedFields    []string `json:"locked_fields,omitempty" jsonschema:"Fields to lock from automatic updates (e.g. Name, Overview, Genres)"`
	DryRun          *bool    `json:"dry_run,omitempty" jsonschema:"For batch_update: preview changes without applying (default true). Set to false with confirm=true to apply."`
	Confirm         *bool    `json:"confirm,omitempty" jsonschema:"Set to true to confirm batch_update operation"`
}

type SubtitlesLyricsInput struct {
	Action            string   `json:"action" jsonschema:"Action: search_subtitles, download_subtitle, delete_subtitle, batch_download_subtitles (search and download for multiple items), upload_subtitle (upload subtitle file as base64), get_lyrics, search_lyrics, download_lyrics, delete_lyrics"`
	ItemID            string   `json:"item_id,omitempty" jsonschema:"Media item ID"`
	ItemIDs           []string `json:"item_ids,omitempty" jsonschema:"Multiple item IDs for batch_download_subtitles (max 25)"`
	Language          string   `json:"language,omitempty" jsonschema:"Language code for search (e.g. en, es, fr, de)"`
	SubtitleID        string   `json:"subtitle_id,omitempty" jsonschema:"Subtitle ID from search results (for download_subtitle)"`
	LyricID           string   `json:"lyric_id,omitempty" jsonschema:"Lyric ID from search results (for download_lyrics)"`
	SubtitleIndex     *int     `json:"subtitle_index,omitempty" jsonschema:"Subtitle stream index (for delete_subtitle)"`
	SubtitleData      string   `json:"subtitle_data,omitempty" jsonschema:"Base64-encoded subtitle file content for upload_subtitle"`
	SubtitleFormat    string   `json:"subtitle_format,omitempty" jsonschema:"Subtitle format for upload_subtitle (e.g. srt, ass, vtt)"`
	SubtitleLanguage  string   `json:"subtitle_language,omitempty" jsonschema:"Language code for upload_subtitle (e.g. eng, spa, fre)"`
	IsForced          *bool    `json:"is_forced,omitempty" jsonschema:"Mark subtitle as forced for upload_subtitle"`
	IsHearingImpaired *bool    `json:"is_hearing_impaired,omitempty" jsonschema:"Mark subtitle as hearing impaired (SDH) for upload_subtitle"`
	Confirm           *bool    `json:"confirm,omitempty" jsonschema:"Set to true to confirm destructive operations (delete_subtitle, delete_lyrics, batch_download_subtitles)"`
}

type ImagesInput struct {
	Action     string `json:"action" jsonschema:"Action: list (images for item), get_url (direct image URL), remote_list (browse online images), remote_download (download remote image), upload (upload image from base64 data)"`
	ItemID     string `json:"item_id" jsonschema:"Item ID"`
	ImageType  string `json:"image_type,omitempty" jsonschema:"Image type: Primary, Backdrop, Logo, Thumb, Banner, Art, Disc (default Primary)"`
	Provider   string `json:"provider_name,omitempty" jsonschema:"Provider for remote_list (e.g. TheMovieDb)"`
	ImageURL   string `json:"image_url,omitempty" jsonschema:"Remote image URL for remote_download"`
	ImageIndex *int   `json:"image_index,omitempty" jsonschema:"Image index for multiple images of same type (default 0)"`
	ImageData  string `json:"image_data,omitempty" jsonschema:"Base64-encoded image data for upload (JPEG or PNG)"`
}

type DevicesInput struct {
	Action   string `json:"action" jsonschema:"Action: list (connected devices), get (device info), delete (remove device), api_keys (list API keys), create_api_key, revoke_api_key"`
	DeviceID string `json:"device_id,omitempty" jsonschema:"Device ID for get or delete"`
	Key      string `json:"key,omitempty" jsonschema:"API key to revoke"`
	AppName  string `json:"app_name,omitempty" jsonschema:"Application name for create_api_key"`
	Limit    int    `json:"limit,omitempty" jsonschema:"Maximum results (default 50)"`
	Days     int    `json:"days,omitempty" jsonschema:"For list: only show devices active within N days (default 30)"`
	Confirm  *bool  `json:"confirm,omitempty" jsonschema:"Set to true to confirm destructive operations (delete, revoke_api_key)"`
}

// --- Server Management ---

type ServerManageInput struct {
	Action   string `json:"action" jsonschema:"Action: get_config (full server config), get_config_section (read named section), update_config_section (replace section — requires confirm), list_backups (existing backups), create_backup (create new backup), restore_backup (restore from backup — requires confirm)"`
	Key      string `json:"key,omitempty" jsonschema:"Config section key for get_config_section and update_config_section (e.g. 'encoding' for transcoding settings)"`
	Config   any    `json:"config,omitempty" jsonschema:"Complete config object for update_config_section — replaces the entire section"`
	FileName string `json:"file_name,omitempty" jsonschema:"Backup archive filename for restore_backup"`
	Confirm  *bool  `json:"confirm,omitempty" jsonschema:"Set to true to confirm update_config_section or restore_backup"`
}

// --- Item Extras ---

type ItemExtrasInput struct {
	Action string `json:"action" jsonschema:"Action: playback_info (transcoding/direct play diagnostics), special_features (behind-the-scenes, extras), theme_songs, theme_videos, local_trailers, segments (intro/outro markers), download_url (construct download link)"`
	ItemID string `json:"item_id" jsonschema:"Item ID"`
}

// --- Video Management ---

type VideosInput struct {
	Action  string   `json:"action" jsonschema:"Action: merge_versions (combine multiple files as alternate versions), split_versions (undo merge)"`
	ItemIDs []string `json:"item_ids,omitempty" jsonschema:"Item IDs to merge (2+ required for merge_versions)"`
	ItemID  string   `json:"item_id,omitempty" jsonschema:"Item ID with alternate sources (for split_versions)"`
	Confirm *bool    `json:"confirm,omitempty" jsonschema:"Set to true to confirm merge or split operation"`
}

// --- Analytics ---

type AnalyticsInput struct {
	Action   string `json:"action" jsonschema:"Action: library_stats (counts by type), library_size (total storage by type), size_report (rank individual items or series by total storage size — defaults to Series), codec_report (codec/resolution distribution), never_played (unplayed items), recently_added (items added within N days), duplicate_check (items with same name and year), played_status (per-user played/unplayed status for an item — requires item_id)"`
	ItemID   string `json:"item_id,omitempty" jsonschema:"Item ID for played_status (series, movie, or other item — get from search or browse)"`
	ParentID string `json:"parent_id,omitempty" jsonschema:"Library ID to scope results (get from jellyfin_libraries)"`
	Type     string `json:"type,omitempty" jsonschema:"Item type filter: Movie, Series, Episode, Audio, MusicAlbum"`
	Limit    int    `json:"limit,omitempty" jsonschema:"Maximum results (default 50)"`
	Days     int    `json:"days,omitempty" jsonschema:"Lookback period in days for recently_added (default 30)"`
}
