# go_mp3

A command-line tool to download YouTube playlists or individual albums as MP3 files, organized by artist and album, with automatic metadata tagging.

## Features

- Download entire playlists or a single album from YouTube as MP3.
- Organizes downloads into `music/<Artist>/<Album>.mp3`.
- Automatically tags MP3 files with artist, album, and title metadata.
- Skips alternate versions (live, acoustic, remix, etc.) by default.
- Attempts to ingore/remove random tagging from YouTube titles (work in progress).
- Supports random delays between playlist downloads to avoid rate limits.

## Requirements

- [Go](https://golang.org/dl/) 1.24+
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) installed and in your PATH
- [ffmpeg](https://ffmpeg.org/) installed and in your PATH

## Installation

Clone the repository and build:

```sh
git clone https://github.com/yourusername/go_mp3.git
cd go_mp3
go build -o go_mp3 ./cmd
```

## Usage

### Downloading a List of Playlists

Prepare a playlist text file (see [`playlists/`](playlists/) for examples). Each line should be:

```
Artist - Album|YouTubePlaylistURL
```

Example (`playlists/disturbed.txt`):

```
Disturbed - Believe|https://www.youtube.com/watch?v=SbS2iOlwTTQ&list=...
Disturbed - Indestructible|https://www.youtube.com/watch?v=kjRWG0tKD4A&list=...
```

`Artist - Album|` syntax is optional, but issues may arise if omitted...

Run:

```sh
./go_mp3 playlists/disturbed.txt
```

All albums in the file will be downloaded and tagged.

### Output

MP3 files will be saved in the `music/<Artist>/` directory, named after the song title, and tagged with artist, album, and title if using playlist with artist/album syntax.
If running without artist/album syntax (or doing individuals) songs will still go to the `music` folder by cleaned title only (no metadata tagging).

## Notes

- Alternate versions (live, acoustic, remix, etc.) are skipped automatically.
- The tool applies a random delay between playlists to avoid triggering YouTube rate limits.

## License

MIT License