package gitty

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

func fakeNew(fakeRepo Repository) Gitty {
	return &Git{repo: fakeRepo}
}

func TestNewRepo(t *testing.T) {
	r := New()
	assert.NotNil(t, r)
}

func TestStatus(t *testing.T) {
	// Discard output during tests.
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	tests := []struct {
		name     string
		repo     Repository
		expected error
	}{
		{
			name:     "success status",
			repo:     fakeRepository(&mockSuccess{}),
			expected: nil,
		},
		{
			name:     "error status",
			repo:     fakeRepository(&mockError{}),
			expected: fmt.Errorf("failed to check status: %v", errMockRateLimit),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := fakeNew(test.repo)
			err := g.Status(context.Background())
			assert.Equal(t, test.expected, err)
		})
	}
}

func TestDownload(t *testing.T) {
	// Discard output during tests.
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	tests := []struct {
		name     string
		repo     Repository
		url      string
		paths    []string
		expected error
	}{
		{
			name:     "success download",
			repo:     fakeRepository(&mockSuccess{}),
			url:      "https://github.com/owner/repo/tree/branch/directory",
			paths:    []string{contentsData[0].GetPath(), contentsData[1].GetPath()},
			expected: nil,
		},
		{
			name:     "error extract",
			repo:     fakeRepository(&mockSuccess{}),
			url:      gofakeit.URL(),
			expected: ErrNotValidURL,
		},
		{
			name:     "error content",
			repo:     fakeRepository(&mockError{}),
			url:      "https://github.com/owner/repo/tree/branch/directory",
			expected: fmt.Errorf("failed to download: %v", errMockContents),
		},
		{
			name:     "error download file",
			repo:     fakeRepository(&mockError{}),
			url:      fmt.Sprintf("https://github.com/owner/repo/tree/branch/%s", testDownloadFail),
			expected: fmt.Errorf("failed to download: %v", ErrInvalidPathURL),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expected == nil {
				t.Cleanup(func() {
					deleteTestFiles(t, test.paths...)
				})
			}
			g := fakeNew(test.repo)
			err := g.Download(context.Background(), test.url)
			assert.Equal(t, test.expected, err)
		})
	}
}

func TestAuth(t *testing.T) {
	// Discard output during tests.
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	tests := []struct {
		name     string
		repo     Repository
		expected error
	}{
		{
			name:     "success status",
			repo:     fakeRepository(&mockSuccess{}),
			expected: nil,
		},
		{
			name:     "error status",
			repo:     fakeRepository(&mockError{}),
			expected: fmt.Errorf("failed to check auth: %v", errMockGetUser),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := fakeNew(test.repo)
			err := g.Auth(context.Background())
			assert.Equal(t, test.expected, err)
		})
	}
}
