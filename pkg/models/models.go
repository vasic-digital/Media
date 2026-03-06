// Package models defines generic media domain types used across the
// detection, analysis, and provider pipeline.
//
// These types are framework-agnostic and can be used in any project
// that needs to represent media entities, their files, quality info,
// collections, and search parameters.
package models

import (
	"encoding/json"
	"time"
)

// MediaType represents a category of media content.
type MediaType struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	DetectionPatterns []string  `json:"detection_patterns"`
	MetadataProviders []string  `json:"metadata_providers"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// MediaItem represents a detected media item with aggregated metadata.
type MediaItem struct {
	ID            int64      `json:"id"`
	MediaTypeID   int64      `json:"media_type_id"`
	MediaType     *MediaType `json:"media_type,omitempty"`
	Title         string     `json:"title"`
	OriginalTitle *string    `json:"original_title,omitempty"`
	Year          *int       `json:"year,omitempty"`
	Description   *string    `json:"description,omitempty"`
	Genre         []string   `json:"genre,omitempty"`
	Director      *string    `json:"director,omitempty"`
	CastCrew      *CastCrew  `json:"cast_crew,omitempty"`
	Rating        *float64   `json:"rating,omitempty"`
	Runtime       *int       `json:"runtime,omitempty"`
	Language      *string    `json:"language,omitempty"`
	Country       *string    `json:"country,omitempty"`
	Status        string     `json:"status"`
	ParentID      *int64     `json:"parent_id,omitempty"`
	SeasonNumber  *int       `json:"season_number,omitempty"`
	EpisodeNumber *int       `json:"episode_number,omitempty"`
	TrackNumber   *int       `json:"track_number,omitempty"`
	FirstDetected time.Time  `json:"first_detected"`
	LastUpdated   time.Time  `json:"last_updated"`

	ExternalMetadata []ExternalMetadata `json:"external_metadata,omitempty"`
	Files            []MediaFile        `json:"files,omitempty"`
	Collections      []MediaCollection  `json:"collections,omitempty"`
	UserMetadata     *UserMetadata      `json:"user_metadata,omitempty"`
}

// CastCrew represents cast and crew information.
type CastCrew struct {
	Director   *string  `json:"director,omitempty"`
	Writers    []string `json:"writers,omitempty"`
	Actors     []Actor  `json:"actors,omitempty"`
	Producers  []string `json:"producers,omitempty"`
	Musicians  []string `json:"musicians,omitempty"`
	Developers []string `json:"developers,omitempty"`
}

// Actor represents an actor with their character.
type Actor struct {
	Name      string `json:"name"`
	Character string `json:"character,omitempty"`
	Order     int    `json:"order,omitempty"`
}

// ExternalMetadata represents metadata from external sources.
type ExternalMetadata struct {
	ID          int64     `json:"id"`
	MediaItemID int64     `json:"media_item_id"`
	Provider    string    `json:"provider"`
	ExternalID  string    `json:"external_id"`
	Data        string    `json:"data"`
	Rating      *float64  `json:"rating,omitempty"`
	ReviewURL   *string   `json:"review_url,omitempty"`
	CoverURL    *string   `json:"cover_url,omitempty"`
	TrailerURL  *string   `json:"trailer_url,omitempty"`
	LastFetched time.Time `json:"last_fetched"`
}

// MediaFile represents an individual file linked to a media entity.
type MediaFile struct {
	ID            int64           `json:"id"`
	MediaItemID   int64           `json:"media_item_id"`
	FileID        *int64          `json:"file_id,omitempty"`
	IsPrimary     bool            `json:"is_primary"`
	FilePath      string          `json:"file_path"`
	StorageRoot   string          `json:"storage_root"`
	Filename      string          `json:"filename"`
	FileSize      int64           `json:"file_size"`
	FileExtension *string         `json:"file_extension,omitempty"`
	QualityInfo   *QualityInfo    `json:"quality_info,omitempty"`
	Language      *string         `json:"language,omitempty"`
	Subtitles     []SubtitleTrack `json:"subtitles,omitempty"`
	AudioTracks   []AudioTrack    `json:"audio_tracks,omitempty"`
	Duration      *int            `json:"duration,omitempty"`
	Checksum      *string         `json:"checksum,omitempty"`
	LastVerified  time.Time       `json:"last_verified"`
	CreatedAt     time.Time       `json:"created_at"`
}

// QualityInfo represents file quality information.
type QualityInfo struct {
	Resolution     *Resolution `json:"resolution,omitempty"`
	Bitrate        *int        `json:"bitrate,omitempty"`
	VideoCodec     *string     `json:"video_codec,omitempty"`
	AudioCodec     *string     `json:"audio_codec,omitempty"`
	FrameRate      *float64    `json:"frame_rate,omitempty"`
	AspectRatio    *string     `json:"aspect_ratio,omitempty"`
	ColorDepth     *int        `json:"color_depth,omitempty"`
	HDR            bool        `json:"hdr,omitempty"`
	QualityProfile *string     `json:"quality_profile,omitempty"`
	Source         *string     `json:"source,omitempty"`
	QualityScore   int         `json:"quality_score"`
}

// IsBetterThan returns true if this quality is higher than other.
func (qi *QualityInfo) IsBetterThan(other *QualityInfo) bool {
	if qi == nil || other == nil {
		return qi != nil
	}
	return qi.QualityScore > other.QualityScore
}

// DisplayName returns a human-readable quality name.
func (qi *QualityInfo) DisplayName() string {
	if qi.QualityProfile != nil {
		return *qi.QualityProfile
	}
	if qi.Resolution != nil {
		return qi.Resolution.DisplayName()
	}
	return "Unknown"
}

// Resolution represents video resolution dimensions.
type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// DisplayName returns a human-readable resolution string.
func (r *Resolution) DisplayName() string {
	switch {
	case r.Width >= 3840:
		return "4K/UHD"
	case r.Width >= 1920:
		return "1080p"
	case r.Width >= 1280:
		return "720p"
	case r.Width >= 720:
		return "480p/DVD"
	default:
		return "Low Quality"
	}
}

// SubtitleTrack represents subtitle information.
type SubtitleTrack struct {
	Language string `json:"language"`
	Format   string `json:"format"`
	Forced   bool   `json:"forced,omitempty"`
	Default  bool   `json:"default,omitempty"`
}

// AudioTrack represents audio track information.
type AudioTrack struct {
	Language   string `json:"language"`
	Codec      string `json:"codec"`
	Channels   string `json:"channels"`
	Bitrate    *int   `json:"bitrate,omitempty"`
	SampleRate *int   `json:"sample_rate,omitempty"`
	Default    bool   `json:"default,omitempty"`
	Commentary bool   `json:"commentary,omitempty"`
}

// MediaCollection represents collections of related media.
type MediaCollection struct {
	ID             int64                 `json:"id"`
	Name           string                `json:"name"`
	CollectionType string                `json:"collection_type"`
	Description    *string               `json:"description,omitempty"`
	TotalItems     int                   `json:"total_items"`
	ExternalIDs    map[string]string     `json:"external_ids,omitempty"`
	CoverURL       *string               `json:"cover_url,omitempty"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	Items          []MediaCollectionItem `json:"items,omitempty"`
}

// MediaCollectionItem represents an item within a collection.
type MediaCollectionItem struct {
	ID             int64      `json:"id"`
	CollectionID   int64      `json:"collection_id"`
	MediaItemID    int64      `json:"media_item_id"`
	MediaItem      *MediaItem `json:"media_item,omitempty"`
	SequenceNumber *int       `json:"sequence_number,omitempty"`
	SeasonNumber   *int       `json:"season_number,omitempty"`
	ReleaseOrder   *int       `json:"release_order,omitempty"`
}

// UserMetadata represents user-specific metadata for a media item.
type UserMetadata struct {
	ID            int64      `json:"id"`
	MediaItemID   int64      `json:"media_item_id"`
	UserID        int64      `json:"user_id"`
	UserRating    *float64   `json:"user_rating,omitempty"`
	WatchedStatus *string    `json:"watched_status,omitempty"`
	WatchedDate   *time.Time `json:"watched_date,omitempty"`
	PersonalNotes *string    `json:"personal_notes,omitempty"`
	Tags          []string   `json:"tags,omitempty"`
	Favorite      bool       `json:"favorite"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// SearchRequest represents search parameters for media queries.
type SearchRequest struct {
	Query         string     `json:"query"`
	MediaTypes    []string   `json:"media_types"`
	Year          *int       `json:"year"`
	YearRange     *YearRange `json:"year_range"`
	Genre         []string   `json:"genre"`
	Quality       []string   `json:"quality"`
	Language      []string   `json:"language"`
	MinRating     *float64   `json:"min_rating"`
	HasExternals  *bool      `json:"has_externals"`
	StorageRoots  []string   `json:"storage_roots"`
	WatchedStatus *string    `json:"watched_status"`
	SortBy        string     `json:"sort_by"`
	SortOrder     string     `json:"sort_order"`
	Limit         int        `json:"limit"`
	Offset        int        `json:"offset"`
}

// YearRange represents a range of years.
type YearRange struct {
	From int `json:"from"`
	To   int `json:"to"`
}

// MarshalGenre returns genre as JSON bytes.
func (mi *MediaItem) MarshalGenre() ([]byte, error) {
	return json.Marshal(mi.Genre)
}

// UnmarshalGenre parses genre from JSON bytes.
func (mi *MediaItem) UnmarshalGenre(data []byte) error {
	return json.Unmarshal(data, &mi.Genre)
}

// MarshalCastCrew returns cast/crew as JSON bytes.
func (mi *MediaItem) MarshalCastCrew() ([]byte, error) {
	return json.Marshal(mi.CastCrew)
}

// UnmarshalCastCrew parses cast/crew from JSON bytes.
func (mi *MediaItem) UnmarshalCastCrew(data []byte) error {
	return json.Unmarshal(data, &mi.CastCrew)
}
