package utils

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"go_mp3/internal/models"
)

// LoadPlaylistsFromTxt parses a playlist.txt file.
// Each line: "Artist - Album|PlaylistURL"
func LoadPlaylistsFromTxt(path string) ([]models.Playlist, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var playlists []models.Playlist
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // skip empty or comments
		}

		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line (missing '|'): %s", line)
		}

		desc := strings.TrimSpace(parts[0])
		url := strings.TrimSpace(parts[1])

		artist := ""
		album := ""
		// Parse "Artist - Album" from desc if possible
		descParts := strings.SplitN(desc, " - ", 2)
		if len(descParts) == 2 {
			artist = descParts[0]
			album = descParts[1]
		} else {
			artist = desc // fallback if no album
		}

		playlists = append(playlists, models.Playlist{
			URL:    url,
			Artist: artist,
			Album:  album,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return playlists, nil
}

func RandomDelay(minSec, maxSec int) {
	rand.Seed(time.Now().UnixNano())
	delay := rand.Intn(maxSec-minSec+1) + minSec
	fmt.Printf("Sleeping for %d seconds between playlists...\n", delay)
	time.Sleep(time.Duration(delay) * time.Second)
}
