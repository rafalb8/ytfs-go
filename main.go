package main

import (
	"io"
	"os"

	"github.com/kkdai/youtube/v2"
)

func main() {
	client := youtube.Client{}

	playlist, err := client.GetPlaylist("https://www.youtube.com/playlist?list=PLMC9KNkIncKtPzgY-5rmhvj7fax8fdxoj")
	if err != nil {
		panic(err)
	}

	for _, entry := range playlist.Videos {
		video, err := client.VideoFromPlaylistEntry(entry)
		if err != nil {
			panic(err)
		}

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

		file, _ := os.Create(video.Title + ".mp4")

		io.Copy(file, reader)

		reader.Close()
		break
	}
}
