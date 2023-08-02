package fsutil

import (
	"context"
	gofs "io/fs"
	"os"
	"testing"

	"github.com/containerd/continuity/fs/fstest"
	"github.com/stretchr/testify/require"
	"github.com/tonistiigi/fsutil/types"
)

func TestWalk(t *testing.T) {
	tmpDir := t.TempDir()

	apply := fstest.Apply(
		fstest.CreateDir("dir", 0700),
		fstest.CreateFile("dir/foo", []byte("contents"), 0600),
	)
	require.NoError(t, apply.Apply(tmpDir))

	f := NewFS(tmpDir)
	paths := []string{}
	files := []gofs.DirEntry{}
	err := f.Walk(context.TODO(), "", func(path string, entry gofs.DirEntry, err error) error {
		require.NoError(t, err)
		paths = append(paths, path)
		files = append(files, entry)
		return nil
	})
	require.NoError(t, err)

	require.Equal(t, paths, []string{"dir", "dir/foo"})
	require.Len(t, files, 2)
	require.Equal(t, "dir", files[0].Name())
	require.Equal(t, "foo", files[1].Name())

	fis := []gofs.FileInfo{}
	for _, f := range files {
		fi, err := f.Info()
		require.NoError(t, err)
		fis = append(fis, fi)
	}
	require.Equal(t, "dir", fis[0].Name())
	require.Equal(t, "foo", fis[1].Name())

	require.Equal(t, len("contents"), int(fis[1].Size()))

	require.Equal(t, "dir", fis[0].(*StatInfo).Path)
	require.Equal(t, "dir/foo", fis[1].(*StatInfo).Path)
}

func TestWalkDir(t *testing.T) {
	tmpDir := t.TempDir()
	apply := fstest.Apply(
		fstest.CreateDir("dir", 0700),
		fstest.CreateFile("dir/foo", []byte("contents"), 0600),
	)
	require.NoError(t, apply.Apply(tmpDir))

	tmpDir2 := t.TempDir()
	apply = fstest.Apply(
		fstest.CreateDir("dir", 0700),
		fstest.CreateFile("dir/bar", []byte("contents2"), 0600),
	)
	require.NoError(t, apply.Apply(tmpDir2))

	f, err := SubDirFS([]Dir{
		{
			Stat: types.Stat{
				Mode: uint32(os.ModeDir | 0755),
				Path: "1",
			},
			FS: NewFS(tmpDir),
		},
		{
			Stat: types.Stat{
				Mode: uint32(os.ModeDir | 0755),
				Path: "2",
			},
			FS: NewFS(tmpDir2),
		},
	})
	require.NoError(t, err)
	paths := []string{}
	files := []gofs.DirEntry{}
	err = f.Walk(context.TODO(), "", func(path string, entry gofs.DirEntry, err error) error {
		require.NoError(t, err)
		paths = append(paths, path)
		files = append(files, entry)
		return nil
	})
	require.NoError(t, err)

	require.Equal(t, paths, []string{"1", "1/dir", "1/dir/foo", "2", "2/dir", "2/dir/bar"})
	require.Equal(t, "1", files[0].Name())
	require.Equal(t, "dir", files[1].Name())
	require.Equal(t, "foo", files[2].Name())
}
