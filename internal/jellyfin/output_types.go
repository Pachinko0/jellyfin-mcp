package jellyfin

type MediaItem struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Type              string  `json:"type"`
	Year              int     `json:"year,omitempty"`
	Overview          string  `json:"overview,omitempty"`
	CommunityRating   float64 `json:"community_rating,omitempty"`
	OfficialRating    string  `json:"official_rating,omitempty"`
	RuntimeMinutes    int64   `json:"runtime_minutes,omitempty"`
	Played            bool    `json:"played,omitempty"`
	Progress          string  `json:"progress,omitempty"`
	Favorite          bool    `json:"favorite,omitempty"`
	SeriesName        string  `json:"series_name,omitempty"`
	IndexNumber       int     `json:"index_number,omitempty"`
	ParentIndexNumber int     `json:"parent_index_number,omitempty"`
	LastPlayed        string  `json:"last_played,omitempty"`
	DateAdded         string  `json:"date_added,omitempty"`
}

type ItemListOutput struct {
	TotalCount int         `json:"total_count"`
	Shown      int         `json:"shown"`
	Items      []MediaItem `json:"items"`
}

type WatchlistOutput struct {
	TotalCount int              `json:"total_count"`
	Page       int              `json:"page"`
	TotalPages int              `json:"total_pages"`
	Items      []map[string]any `json:"items"`
}

// --- Libraries ---

type LibraryInfo struct {
	Name           string   `json:"name"`
	CollectionType string   `json:"collection_type"`
	ItemID         string   `json:"item_id"`
	Paths          []string `json:"paths,omitempty"`
}

type LibraryListOutput struct {
	Count     int           `json:"count"`
	Libraries []LibraryInfo `json:"libraries"`
}

// --- Sessions ---

type NowPlayingInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	RuntimeMinutes int64  `json:"runtime_minutes,omitempty"`
}

type PlayStateInfo struct {
	IsPaused        bool  `json:"is_paused,omitempty"`
	PositionSeconds int64 `json:"position_seconds,omitempty"`
	PositionTicks   int64 `json:"position_ticks,omitempty"`
	Volume          int   `json:"volume,omitempty"`
}

type SessionInfo struct {
	SessionID    string          `json:"session_id"`
	User         string          `json:"user"`
	Client       string          `json:"client"`
	DeviceName   string          `json:"device_name"`
	LastActivity string          `json:"last_activity"`
	NowPlaying   *NowPlayingInfo `json:"now_playing,omitempty"`
	PlayState    *PlayStateInfo  `json:"play_state,omitempty"`
}

type SessionsOutput struct {
	Sessions []SessionInfo `json:"sessions,omitempty"`
	Resume   []MediaItem   `json:"resume,omitempty"`
}

// --- TV Shows ---

type SeasonInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	SeasonNumber int    `json:"season_number"`
	EpisodeCount int    `json:"episode_count,omitempty"`
}

type EpisodeInfo struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	SeasonNumber    int     `json:"season_number,omitempty"`
	EpisodeNumber   int     `json:"episode_number,omitempty"`
	Overview        string  `json:"overview,omitempty"`
	CommunityRating float64 `json:"community_rating,omitempty"`
	RuntimeMinutes  int64   `json:"runtime_minutes,omitempty"`
	Played          bool    `json:"played,omitempty"`
	Progress        string  `json:"progress,omitempty"`
	PremiereDate    string  `json:"premiere_date,omitempty"`
}

type TVShowsOutput struct {
	Seasons  []SeasonInfo  `json:"seasons,omitempty"`
	Episodes []EpisodeInfo `json:"episodes,omitempty"`
	NextUp   []MediaItem   `json:"next_up,omitempty"`
}

// --- Recommendations ---

type RecommendationsOutput struct {
	Items  []MediaItem      `json:"items"`
	Ranked []map[string]any `json:"ranked,omitempty"`
}

// --- Detailed Item ---

type PersonInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Role string `json:"role,omitempty"`
}

type UserDataInfo struct {
	Played           bool    `json:"played"`
	Favorite         bool    `json:"favorite"`
	PlayCount        int     `json:"play_count,omitempty"`
	PlayedPercentage float64 `json:"played_percentage,omitempty"`
	LastPlayed       string  `json:"last_played,omitempty"`
}

type AudioStreamInfo struct {
	Index        int    `json:"index"`
	Codec        string `json:"codec"`
	Channels     int    `json:"channels,omitempty"`
	Language     string `json:"language,omitempty"`
	DisplayTitle string `json:"display_title,omitempty"`
	IsDefault    bool   `json:"is_default,omitempty"`
}

type SubtitleStreamInfo struct {
	Index        int    `json:"index"`
	Codec        string `json:"codec"`
	Language     string `json:"language,omitempty"`
	DisplayTitle string `json:"display_title,omitempty"`
	IsExternal   bool   `json:"is_external,omitempty"`
	IsDefault    bool   `json:"is_default,omitempty"`
	IsForced     bool   `json:"is_forced,omitempty"`
}

type MediaSourceInfo struct {
	Container       string               `json:"container"`
	Path            string               `json:"path,omitempty"`
	BitrateKbps     int64                `json:"bitrate_kbps,omitempty"`
	SizeMB          int64                `json:"size_mb,omitempty"`
	VideoCodec      string               `json:"video_codec,omitempty"`
	Resolution      string               `json:"resolution,omitempty"`
	VideoProfile    string               `json:"video_profile,omitempty"`
	VideoBitDepth   int                  `json:"video_bit_depth,omitempty"`
	VideoRange      string               `json:"video_range,omitempty"`
	AudioStreams    []AudioStreamInfo    `json:"audio_streams,omitempty"`
	SubtitleStreams []SubtitleStreamInfo `json:"subtitle_streams,omitempty"`
}

type DetailedItemOutput struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Type               string            `json:"type"`
	Year               int               `json:"year,omitempty"`
	Overview           string            `json:"overview,omitempty"`
	CommunityRating    float64           `json:"community_rating,omitempty"`
	OfficialRating     string            `json:"official_rating,omitempty"`
	CriticRating       float64           `json:"critic_rating,omitempty"`
	RuntimeMinutes     int64             `json:"runtime_minutes,omitempty"`
	PremiereDate       string            `json:"premiere_date,omitempty"`
	EndDate            string            `json:"end_date,omitempty"`
	Taglines           []string          `json:"taglines,omitempty"`
	Genres             []string          `json:"genres,omitempty"`
	Studios            []string          `json:"studios,omitempty"`
	People             []PersonInfo      `json:"people,omitempty"`
	ProviderIDs        map[string]string `json:"provider_ids,omitempty"`
	ExternalURLs       map[string]string `json:"external_urls,omitempty"`
	UserData           *UserDataInfo     `json:"user_data,omitempty"`
	FilePath           string            `json:"file_path,omitempty"`
	FileName           string            `json:"file_name,omitempty"`
	MediaSources       []MediaSourceInfo `json:"media_sources,omitempty"`
	Artists            []string          `json:"artists,omitempty"`
	Album              string            `json:"album,omitempty"`
	Status             string            `json:"status,omitempty"`
	HasSubtitles       bool              `json:"has_subtitles,omitempty"`
	HasLyrics          bool              `json:"has_lyrics,omitempty"`
	ChildCount         int               `json:"child_count,omitempty"`
	RecursiveItemCount int               `json:"recursive_item_count,omitempty"`
	SeriesName         string            `json:"series_name,omitempty"`
	IndexNumber        int               `json:"index_number,omitempty"`
	ParentIndexNumber  int               `json:"parent_index_number,omitempty"`
	Played             bool              `json:"played,omitempty"`
	Progress           string            `json:"progress,omitempty"`
	Favorite           bool              `json:"favorite,omitempty"`
}

// --- Analytics ---

type TypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type TypeSize struct {
	Type   string `json:"type"`
	Files  int    `json:"files"`
	SizeGB string `json:"size_gb"`
	SizeMB int64  `json:"size_mb"`
}

type CodecDistribution struct {
	TotalMediaSources int            `json:"total_media_sources"`
	VideoCodecs       map[string]int `json:"video_codecs"`
	AudioCodecs       map[string]int `json:"audio_codecs"`
	Containers        map[string]int `json:"containers"`
	Resolutions       map[string]int `json:"resolutions"`
	VideoRanges       map[string]int `json:"video_ranges"`
	BitDepths         map[string]int `json:"bit_depths"`
}

type SizeEntry struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Type   string `json:"type"`
	Files  int    `json:"files"`
	SizeGB string `json:"size_gb"`
	SizeMB int64  `json:"size_mb"`
}

type PlayedStatusUser struct {
	Name           string `json:"name"`
	Played         bool   `json:"played,omitempty"`
	PlayCount      int    `json:"play_count,omitempty"`
	LastPlayed     string `json:"last_played,omitempty"`
	EpisodesPlayed int    `json:"episodes_played,omitempty"`
	TotalEpisodes  int    `json:"total_episodes,omitempty"`
}

type DuplicateGroup struct {
	Name   string          `json:"name"`
	Year   int             `json:"year,omitempty"`
	Count  int             `json:"count"`
	Copies []DuplicateCopy `json:"copies"`
}

type DuplicateCopy struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Year int    `json:"year,omitempty"`
	Path string `json:"path,omitempty"`
}

type AnalyticsOutput struct {
	// library_stats
	Stats []TypeCount `json:"stats,omitempty"`
	// library_size
	TotalSizeGB string     `json:"total_size_gb,omitempty"`
	TotalSizeMB int64      `json:"total_size_mb,omitempty"`
	ByType      []TypeSize `json:"by_type,omitempty"`
	// size_report
	SizeReport []SizeEntry `json:"size_report,omitempty"`
	// codec_report
	CodecReport *CodecDistribution `json:"codec_report,omitempty"`
	// never_played, recently_added
	Items      []MediaItem `json:"items,omitempty"`
	TotalCount int         `json:"total_count,omitempty"`
	// duplicate_check
	Duplicates []DuplicateGroup `json:"duplicates,omitempty"`
	// played_status
	ItemSummary *PlayedStatusItem  `json:"item_summary,omitempty"`
	Users       []PlayedStatusUser `json:"users,omitempty"`
}

type PlayedStatusItem struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	TotalEpisodes int    `json:"total_episodes,omitempty"`
}

func ToMediaItem(m map[string]any) MediaItem {
	item := MediaItem{
		ID:   GetString(m, "id"),
		Name: GetString(m, "name"),
		Type: GetString(m, "type"),
	}
	item.Year = GetNum[int](m, "year")
	item.Overview = GetString(m, "overview")
	item.CommunityRating = GetNum[float64](m, "community_rating")
	item.OfficialRating = GetString(m, "official_rating")
	item.RuntimeMinutes = GetNum[int64](m, "runtime_minutes")
	item.Played = GetBool(m, "played")
	item.Progress = GetString(m, "progress")
	item.Favorite = GetBool(m, "favorite")
	item.SeriesName = GetString(m, "series_name")
	item.IndexNumber = GetNum[int](m, "index_number")
	item.ParentIndexNumber = GetNum[int](m, "parent_index_number")
	item.LastPlayed = GetString(m, "last_played")
	item.DateAdded = GetString(m, "date_added")
	return item
}

func ToMediaItems(items []map[string]any) []MediaItem {
	out := make([]MediaItem, 0, len(items))
	for _, m := range items {
		out = append(out, ToMediaItem(m))
	}
	return out
}

func ToLibraryInfo(m map[string]any) LibraryInfo {
	info := LibraryInfo{
		Name:           GetString(m, "name"),
		CollectionType: GetString(m, "collection_type"),
		ItemID:         GetString(m, "item_id"),
	}
	if paths, ok := m["paths"].([]string); ok {
		info.Paths = paths
	} else if raw, ok := m["paths"].([]any); ok {
		for _, p := range raw {
			if s, ok := p.(string); ok {
				info.Paths = append(info.Paths, s)
			}
		}
	}
	return info
}

func ToLibraryInfos(items []map[string]any) []LibraryInfo {
	out := make([]LibraryInfo, 0, len(items))
	for _, m := range items {
		out = append(out, ToLibraryInfo(m))
	}
	return out
}

func ToSessionInfo(m map[string]any) SessionInfo {
	s := SessionInfo{
		SessionID:    GetString(m, "session_id"),
		User:         GetString(m, "user"),
		Client:       GetString(m, "client"),
		DeviceName:   GetString(m, "device_name"),
		LastActivity: GetString(m, "last_activity"),
	}
	if np, ok := m["now_playing"].(map[string]any); ok {
		info := &NowPlayingInfo{
			ID:   GetString(np, "id"),
			Name: GetString(np, "name"),
			Type: GetString(np, "type"),
		}
		info.RuntimeMinutes = GetNum[int64](np, "runtime_minutes")
		s.NowPlaying = info
	}
	if ps, ok := m["play_state"].(map[string]any); ok {
		state := &PlayStateInfo{}
		state.IsPaused = GetBool(ps, "is_paused")
		state.PositionSeconds = GetNum[int64](ps, "position_seconds")
		state.PositionTicks = GetNum[int64](ps, "position_ticks")
		state.Volume = GetNum[int](ps, "volume")
		s.PlayState = state
	}
	return s
}

func ToSessionInfos(items []map[string]any) []SessionInfo {
	out := make([]SessionInfo, 0, len(items))
	for _, m := range items {
		out = append(out, ToSessionInfo(m))
	}
	return out
}

func ToDetailedItemOutput(m map[string]any) *DetailedItemOutput {
	o := &DetailedItemOutput{
		ID:             GetString(m, "id"),
		Name:           GetString(m, "name"),
		Type:           GetString(m, "type"),
		Overview:       GetString(m, "overview"),
		OfficialRating: GetString(m, "official_rating"),
		PremiereDate:   GetString(m, "premiere_date"),
		EndDate:        GetString(m, "end_date"),
		FilePath:       GetString(m, "file_path"),
		FileName:       GetString(m, "file_name"),
		Album:          GetString(m, "album"),
		Status:         GetString(m, "status"),
		SeriesName:     GetString(m, "series_name"),
		Progress:       GetString(m, "progress"),
	}
	o.Year = GetNum[int](m, "year")
	o.CommunityRating = GetNum[float64](m, "community_rating")
	o.CriticRating = GetNum[float64](m, "critic_rating")
	o.RuntimeMinutes = GetNum[int64](m, "runtime_minutes")
	o.Played = GetBool(m, "played")
	o.Favorite = GetBool(m, "favorite")
	o.HasSubtitles = GetBool(m, "has_subtitles")
	o.HasLyrics = GetBool(m, "has_lyrics")
	o.IndexNumber = GetNum[int](m, "index_number")
	o.ParentIndexNumber = GetNum[int](m, "parent_index_number")
	o.ChildCount = GetNum[int](m, "child_count")
	o.RecursiveItemCount = GetNum[int](m, "recursive_item_count")
	o.Taglines = ToStringSlice(m["taglines"])
	o.Genres = ToStringSlice(m["genres"])
	o.Studios = ToStringSlice(m["studios"])
	o.Artists = ToStringSlice(m["artists"])

	if people, ok := m["people"].([]map[string]any); ok {
		for _, p := range people {
			o.People = append(o.People, PersonInfo{
				Name: GetString(p, "name"),
				Type: GetString(p, "type"),
				Role: GetString(p, "role"),
			})
		}
	}
	if pids, ok := m["provider_ids"].(map[string]string); ok {
		o.ProviderIDs = pids
	}
	if urls, ok := m["external_urls"].(map[string]string); ok {
		o.ExternalURLs = urls
	}
	if ud, ok := m["user_data"].(map[string]any); ok {
		info := &UserDataInfo{}
		info.Played = GetBool(ud, "played")
		info.Favorite = GetBool(ud, "favorite")
		info.PlayCount = GetNum[int](ud, "play_count")
		info.PlayedPercentage = GetNum[float64](ud, "played_percentage")
		info.LastPlayed = GetString(ud, "last_played")
		o.UserData = info
	}
	if sources, ok := m["media_sources"].([]map[string]any); ok {
		for _, src := range sources {
			ms := MediaSourceInfo{
				Container:    GetString(src, "container"),
				Path:         GetString(src, "path"),
				VideoCodec:   GetString(src, "video_codec"),
				Resolution:   GetString(src, "resolution"),
				VideoProfile: GetString(src, "video_profile"),
				VideoRange:   GetString(src, "video_range"),
			}
			ms.BitrateKbps = GetNum[int64](src, "bitrate_kbps")
			ms.SizeMB = GetNum[int64](src, "size_mb")
			ms.VideoBitDepth = GetNum[int](src, "video_bit_depth")
			if audioRaw, ok := src["audio_streams"].([]map[string]any); ok {
				for _, a := range audioRaw {
					as := AudioStreamInfo{
						Codec:        GetString(a, "codec"),
						Language:     GetString(a, "language"),
						DisplayTitle: GetString(a, "display_title"),
					}
					as.Index = GetNum[int](a, "index")
					as.Channels = GetNum[int](a, "channels")
					as.IsDefault = GetBool(a, "is_default")
					ms.AudioStreams = append(ms.AudioStreams, as)
				}
			}
			if subRaw, ok := src["subtitle_streams"].([]map[string]any); ok {
				for _, s := range subRaw {
					ss := SubtitleStreamInfo{
						Codec:        GetString(s, "codec"),
						Language:     GetString(s, "language"),
						DisplayTitle: GetString(s, "display_title"),
					}
					ss.Index = GetNum[int](s, "index")
					ss.IsExternal = GetBool(s, "is_external")
					ss.IsDefault = GetBool(s, "is_default")
					ss.IsForced = GetBool(s, "is_forced")
					ms.SubtitleStreams = append(ms.SubtitleStreams, ss)
				}
			}
			o.MediaSources = append(o.MediaSources, ms)
		}
	}
	return o
}
