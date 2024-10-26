package gitty

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

func TestGetGitHubRepo(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expected    string
		expectedErr error
	}{
		{
			name:        "valid url with https://github.com/",
			url:         "https://github.com/owner/repo/tree/branch/directory",
			expected:    "owner/repo/tree/branch/directory",
			expectedErr: nil,
		},
		{
			name:        "valid url with github.com/",
			url:         "github.com/owner/repo/tree/branch/directory1/directory2",
			expected:    "owner/repo/tree/branch/directory1/directory2",
			expectedErr: nil,
		},
		{
			name:        "valid file url with github.com/",
			url:         "github.com/owner/repo/tree/branch/directory1/directory2/file.txt",
			expected:    "owner/repo/tree/branch/directory1/directory2/file.txt",
			expectedErr: nil,
		},
		{
			name:        "invalid https url",
			url:         "https://gitlab.com/owner/repo/tree/branch/directory",
			expected:    "",
			expectedErr: ErrNotValidURL,
		},
		{
			name:        "invalid url",
			url:         "gitlab.com/owner/repo/tree/branch/directory",
			expected:    "",
			expectedErr: ErrNotValidURL,
		},
		{
			name:        "invalid https url format",
			url:         "https://github.com/owner/repo",
			expected:    "",
			expectedErr: ErrNotValidFormat,
		},
		{
			name:        "invalid url format",
			url:         "github.com/owner/repo",
			expected:    "",
			expectedErr: ErrNotValidFormat,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, err := getGitHubRepo(test.url)
			assert.Equal(t, test.expected, repo)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectedErr error
	}{
		{
			name:        "valid format 1",
			input:       "owner/repo/tree/branch/directory",
			expected:    "owner/repo/tree/branch/directory",
			expectedErr: nil,
		},
		{
			name:        "valid format 2",
			input:       "owner/repo/tree/branch/directory1/directory2",
			expected:    "owner/repo/tree/branch/directory1/directory2",
			expectedErr: nil,
		},
		{
			name:        "valid format 3",
			input:       "owner/repo/tree/branch/directory1/directory2/file.txt",
			expected:    "owner/repo/tree/branch/directory1/directory2/file.txt",
			expectedErr: nil,
		},
		{
			name:        "invalid format 1",
			input:       "owner/repo/directory",
			expected:    "",
			expectedErr: ErrNotValidFormat,
		},
		{
			name:        "invalid format 2",
			input:       "owner/repo.git",
			expected:    "",
			expectedErr: ErrNotValidFormat,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path, err := validate(test.input)
			assert.Equal(t, test.expected, path)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}

type errReader int

var errMockReadAll = errors.New("mock readall body error")

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errMockReadAll
}

func TestSaveFile(t *testing.T) {
	// Discard output during tests.
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	mkdirErr := func() error {
		switch runtime.GOOS {
		case "windows":
			return syscall.Errno(123)
		case "linux":
			return syscall.Errno(36)
		case "darwin":
			return syscall.Errno(63)
		default:
			// Otherwise, raise an error.
			return fmt.Errorf("%v is currently unsupported to test mdkirall", runtime.GOOS)
		}
	}
	tests := []struct {
		name     string
		base     string
		path     string
		body     io.Reader
		expected error
	}{
		{
			name:     "save file successfully",
			base:     "tmp",
			path:     fmt.Sprintf("tmp/%s.txt", gofakeit.LoremIpsumWord()),
			body:     bytes.NewBufferString("test data"),
			expected: nil,
		},
		{
			name:     "error reading body",
			base:     "tmp1",
			path:     "tmp1/file1.txt",
			body:     errReader(0),
			expected: errMockReadAll,
		},
		{
			name:     "invalid base path",
			base:     "/nonexistent/base",
			path:     "path/to/dir/file2.txt",
			body:     bytes.NewBufferString("test data"),
			expected: fmt.Errorf("Rel: can't make %s relative to %s", "path/to/dir/file2.txt", "/nonexistent/base"),
		},
		{
			name:     "error creating directory",
			base:     strings.Repeat("a", 256),
			path:     strings.Repeat("a", 256) + "/nofile.txt",
			body:     bytes.NewBufferString("test data"),
			expected: &fs.PathError{Op: "mkdir", Path: strings.Repeat("a", 256), Err: mkdirErr()},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := saveFile(test.base, test.path, test.body)
			assert.Equal(t, test.expected, err)
			if test.expected == nil {
				err = os.Remove(test.path)
				assert.Nil(t, err)
				err = os.RemoveAll(test.base)
				assert.Nil(t, err)
			}
		})
	}
}

func TestExactPath(t *testing.T) {
	tests := []struct {
		name        string
		base        string
		path        string
		expected    string
		expectedErr error
	}{
		{
			name:        "same directory",
			base:        "path/to/dir",
			path:        "path/to/dir/file.txt",
			expected:    "dir/file.txt",
			expectedErr: nil,
		},
		{
			name:        "subdirectory",
			base:        "path/to/dir",
			path:        "path/to/dir/sub/file.txt",
			expected:    "dir/sub/file.txt",
			expectedErr: nil,
		},
		{
			name:        "different base directory",
			base:        "path/to/base",
			path:        "path/to/dir/file.txt",
			expected:    "dir/file.txt",
			expectedErr: nil,
		},
		{
			name:        "invalid base path",
			base:        "/nonexistent/base",
			path:        "path/to/dir/file.txt",
			expected:    "",
			expectedErr: fmt.Errorf("Rel: can't make %s relative to %s", "path/to/dir/file.txt", "/nonexistent/base"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := exactPath(test.base, test.path)
			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expected, filepath.ToSlash(result))
		})
	}
}
