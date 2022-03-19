package main

import (
	"io"

	"github.com/kkdai/youtube/v2"
)

var client = youtube.Client{}

func GetPlaylist(url string) *youtube.Playlist {
	playlist, err := client.GetPlaylist(url)
	if err != nil {
		panic(err)
	}

	return playlist
}

func GetAudioReader(video *youtube.Video) io.ReadCloser {
	var audioFormat *youtube.Format
	audioFormats := video.Formats.Type("audio")
	if len(audioFormats) > 0 {
		audioFormats.Sort()
		audioFormat = &audioFormats[0]
	}

	if audioFormat == nil {
		panic("no audio format found after filtering")
	}

	reader, _, err := client.GetStream(video, audioFormat)
	if err != nil {
		panic(err)
	}

	return reader
}
