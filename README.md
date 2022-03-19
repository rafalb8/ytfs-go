# YTFS-Go

Simple fuse filesystem which mount youtube playlist.

Only audio for now.

## Usage

Building:

    go build

Running:

    ./ytfs-go MOUNTPOINT PLAYLIST_URL

Example:

    ./ytfs-go "$HOME/yt-music" "https://www.youtube.com/playlist?list=PLMC9KNkIncKtPzgY-5rmhvj7fax8fdxoj"



