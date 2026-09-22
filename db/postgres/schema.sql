-- Consolidated PostgreSQL schema for Navidrome Serverless (NeonDB)

CREATE TABLE IF NOT EXISTS library (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    path TEXT NOT NULL UNIQUE,
    remote_path TEXT NOT NULL DEFAULT '',
    last_scan_at TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01 00:00:00+00',
    last_scan_started_at TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01 00:00:00+00',
    full_scan_in_progress BOOLEAN NOT NULL DEFAULT FALSE,
    total_songs INT NOT NULL DEFAULT 0,
    total_albums INT NOT NULL DEFAULT 0,
    total_artists INT NOT NULL DEFAULT 0,
    total_folders INT NOT NULL DEFAULT 0,
    total_files INT NOT NULL DEFAULT 0,
    total_missing_files INT NOT NULL DEFAULT 0,
    total_size BIGINT NOT NULL DEFAULT 0,
    total_duration REAL NOT NULL DEFAULT 0,
    default_new_users BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS "user" (
    id VARCHAR(255) PRIMARY KEY,
    user_name VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL DEFAULT '',
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL DEFAULT '',
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    token_epoch INT NOT NULL DEFAULT 0,
    scrobble_filter TEXT NOT NULL DEFAULT '',
    last_login_at TIMESTAMPTZ,
    last_access_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_library (
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    library_id INT NOT NULL REFERENCES library(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, library_id)
);
CREATE INDEX IF NOT EXISTS idx_user_library_user_id ON user_library(user_id);
CREATE INDEX IF NOT EXISTS idx_user_library_library_id ON user_library(library_id);

CREATE TABLE IF NOT EXISTS user_props (
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (user_id, key)
);

CREATE TABLE IF NOT EXISTS artist (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL DEFAULT '',
    sort_artist_name VARCHAR(255) NOT NULL DEFAULT '',
    order_artist_name VARCHAR(255) NOT NULL DEFAULT '',
    mbz_artist_id VARCHAR(255) NOT NULL DEFAULT '',
    biography TEXT NOT NULL DEFAULT '',
    small_image_url VARCHAR(255) NOT NULL DEFAULT '',
    medium_image_url VARCHAR(255) NOT NULL DEFAULT '',
    large_image_url VARCHAR(255) NOT NULL DEFAULT '',
    external_url VARCHAR(255) NOT NULL DEFAULT '',
    similar_artists TEXT NOT NULL DEFAULT '',
    external_info_updated_at TIMESTAMPTZ,
    missing BOOLEAN NOT NULL DEFAULT FALSE,
    uploaded_image VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_artist_name ON artist(name);

CREATE TABLE IF NOT EXISTS library_artist (
    library_id INT NOT NULL REFERENCES library(id) ON DELETE CASCADE,
    artist_id VARCHAR(255) NOT NULL REFERENCES artist(id) ON DELETE CASCADE,
    stats TEXT NOT NULL DEFAULT '{}',
    PRIMARY KEY (library_id, artist_id)
);

CREATE TABLE IF NOT EXISTS album (
    id VARCHAR(255) PRIMARY KEY,
    library_id INT NOT NULL REFERENCES library(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL DEFAULT '',
    artist_id VARCHAR(255) NOT NULL DEFAULT '',
    artist VARCHAR(255) NOT NULL DEFAULT '',
    album_artist VARCHAR(255) NOT NULL DEFAULT '',
    album_artist_id VARCHAR(255) NOT NULL DEFAULT '',
    embed_art_path VARCHAR(255) NOT NULL DEFAULT '',
    cover_art_path VARCHAR(255) NOT NULL DEFAULT '',
    cover_art_id VARCHAR(255) NOT NULL DEFAULT '',
    max_year INT NOT NULL DEFAULT 0,
    min_year INT NOT NULL DEFAULT 0,
    date VARCHAR(255) NOT NULL DEFAULT '',
    max_original_year INT NOT NULL DEFAULT 0,
    min_original_year INT NOT NULL DEFAULT 0,
    original_date VARCHAR(255) NOT NULL DEFAULT '',
    release_date VARCHAR(255) NOT NULL DEFAULT '',
    compilation BOOLEAN NOT NULL DEFAULT FALSE,
    comment TEXT NOT NULL DEFAULT '',
    song_count INT NOT NULL DEFAULT 0,
    duration REAL NOT NULL DEFAULT 0,
    size BIGINT NOT NULL DEFAULT 0,
    discs TEXT NOT NULL DEFAULT '{}',
    sort_album_name VARCHAR(255) NOT NULL DEFAULT '',
    sort_album_artist_name VARCHAR(255) NOT NULL DEFAULT '',
    order_album_name VARCHAR(255) NOT NULL DEFAULT '',
    order_album_artist_name VARCHAR(255) NOT NULL DEFAULT '',
    catalog_num VARCHAR(255) NOT NULL DEFAULT '',
    mbz_album_id VARCHAR(255) NOT NULL DEFAULT '',
    mbz_album_artist_id VARCHAR(255) NOT NULL DEFAULT '',
    mbz_album_type VARCHAR(255) NOT NULL DEFAULT '',
    mbz_album_comment TEXT NOT NULL DEFAULT '',
    mbz_release_group_id VARCHAR(255) NOT NULL DEFAULT '',
    folder_ids TEXT NOT NULL DEFAULT '',
    explicit_status VARCHAR(255) NOT NULL DEFAULT '',
    rg_album_gain DOUBLE PRECISION,
    rg_album_peak DOUBLE PRECISION,
    description TEXT NOT NULL DEFAULT '',
    small_image_url VARCHAR(255) NOT NULL DEFAULT '',
    medium_image_url VARCHAR(255) NOT NULL DEFAULT '',
    large_image_url VARCHAR(255) NOT NULL DEFAULT '',
    external_url VARCHAR(255) NOT NULL DEFAULT '',
    external_info_updated_at TIMESTAMPTZ,
    genre VARCHAR(255) NOT NULL DEFAULT '',
    tags TEXT NOT NULL DEFAULT '',
    missing BOOLEAN NOT NULL DEFAULT FALSE,
    imported_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_album_name ON album(name);
CREATE INDEX IF NOT EXISTS idx_album_artist ON album(artist);
CREATE INDEX IF NOT EXISTS idx_album_artist_id ON album(artist_id);
CREATE INDEX IF NOT EXISTS idx_album_library_id ON album(library_id);
CREATE INDEX IF NOT EXISTS idx_album_genre ON album(genre);
CREATE INDEX IF NOT EXISTS idx_album_min_year ON album(min_year);

CREATE TABLE IF NOT EXISTS album_artists (
    album_id VARCHAR(255) NOT NULL REFERENCES album(id) ON DELETE CASCADE,
    artist_id VARCHAR(255) NOT NULL REFERENCES artist(id) ON DELETE CASCADE,
    role VARCHAR(255) NOT NULL DEFAULT '',
    sub_role VARCHAR(255) NOT NULL DEFAULT '',
    PRIMARY KEY (artist_id, album_id, role, sub_role)
);
CREATE INDEX IF NOT EXISTS idx_album_artists_album_id ON album_artists(album_id);

CREATE TABLE IF NOT EXISTS folder (
    id VARCHAR(255) PRIMARY KEY,
    library_id INT NOT NULL REFERENCES library(id) ON DELETE CASCADE,
    path TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    parent_id VARCHAR(255) NOT NULL DEFAULT '',
    num_audio_files INT NOT NULL DEFAULT 0,
    num_playlists INT NOT NULL DEFAULT 0,
    image_files TEXT NOT NULL DEFAULT '[]',
    images_updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    hash VARCHAR(255) NOT NULL DEFAULT '',
    missing BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_folder_library_id ON folder(library_id);
CREATE INDEX IF NOT EXISTS idx_folder_parent_id ON folder(parent_id);

CREATE TABLE IF NOT EXISTS media_file (
    id VARCHAR(255) PRIMARY KEY,
    pid VARCHAR(255) NOT NULL DEFAULT '',
    library_id INT NOT NULL REFERENCES library(id) ON DELETE CASCADE,
    folder_id VARCHAR(255) NOT NULL DEFAULT '',
    path TEXT NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL DEFAULT '',
    album VARCHAR(255) NOT NULL DEFAULT '',
    artist VARCHAR(255) NOT NULL DEFAULT '',
    artist_id VARCHAR(255) NOT NULL DEFAULT '',
    album_artist VARCHAR(255) NOT NULL DEFAULT '',
    album_artist_id VARCHAR(255) NOT NULL DEFAULT '',
    album_id VARCHAR(255) NOT NULL DEFAULT '' REFERENCES album(id) ON DELETE CASCADE,
    has_cover_art BOOLEAN NOT NULL DEFAULT FALSE,
    track_number INT NOT NULL DEFAULT 0,
    disc_number INT NOT NULL DEFAULT 0,
    disc_subtitle VARCHAR(255) NOT NULL DEFAULT '',
    year INT NOT NULL DEFAULT 0,
    date VARCHAR(255) NOT NULL DEFAULT '',
    original_year INT NOT NULL DEFAULT 0,
    original_date VARCHAR(255) NOT NULL DEFAULT '',
    release_year INT NOT NULL DEFAULT 0,
    release_date VARCHAR(255) NOT NULL DEFAULT '',
    size BIGINT NOT NULL DEFAULT 0,
    suffix VARCHAR(255) NOT NULL DEFAULT '',
    duration REAL NOT NULL DEFAULT 0,
    bit_rate INT NOT NULL DEFAULT 0,
    sample_rate INT NOT NULL DEFAULT 0,
    bit_depth INT,
    channels INT NOT NULL DEFAULT 0,
    codec VARCHAR(255) NOT NULL DEFAULT '',
    probe_data TEXT NOT NULL DEFAULT '',
    genre VARCHAR(255) NOT NULL DEFAULT '',
    sort_title VARCHAR(255) NOT NULL DEFAULT '',
    sort_album_name VARCHAR(255) NOT NULL DEFAULT '',
    sort_artist_name VARCHAR(255) NOT NULL DEFAULT '',
    sort_album_artist_name VARCHAR(255) NOT NULL DEFAULT '',
    order_title VARCHAR(255) NOT NULL DEFAULT '',
    order_album_name VARCHAR(255) NOT NULL DEFAULT '',
    order_artist_name VARCHAR(255) NOT NULL DEFAULT '',
    order_album_artist_name VARCHAR(255) NOT NULL DEFAULT '',
    compilation BOOLEAN NOT NULL DEFAULT FALSE,
    comment TEXT NOT NULL DEFAULT '',
    lyrics TEXT NOT NULL DEFAULT '',
    bpm INT,
    explicit_status VARCHAR(255) NOT NULL DEFAULT '',
    catalog_num VARCHAR(255) NOT NULL DEFAULT '',
    mbz_recording_id VARCHAR(255) NOT NULL DEFAULT '',
    mbz_release_track_id VARCHAR(255) NOT NULL DEFAULT '',
    mbz_album_id VARCHAR(255) NOT NULL DEFAULT '',
    mbz_release_group_id VARCHAR(255) NOT NULL DEFAULT '',
    mbz_artist_id VARCHAR(255) NOT NULL DEFAULT '',
    mbz_album_artist_id VARCHAR(255) NOT NULL DEFAULT '',
    mbz_album_type VARCHAR(255) NOT NULL DEFAULT '',
    mbz_album_comment TEXT NOT NULL DEFAULT '',
    rg_album_gain DOUBLE PRECISION,
    rg_album_peak DOUBLE PRECISION,
    rg_track_gain DOUBLE PRECISION,
    rg_track_peak DOUBLE PRECISION,
    tags TEXT NOT NULL DEFAULT '',
    participants TEXT NOT NULL DEFAULT '',
    missing BOOLEAN NOT NULL DEFAULT FALSE,
    birth_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_media_file_title ON media_file(title);
CREATE INDEX IF NOT EXISTS idx_media_file_album_id ON media_file(album_id);
CREATE INDEX IF NOT EXISTS idx_media_file_artist_id ON media_file(artist_id);
CREATE INDEX IF NOT EXISTS idx_media_file_library_id ON media_file(library_id);
CREATE INDEX IF NOT EXISTS idx_media_file_path ON media_file(path);

CREATE TABLE IF NOT EXISTS media_file_artists (
    media_file_id VARCHAR(255) NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    artist_id VARCHAR(255) NOT NULL REFERENCES artist(id) ON DELETE CASCADE,
    role VARCHAR(255) NOT NULL DEFAULT '',
    sub_role VARCHAR(255) NOT NULL DEFAULT '',
    PRIMARY KEY (artist_id, media_file_id, role, sub_role)
);
CREATE INDEX IF NOT EXISTS idx_media_file_artists_media_file_id ON media_file_artists(media_file_id);

CREATE TABLE IF NOT EXISTS playlist (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT NOT NULL DEFAULT '',
    duration REAL NOT NULL DEFAULT 0,
    size BIGINT NOT NULL DEFAULT 0,
    song_count INT NOT NULL DEFAULT 0,
    owner_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    public BOOLEAN NOT NULL DEFAULT FALSE,
    path TEXT NOT NULL DEFAULT '',
    sync BOOLEAN NOT NULL DEFAULT FALSE,
    uploaded_image VARCHAR(255) NOT NULL DEFAULT '',
    external_image_url VARCHAR(255) NOT NULL DEFAULT '',
    imported_hash VARCHAR(255) NOT NULL DEFAULT '',
    rules TEXT,
    evaluated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_playlist_name ON playlist(name);
CREATE INDEX IF NOT EXISTS idx_playlist_owner_id ON playlist(owner_id);

CREATE TABLE IF NOT EXISTS playlist_tracks (
    id INT NOT NULL,
    playlist_id VARCHAR(255) NOT NULL REFERENCES playlist(id) ON DELETE CASCADE,
    media_file_id VARCHAR(255) NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    PRIMARY KEY (playlist_id, id)
);

CREATE TABLE IF NOT EXISTS annotation (
    ann_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    item_id VARCHAR(255) NOT NULL,
    item_type VARCHAR(255) NOT NULL,
    play_count BIGINT DEFAULT 0,
    play_date TIMESTAMPTZ,
    rating INT DEFAULT 0,
    rated_at TIMESTAMPTZ,
    starred BOOLEAN NOT NULL DEFAULT FALSE,
    starred_at TIMESTAMPTZ,
    UNIQUE (user_id, item_id, item_type)
);
CREATE INDEX IF NOT EXISTS idx_annotation_user_item ON annotation(user_id, item_id);

CREATE TABLE IF NOT EXISTS property (
    id VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS scrobbles (
    id BIGSERIAL PRIMARY KEY,
    media_file_id VARCHAR(255) NOT NULL REFERENCES media_file(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    submission_time BIGINT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_scrobbles_user_time ON scrobbles(user_id, submission_time);

CREATE TABLE IF NOT EXISTS player (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL DEFAULT '',
    user_agent VARCHAR(255) NOT NULL DEFAULT '',
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    client VARCHAR(255) NOT NULL DEFAULT '',
    ip VARCHAR(255) NOT NULL DEFAULT '',
    last_seen TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    transcoding_id VARCHAR(255) NOT NULL DEFAULT '',
    max_bit_rate INT NOT NULL DEFAULT 0,
    report_real_path BOOLEAN NOT NULL DEFAULT FALSE,
    scrobble_enabled BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS transcoding (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL DEFAULT '',
    target_format VARCHAR(255) NOT NULL DEFAULT '',
    command TEXT NOT NULL DEFAULT '',
    default_bit_rate INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS share (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    description TEXT NOT NULL DEFAULT '',
    downloadable BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ,
    last_visited_at TIMESTAMPTZ,
    resource_ids TEXT NOT NULL DEFAULT '',
    resource_type VARCHAR(255) NOT NULL DEFAULT '',
    contents TEXT NOT NULL DEFAULT '',
    format VARCHAR(255) NOT NULL DEFAULT '',
    max_bit_rate INT NOT NULL DEFAULT 0,
    visit_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS radio (
    id VARCHAR(255) PRIMARY KEY,
    stream_url TEXT NOT NULL DEFAULT '',
    name VARCHAR(255) NOT NULL DEFAULT '',
    home_page_url TEXT NOT NULL DEFAULT '',
    uploaded_image VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS artwork (
    hash TEXT PRIMARY KEY,
    mime TEXT NOT NULL,
    width INT NOT NULL DEFAULT 0,
    height INT NOT NULL DEFAULT 0,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    blur_hash TEXT NOT NULL DEFAULT '',
    thumb_hash TEXT NOT NULL DEFAULT '',
    dominant_color TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS item_artwork (
    item_kind TEXT NOT NULL,
    item_id TEXT NOT NULL,
    image_type TEXT NOT NULL DEFAULT 'primary',
    hash TEXT NOT NULL DEFAULT '',
    source TEXT NOT NULL DEFAULT '',
    source_path TEXT NOT NULL DEFAULT '',
    ref_mtime BIGINT NOT NULL DEFAULT 0,
    attempted_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    PRIMARY KEY (item_kind, item_id, image_type)
);
CREATE INDEX IF NOT EXISTS idx_item_artwork_hash ON item_artwork(hash);

CREATE TABLE IF NOT EXISTS artwork_queue (
    item_kind TEXT NOT NULL,
    item_id TEXT NOT NULL,
    image_type TEXT NOT NULL DEFAULT 'primary',
    priority INT NOT NULL DEFAULT 0,
    attempts INT NOT NULL DEFAULT 0,
    retry_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    enqueued_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (item_kind, item_id, image_type)
);
CREATE INDEX IF NOT EXISTS idx_artwork_queue_drain ON artwork_queue(item_kind, priority DESC, enqueued_at, retry_at);
