package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
	"github.com/kkdai/youtube/v2"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT PLAYLIST_URL\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	// playlist := GetPlaylist("https://www.youtube.com/playlist?list=PLMC9KNkIncKtPzgY-5rmhvj7fax8fdxoj")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(2)
	}

	mountpoint := flag.Arg(0)
	playlistURL := flag.Arg(1)

	// Get Playlist videos
	playlist := GetPlaylist(playlistURL)

	files := map[string]*File{}
	dirEntries := []fuse.Dirent{}
	for i, entry := range playlist.Videos {
		// Create dir entry
		dirEntry := fuse.Dirent{
			Name:  entry.Title + ".mp4",
			Inode: uint64(i + 2),
			Type:  fuse.DT_Block,
		}
		dirEntries = append(dirEntries, dirEntry)

		files[dirEntry.Name] = &File{
			Inode:         dirEntry.Inode,
			PlaylistEntry: entry,
		}
		fmt.Println(dirEntry.Name)
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

	err = fs.Serve(c, filesys)
	if err != nil {
		log.Fatal(err)
	}
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
	Inode         uint64
	PlaylistEntry *youtube.PlaylistEntry
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = f.Inode
	a.Mode = 0o444
	a.Size = uint64(f.PlaylistEntry.Duration.Seconds() * 16000)
	return nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	video, err := client.VideoFromPlaylistEntryContext(ctx, f.PlaylistEntry)
	if err != nil {
		return nil, err
	}

	reader := GetAudioReader(video)
	return ioutil.ReadAll(reader)
}
