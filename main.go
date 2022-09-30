package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
	"github.com/rafalb8/ytfs-go/client"
	"github.com/rafalb8/ytfs-go/filesystem"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT PLAYLIST_URL\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.StringVar(&client.SelectedAudioFormat, "a", "aac", "Set audio format (aac, opus, wav, mp3)")
	flag.Usage = usage
	flag.Parse()

	defFormat, exists := client.AudioFormatMap[client.SelectedAudioFormat]

	if flag.NArg() != 2 || !exists {
		usage()
		os.Exit(2)
	}

	mountpoint := flag.Arg(0)
	playlistURL := flag.Arg(1)

	// Get Playlist videos
	playlist, err := client.GetPlaylist(playlistURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bytesPerSec := 16000

	if client.SelectedAudioFormat == "wav" {
		bytesPerSec = 176400
	}

	files := map[string]*filesystem.File{}
	dirEntries := []fuse.Dirent{}

	fmt.Println("Videos found:")
	for i, entry := range playlist.Videos {
		title := strings.ReplaceAll(entry.Title, "/", "|")

		// Create dir entry
		dirEntry := fuse.Dirent{
			Name:  title + "." + defFormat.Extension,
			Inode: uint64(i + 2),
			Type:  fuse.DT_Block,
		}
		dirEntries = append(dirEntries, dirEntry)

		// Create file entry
		files[dirEntry.Name] = &filesystem.File{
			Title:         title,
			Inode:         dirEntry.Inode,
			PlaylistEntry: entry,
			Size:          uint64(int(entry.Duration.Seconds()+1) * bytesPerSec),
		}
		fmt.Printf("%d. %s\n", i+1, title)
	}

	fs := &filesystem.FS{
		Files:      files,
		DIREntries: dirEntries,
	}

	run(mountpoint, fs)
}

func run(mountpoint string, filesys *filesystem.FS) {
	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("youtubefs"),
		fuse.Subtype("ytfs-go"),
	)

	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	fmt.Println("\nStarting filesystem")
	err = fs.Serve(c, filesys)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Stopping filesystem")
	// fuse.Unmount(mountpoint)
}
