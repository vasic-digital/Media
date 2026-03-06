package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolution_DisplayName(t *testing.T) {
	tests := []struct {
		name   string
		res    Resolution
		expect string
	}{
		{"4K", Resolution{3840, 2160}, "4K/UHD"},
		{"1080p", Resolution{1920, 1080}, "1080p"},
		{"720p", Resolution{1280, 720}, "720p"},
		{"480p", Resolution{720, 480}, "480p/DVD"},
		{"low", Resolution{320, 240}, "Low Quality"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.res.DisplayName())
		})
	}
}

func TestQualityInfo_IsBetterThan(t *testing.T) {
	high := &QualityInfo{QualityScore: 90}
	low := &QualityInfo{QualityScore: 50}

	assert.True(t, high.IsBetterThan(low))
	assert.False(t, low.IsBetterThan(high))
	assert.True(t, high.IsBetterThan(nil))
	assert.False(t, (*QualityInfo)(nil).IsBetterThan(high))
}

func TestQualityInfo_DisplayName(t *testing.T) {
	profile := "BluRay-1080p"
	qi := &QualityInfo{QualityProfile: &profile}
	assert.Equal(t, "BluRay-1080p", qi.DisplayName())

	qi2 := &QualityInfo{Resolution: &Resolution{1920, 1080}}
	assert.Equal(t, "1080p", qi2.DisplayName())

	qi3 := &QualityInfo{}
	assert.Equal(t, "Unknown", qi3.DisplayName())
}

func TestMediaItem_MarshalGenre(t *testing.T) {
	mi := &MediaItem{Genre: []string{"Action", "Sci-Fi"}}
	data, err := mi.MarshalGenre()
	require.NoError(t, err)

	mi2 := &MediaItem{}
	err = mi2.UnmarshalGenre(data)
	require.NoError(t, err)
	assert.Equal(t, mi.Genre, mi2.Genre)
}

func TestMediaItem_MarshalCastCrew(t *testing.T) {
	director := "Nolan"
	mi := &MediaItem{
		CastCrew: &CastCrew{
			Director: &director,
			Actors: []Actor{
				{Name: "Actor A", Character: "Hero", Order: 1},
			},
		},
	}
	data, err := mi.MarshalCastCrew()
	require.NoError(t, err)

	mi2 := &MediaItem{}
	err = mi2.UnmarshalCastCrew(data)
	require.NoError(t, err)
	assert.Equal(t, mi.CastCrew.Director, mi2.CastCrew.Director)
	assert.Len(t, mi2.CastCrew.Actors, 1)
}

func TestMediaItem_JSON(t *testing.T) {
	year := 2020
	mi := MediaItem{
		ID:    1,
		Title: "Test Movie",
		Year:  &year,
		Genre: []string{"Drama"},
	}

	data, err := json.Marshal(mi)
	require.NoError(t, err)

	var decoded MediaItem
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, mi.Title, decoded.Title)
	assert.Equal(t, *mi.Year, *decoded.Year)
}

func TestSearchRequest_Defaults(t *testing.T) {
	sr := SearchRequest{}
	assert.Empty(t, sr.Query)
	assert.Equal(t, 0, sr.Limit)
	assert.Nil(t, sr.YearRange)
}

func TestMediaCollection_JSON(t *testing.T) {
	mc := MediaCollection{
		ID:             1,
		Name:           "MCU",
		CollectionType: "franchise",
		ExternalIDs:    map[string]string{"tmdb": "131292"},
	}

	data, err := json.Marshal(mc)
	require.NoError(t, err)

	var decoded MediaCollection
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "MCU", decoded.Name)
	assert.Equal(t, "131292", decoded.ExternalIDs["tmdb"])
}
