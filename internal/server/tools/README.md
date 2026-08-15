# Tools

31 tools organized into 8 toolsets. Enable specific toolsets with `--toolsets discovery,media,...` or restrict access with `--read-only` and `--disable-destructive`.

## discovery

| Tool | Description |
|------|-------------|
| `jellyfin_libraries` | List all media libraries and their IDs |
| `jellyfin_search` | Search for media by keyword |
| `jellyfin_browse` | Browse and filter by genre, year, studio, person, rating, sort order |
| `jellyfin_get_item` | Full metadata for a specific item (genres, cast, codecs, ratings, provider IDs) |
| `jellyfin_recommendations` | Explainable local recommendations, suggestions, next up, latest additions, and similar items |
| `jellyfin_item_extras` | Special features, trailers, theme songs, intro/outro markers, download URLs |

## media

| Tool | Description |
|------|-------------|
| `jellyfin_tv_shows` | Navigate series structure: seasons, episodes, next up |
| `jellyfin_music` | Browse artists, albums, genres; generate instant mix playlists |
| `jellyfin_people` | Search actors, directors, writers, and studios |

## user

| Tool | Description |
|------|-------------|
| `jellyfin_user_data` | Favorites, ratings, played/unplayed status |
| `jellyfin_playlists` | Create, generate, preview, modify, reorder, and deduplicate playlists |
| `jellyfin_collections` | Create and manage box set collections |

## playback

| Tool | Description |
|------|-------------|
| `jellyfin_sessions` | List active client sessions and resumable items |
| `jellyfin_playback_control` | Play, pause, seek, stop, volume, mute, send messages to clients |
| `jellyfin_play` | Start playback of items on a client (play now, play next, add to queue) |
| `jellyfin_syncplay` | Synchronized group watching sessions |

## admin

| Tool | Description |
|------|-------------|
| `jellyfin_system_info` | Server info, storage, activity logs, log files, playback history |
| `jellyfin_system_control` | Restart or shut down the server (destructive) |
| `jellyfin_users` | Create, delete, update users; manage permissions and Quick Connect |
| `jellyfin_library_manage` | Library scans, metadata refresh, folder management, filesystem browsing |
| `jellyfin_tasks` | View and manage scheduled tasks and triggers |
| `jellyfin_plugins` | Install, configure, enable/disable plugins and repositories |
| `jellyfin_devices` | Manage connected devices and API keys |
| `jellyfin_server` | Read/write server configuration, manage backups |

## content

| Tool | Description |
|------|-------------|
| `jellyfin_metadata` | Search and apply metadata from online providers, manual edits, batch updates |
| `jellyfin_subtitles_lyrics` | Search, download, and manage subtitles and lyrics |
| `jellyfin_images` | List, download, and upload item images |
| `jellyfin_videos` | Merge or split alternate video versions |

## livetv

| Tool | Description |
|------|-------------|
| `jellyfin_live_tv` | Channels, program guide, tuner info |
| `jellyfin_recordings` | DVR recordings, timers, and series recording rules |

## analytics

| Tool | Description |
|------|-------------|
| `jellyfin_analytics` | Library stats, codec reports, duplicates, unplayed items |
