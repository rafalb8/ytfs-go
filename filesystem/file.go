package filesystem

import (
	"context"
	"fmt"
	"io"

	"bazil.org/fuse"
	"github.com/kkdai/youtube/v2"
	"github.com/rafalb8/ytfs-go/client"
)

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
	video, err := client.Client.VideoFromPlaylistEntryContext(ctx, f.PlaylistEntry)
	if err != nil {
		return err
	}

	reader, err := client.GetAudioReader(video)
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
