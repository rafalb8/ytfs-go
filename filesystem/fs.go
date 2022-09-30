package filesystem

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// FS implements the ytfs file system.
type FS struct {
	Files      map[string]*File
	DIREntries []fuse.Dirent
}

func (fs *FS) Root() (fs.Node, error) {
	return &Dir{fs}, nil
}
