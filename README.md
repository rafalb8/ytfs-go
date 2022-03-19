# YTFS-Go

Simple fuse file system which enables you to play YouTube playlist as files.
Based on FUSE, written in Go.

Only audio for now.

# Installation

You can install ytfs-go with go:
    
    $ go install github.com/rafalb8/ytfs-go@latest

Default install location: `~/go/bin`

# Usage

Mount in empty directory

    $ mkdir youtube
    $ ~/go/bin/ytfs-go youtube "https://www.youtube.com/playlist?list=PLMC9KNkIncKtPzgY-5rmhvj7fax8fdxoj"


# Dependencies

- FUSE - [bazil/fuse](https://github.com/bazil/fuse)
- youtube - [kkdai/youtube](https://github.com/kkdai/youtube)
- FFmpeg - If using `-a wav` or `-a mp3` flag