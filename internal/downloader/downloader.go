package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"go_mp3/internal/metadata"
	"go_mp3/internal/models"
)

type YTEntry struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type PlaylistJSON struct {
	Entries []YTEntry `json:"entries"`
}

func sanitizeFilename(name string) string {
	re := regexp.MustCompile(`[\\/:*?"<>|]`)
	return re.ReplaceAllString(name, "_")
}

func removeTags(title string) string {
	re := regexp.MustCompile(`[\[\(\{<][^\]\)\}>]*(?i:official|lyric[s]?|audio|album|visuali[sz]er|bonus)[^\[\]\)\}>]*[\]\)\}>]|[\[\(\{<][^\]\)\}>]*\bHD\b[^\[\]\)\}>]*[\]\)\}>]
`)
	title = re.ReplaceAllString(title, "")
	title = strings.TrimSpace(title)
	return title
}

func sanitizeTitle(title, artist string) string {
	title = strings.TrimSpace(title)
	artist = strings.TrimSpace(artist)

	pattern := fmt.Sprintf(`(?i)^%s\s*[-–—|:]*\s*`, regexp.QuoteMeta(artist))
	re := regexp.MustCompile(pattern)
	title = re.ReplaceAllString(title, "")

	title = removeTags(title)
	title = strings.TrimSpace(title)
	title = strings.Trim(title, "-–—| ")

	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")

	return sanitizeFilename(title)
}

func isAlternateVersion(title string) bool {
	title = strings.ToLower(strings.TrimSpace(title))

	// Regex: look for (keyword) or [keyword] or even phrases inside () or []
	pattern := regexp.MustCompile(`[\[\(][^)\]]*(acoustic|live|remix|mix|edit|instrumental|version)[^)\]]*[\]\)]`)

	return pattern.MatchString(title)
}

func getOutputDir(artist string) string { return "music/" + sanitizeFilename(artist) }

// getCleanVideoEntries gets video IDs and titles from playlist JSON, filtering out unwanted tags and alternates.
func getCleanVideoEntries(playlistURL string) ([]YTEntry, error) {
	cmd := exec.Command("yt-dlp", "-J", playlistURL)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var pl PlaylistJSON
	if err := json.Unmarshal(output, &pl); err != nil {
		return nil, err
	}

	var filtered []YTEntry
	for _, entry := range pl.Entries {
		titleLower := strings.ToLower(entry.Title)
		if isAlternateVersion(titleLower) {
			continue
		} // skip alternate versions
		filtered = append(filtered, entry)
	}
	return filtered, nil
}

// downloadTrack downloads a single track with retries and logs.
func downloadTrack(url, outputDir string, maxRetries int) error {
	var stderr bytes.Buffer

	cmd := exec.Command("yt-dlp",
		"-x", "--audio-format", "mp3",
		"--sleep-interval", "3",
		"--max-sleep-interval", "7",
		"--retries", "10",
		"--fragment-retries", "10",
		"-o", filepath.Join(outputDir, "%(title)s.%(ext)s"),
		url)
	cmd.Stderr = &stderr

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := cmd.Run()
		if err == nil {
			return nil
		}
		log.Printf("yt-dlp failed (try %d/%d) for %s: %v\nDetails: %s", attempt, maxRetries, url, err, stderr.String())
		if attempt < maxRetries {
			sleepDuration := time.Duration(attempt*5) * time.Second
			log.Printf("Retrying in %v...", sleepDuration)
			time.Sleep(sleepDuration)
			stderr.Reset()
		} else {
			return fmt.Errorf("max retries reached for %s", url)
		}
	}
	return nil
}

func ProcessPlaylist(pl models.Playlist) {
	artist := pl.Artist
	outputDir := getOutputDir(artist)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	fmt.Printf("Processing playlist: %s (Artist: %s, Album: %s)\n", pl.URL, artist, pl.Album)

	videoEntries, err := getCleanVideoEntries(pl.URL)
	if err != nil {
		log.Fatalf("Failed to retrieve playlist entries: %v", err)
	}

	maxRetries := 3
	concurrency := 5
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	// Track titles downloaded for metadata processing
	downloadedTitles := make([]string, 0, len(videoEntries))
	mu := sync.Mutex{} // to guard downloadedTitles

	for _, entry := range videoEntries {
		cleanTitle := sanitizeTitle(entry.Title, artist)
		if isAlternateVersion(cleanTitle) {
			fmt.Printf("Skipping alternate version: %s\n", cleanTitle)
			continue
		}

		trackURL := "https://www.youtube.com/watch?v=" + entry.ID

		sem <- struct{}{}
		wg.Add(1)

		go func(url, title string) {
			defer wg.Done()
			defer func() { <-sem }()

			err := downloadTrack(url, outputDir, maxRetries)
			if err != nil {
				log.Printf("Failed to download %s: %v", url, err)
				return
			}

			// Rename file after download
			originalFilename := filepath.Join(outputDir, entry.Title+".mp3")
			cleanFilename := filepath.Join(outputDir, title+".mp3")

			// Sometimes yt-dlp doesn't produce exactly the title as expected, try to find matching file:
			// We'll just scan outputDir for any .mp3 whose name contains a cleaned version of the raw title (simplified)
			if _, err := os.Stat(originalFilename); os.IsNotExist(err) {
				// fallback: search directory for a close match
				files, _ := os.ReadDir(outputDir)
				found := false
				for _, f := range files {
					if strings.HasSuffix(strings.ToLower(f.Name()), ".mp3") &&
						strings.Contains(strings.ToLower(f.Name()), strings.ToLower(title)) {
						originalFilename = filepath.Join(outputDir, f.Name())
						found = true
						break
					}
				}
				if !found {
					log.Printf("Could not find downloaded file to rename for track %s", title)
					return
				}
			}

			if originalFilename != cleanFilename {
				if err := os.Rename(originalFilename, cleanFilename); err != nil {
					log.Printf("Failed to rename %s to %s: %v", originalFilename, cleanFilename, err)
					return
				}
			}

			mu.Lock()
			downloadedTitles = append(downloadedTitles, cleanFilename)
			mu.Unlock()

			fmt.Printf("Downloaded and renamed: %s\n", cleanFilename)
		}(trackURL, cleanTitle)
	}

	wg.Wait()

	// Apply metadata after all downloads finish
	for _, file := range downloadedTitles {
		base := filepath.Base(file)
		title := strings.TrimSuffix(base, ".mp3")
		err := metadata.ApplyMetadata(file, file, artist, pl.Album, title)
		if err != nil {
			log.Printf("Failed to apply metadata to %s: %v", file, err)
		} else {
			fmt.Printf("Tagged metadata for: %s\n", file)
		}
	}

	// Random delay 10-20 seconds between playlists
	delay := 10 + rand.Intn(11) // 10 to 20 seconds
	fmt.Printf("Sleeping for %d seconds between playlists...\n", delay)
	time.Sleep(time.Duration(delay) * time.Second)
}
