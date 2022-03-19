package main

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/kkdai/youtube/v2"
)

var client = youtube.Client{}

func GetPlaylist(url string) (*youtube.Playlist, error) {
	playlist, err := client.GetPlaylist(url)
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

func GetAudioReader(video *youtube.Video) (io.ReadCloser, error) {
	var audioFormat *youtube.Format
	audioFormats := video.Formats.Type("audio")
	if len(audioFormats) > 0 {
		audioFormats.Sort()
		audioFormat = &audioFormats[0]

		defFormat := AudioFormatMap[AudioFormat]
		for _, f := range audioFormats {
			if strings.Contains(f.MimeType, defFormat.Mime) {
				audioFormat = &f
				break
			}
		}
	}

	if audioFormat == nil {
		return nil, fmt.Errorf("no audio format found after filtering")
	}

	reader, _, err := client.GetStream(video, audioFormat)
	if err != nil {
		return nil, err
	}

	// Skip format for aac and opus
	switch AudioFormat {
	case "aac", "opus":
		return reader, nil
	}

	// Convert other formats
	ffmpeg := exec.Command("ffmpeg", "-i", "pipe:", "-f", AudioFormat, "pipe:")
	ffmpeg.Stdin = reader
	reader, err = ffmpeg.StdoutPipe()

	return reader, ffmpeg.Start()
}
