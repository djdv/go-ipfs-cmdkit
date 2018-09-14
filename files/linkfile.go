package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Symlink struct {
	name         string
	path         string
	Target       string
	stat         os.FileInfo
	resolveDepth int

	reader io.Reader
}

func NewLinkFile(name, path, target string, stat os.FileInfo, resolveDepth int) File {
	return &Symlink{
		name:         name,
		path:         path,
		Target:       target,
		stat:         stat,
		reader:       strings.NewReader(target),
		resolveDepth: resolveDepth,
	}
}

func (lf *Symlink) IsDirectory() bool {
	return false
}

func (lf *Symlink) NextFile() (File, error) {
	return nil, io.EOF
}

func (f *Symlink) FileName() string {
	return f.name
}

func (f *Symlink) Close() error {
	return nil
}

func (f *Symlink) FullPath() string {
	return f.path
}

func (f *Symlink) Read(b []byte) (int, error) {
	return f.reader.Read(b)
}

func (f *Symlink) Size() (int64, error) {
	// 0 default (no deref), 1 -H (deref commandline), N -L=N (deref N levels deep)
	if f.resolveDepth == 0 {
		return int64(len(f.Target)), nil
	}

	var size int64
	err := filepath.Walk(f.Target, recursiveLinkWalk(&size, f.resolveDepth))
	return size, err
}

func recursiveLinkWalk(size *int64, depthLimit int) filepath.WalkFunc {
	var depth int
	return func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		depth++
		defer func() { depth-- }()

		if fi.Mode().IsRegular() {
			*size += fi.Size()
			return nil
		}

		//conditionally recurs symlinks
		if fi.Mode()&os.ModeSymlink != 0 {
			if depth < depthLimit {
				target, err := filepath.EvalSymlinks(p)
				if err != nil {
					return err
				}

				err = filepath.Walk(target, recursiveLinkWalk(size, 1))
				if err != nil {
					return err
				}
				return nil
			}

			// size the link itself
			lStat, err := os.Lstat(p)
			if err != nil {
				return err
			}
			fmt.Printf("%q:%d", lStat.Name(), lStat.Size())
			*size += lStat.Size()
			return nil
		}
		return nil
	}
}
