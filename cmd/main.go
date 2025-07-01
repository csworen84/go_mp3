package main

import (
	"log"
	"os"

	"go_mp3/internal/downloader"
	"go_mp3/internal/utils"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go_mp3 <path to playlist.txt>")
	}
	playlistPath := os.Args[1]

	playlists, err := utils.LoadPlaylistsFromTxt(playlistPath)
	if err != nil {
		log.Fatalf("Failed to load playlist file: %v", err)
	}

	for _, pl := range playlists {
		downloader.ProcessPlaylist(pl)

		utils.RandomDelay(10, 20)
	}
}
