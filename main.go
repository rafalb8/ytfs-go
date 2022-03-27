package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
	"github.com/kkdai/youtube/v2"
)

var SelectedAudioFormat string

type AudioType struct {
	Mime      string
	Extension string
}

var AudioFormatMap = map[string]AudioType{
	"aac":  {"mp4a", "m4a"},
	"opus": {"opus", "webm"},
	"wav":  {"wav", "wav"},
	"mp3":  {"mp3", "mp3"},
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT PLAYLIST_URL\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.StringVar(&SelectedAudioFormat, "a", "aac", "Set audio format (aac, opus, wav, mp3)")
	flag.Usage = usage
	flag.Parse()

	defFormat, exists := AudioFormatMap[SelectedAudioFormat]

	if flag.NArg() != 2 || !exists {
		usage()
		os.Exit(2)
	}

	mountpoint := flag.Arg(0)
	playlistURL := flag.Arg(1)

	// Get Playlist videos
	playlist, err := GetPlaylist(playlistURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	bytesPerSec := 16000

	if SelectedAudioFormat == "wav" {
		bytesPerSec = 176400
	}

	files := map[string]*File{}
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
		files[dirEntry.Name] = &File{
			Title:         title,
			Inode:         dirEntry.Inode,
			PlaylistEntry: entry,
			Size:          uint64(int(entry.Duration.Seconds()+1) * bytesPerSec),
		}
		fmt.Printf("%d. %s\n", i+1, title)
	}

	fs := &FS{
		Files:      files,
		DIREntries: dirEntries,
	}

	run(mountpoint, fs)
}

func run(mountpoint string, filesys *FS) {
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

// FS implements the ytfs file system.
type FS struct {
	Files      map[string]*File
	DIREntries []fuse.Dirent
}

func (fs *FS) Root() (fs.Node, error) {
	return &Dir{fs}, nil
}

// Dir implements both Node and Handle for the root directory.
type Dir struct {
	FS *FS
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0o555
	return nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if file, exists := d.FS.Files[name]; exists {
		return file, nil
	}
	return nil, syscall.ENOENT
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	return d.FS.DIREntries, nil
}

// File implements both Node and Handle for the yt file.
type File struct {
	Title         string
	Inode         uint64
	Size          uint64
	Data          []byte
	PlaylistEntry *youtube.PlaylistEntry
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = f.Inode
	a.Mode = 0o444
	a.Size = f.Size
	return nil
}

func (f *File) ReadAll(ctx context.Context) (data []byte, err error) {
	if len(f.Data) <= 0 {
		err = f.CacheData(ctx)
	}

	return f.Data, err
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	if len(f.Data) <= 0 {
		err := f.CacheData(ctx)
		if err != nil {
			return err
		}
	}

	end := req.Offset + int64(req.Size)
	if end > int64(f.Size) {
		end = int64(f.Size)
	}

	resp.Data = f.Data[req.Offset:end]
	return nil
}

func (f *File) CacheData(ctx context.Context) error {
	video, err := client.VideoFromPlaylistEntryContext(ctx, f.PlaylistEntry)
	if err != nil {
		return err
	}

	reader, err := GetAudioReader(video)
	if err != nil {
		return err
	}

	f.Data = make([]byte, f.Size)

	offset := 0
	for offset < int(f.Size) {
		n, err := reader.Read(f.Data[offset:])
		offset += n

		switch err {
		case io.EOF:
			offset = int(f.Size)
		case nil:
		default:
			return err
		}

		fmt.Printf("\r%s - %.2f%%", f.Title, float32(offset)/float32(f.Size)*100)

	}
	fmt.Println()
	return nil
}
