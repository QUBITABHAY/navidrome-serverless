package r2

import (
	"testing"
)

func TestExtractMetadata_PathFallback(t *testing.T) {
	key := "The Beatles/Abbey Road/01 - Come Together.mp3"
	title, artist, album, albumArtist, year, trackNum, discNum, genre := extractMetadata(key, nil)

	if artist != "The Beatles" {
		t.Errorf("expected artist 'The Beatles', got %q", artist)
	}
	if album != "Abbey Road" {
		t.Errorf("expected album 'Abbey Road', got %q", album)
	}
	if title != "Come Together" {
		t.Errorf("expected title 'Come Together', got %q", title)
	}
	if trackNum != 1 {
		t.Errorf("expected trackNum 1, got %d", trackNum)
	}
	if albumArtist != "The Beatles" {
		t.Errorf("expected albumArtist 'The Beatles', got %q", albumArtist)
	}
	_ = year
	_ = discNum
	_ = genre
}

func TestSplitTrackPrefix(t *testing.T) {
	tests := []struct {
		input     string
		wantNum   int
		wantTitle string
		wantOk    bool
	}{
		{"01 - Yesterday", 1, "Yesterday", true},
		{"12. Something", 12, "Something", true},
		{"03_Octopus Garden", 3, "Octopus Garden", true},
		{"NoPrefixTrack", 0, "NoPrefixTrack", false},
	}

	for _, tt := range tests {
		gotNum, gotTitle, gotOk := splitTrackPrefix(tt.input)
		if gotOk != tt.wantOk || gotNum != tt.wantNum || gotTitle != tt.wantTitle {
			t.Errorf("splitTrackPrefix(%q) = (%d, %q, %v), want (%d, %q, %v)",
				tt.input, gotNum, gotTitle, gotOk, tt.wantNum, tt.wantTitle, tt.wantOk)
		}
	}
}
