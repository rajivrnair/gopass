package fsutil

import (
	"crypto/rand"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanFilename(t *testing.T) {
	t.Parallel()

	m := map[string]string{
		`"§$%&aÜÄ*&b%§"'Ä"c%$"'"`: "a____b______c",
	}
	for k, v := range m {
		out := CleanFilename(k)
		t.Logf("%s -> %s / %s", k, v, out)

		assert.Equal(t, v, out)
	}
}

func TestCleanPath(t *testing.T) { //nolint:paralleltest
	tempdir, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	m := map[string]string{
		".":                                 "",
		"/home/user/../bob/.password-store": "/home/bob/.password-store",
		"/home/user//.password-store":       "/home/user/.password-store",
		tempdir + "/foo.gpg":                tempdir + "/foo.gpg",
	}

	usr, err := user.Current()
	if err == nil {
		hd := usr.HomeDir
		if gph := os.Getenv("GOPASS_HOMEDIR"); gph != "" {
			hd = gph
		}

		m["~/.password-store"] = hd + "/.password-store"
	}

	for in, out := range m {
		got := CleanPath(in)

		// filepath.Abs turns /home/bob into C:\home\bob on Windows
		absOut, err := filepath.Abs(out)
		assert.NoError(t, err)
		assert.Equal(t, absOut, got)
	}
}

func TestIsDir(t *testing.T) {
	t.Parallel()

	tempdir, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	fn := filepath.Join(tempdir, "foo")
	assert.NoError(t, os.WriteFile(fn, []byte("bar"), 0o644))
	assert.Equal(t, true, IsDir(tempdir))
	assert.Equal(t, false, IsDir(fn))
	assert.Equal(t, false, IsDir(filepath.Join(tempdir, "non-existing")))
}

func TestIsFile(t *testing.T) {
	t.Parallel()

	tempdir, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	fn := filepath.Join(tempdir, "foo")
	assert.NoError(t, os.WriteFile(fn, []byte("bar"), 0o644))
	assert.Equal(t, false, IsFile(tempdir))
	assert.Equal(t, true, IsFile(fn))
}

func TestShred(t *testing.T) {
	t.Parallel()

	tempdir, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	fn := filepath.Join(tempdir, "file")
	// test successful shread
	fh, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0o644)
	assert.NoError(t, err)

	buf := make([]byte, 1024)
	for i := 0; i < 10*1024; i++ {
		_, _ = rand.Read(buf)
		_, _ = fh.Write(buf)
	}

	require.NoError(t, fh.Close())
	assert.NoError(t, Shred(fn, 8))
	assert.Equal(t, false, IsFile(fn))

	// test failed
	fh, err = os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0o400)
	assert.NoError(t, err)

	buf = make([]byte, 1024)
	for i := 0; i < 10*1024; i++ {
		_, _ = rand.Read(buf)
		_, _ = fh.Write(buf)
	}

	require.NoError(t, fh.Close())
	assert.Error(t, Shred(fn, 8))
	assert.Equal(t, true, IsFile(fn))
}

func TestIsEmptyDir(t *testing.T) {
	t.Parallel()

	tempdir, err := os.MkdirTemp("", "gopass-")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	fn := filepath.Join(tempdir, "foo", "bar", "baz", "zab")
	require.NoError(t, os.MkdirAll(fn, 0o755))

	isEmpty, err := IsEmptyDir(tempdir)
	assert.NoError(t, err)
	assert.Equal(t, true, isEmpty)

	fn = filepath.Join(fn, ".config.yml")
	require.NoError(t, os.WriteFile(fn, []byte("foo"), 0o644))

	isEmpty, err = IsEmptyDir(tempdir)
	require.NoError(t, err)
	assert.Equal(t, false, isEmpty)
}
