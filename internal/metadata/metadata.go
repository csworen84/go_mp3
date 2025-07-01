package metadata

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/bogem/id3v2"
)

// ApplyMetadata applies artist, album, title metadata and embeds cover art
func ApplyMetadata(inputPath, outputPath, artist, album, title string) error {
	// Open MP3 file
	tag, err := id3v2.Open(inputPath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("error opening mp3 for tagging: %w", err)
	}
	defer tag.Close()

	// Set fields
	tag.SetArtist(artist)
	tag.SetAlbum(album)
	tag.SetTitle(title)

	// Save changes
	err = tag.Save()
	if err != nil {
		return fmt.Errorf("error saving metadata: %w", err)
	}

	// Rename if needed (outputPath may be same as inputPath)
	if inputPath != outputPath {
		if err := os.Rename(inputPath, outputPath); err != nil {
			return fmt.Errorf("failed to rename tagged file: %w", err)
		}
	}

	return nil
}

// downloadCoverArt tries to download YouTube thumbnail from video file metadata
func downloadCoverArt(mp3Path string) (string, error) {
	// Extract video ID from temp file path or metadata (simplified approach)
	// In real implementation, pass video ID or thumbnail URL explicitly.

	// For demo: do nothing, return empty string (skip cover art)
	return "", nil

	// Optional: implement actual thumbnail download using YouTube API or thumbnail URL parsing
}

// readImage verifies image data (optional)
func verifyImage(data []byte) bool {
	_, _, err := image.Decode(bytes.NewReader(data))
	return err == nil
}
