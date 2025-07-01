package models

// Playlist represents a single playlist or mix entry from the config file.
type Playlist struct {
	URL                string `json:"url"`
	Artist             string `json:"artist,omitempty"`
	Album              string `json:"album,omitempty"`
	MixName            string `json:"mix_name,omitempty"`
	IncludeAltVersions bool   `json:"include_alt_versions"`
}

// Song represents an individual song/video extracted from a playlist.
type Song struct {
	Title     string
	Artist    string
	Album     string
	URL       string
	FilePath  string
	Thumbnail string
}
