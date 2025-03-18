package gitty

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fakeNew(fakeRepo Repository) Gitty {
	return &Git{repo: fakeRepo}
}

func TestNewRepo(t *testing.T) {
	t.Parallel()
	r := New()
	assert.NotNil(t, r)
}

func TestStatus(t *testing.T) {
	t.Parallel()
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
			expected: fmt.Errorf("failed to check status: %w", errMockRateLimit),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := fakeNew(test.repo)
			err := g.Status(context.Background())
			assert.Equal(t, test.expected, err)
		})
	}
}

func TestDownload(t *testing.T) {
	t.Parallel()

	fakeBase := fmt.Sprintf("%s_%d", gofakeit.LoremIpsumWord(), gofakeit.Int())
	fakeFirstPath := fmt.Sprintf("%s/%s_%d.txt", fakeBase, gofakeit.LoremIpsumWord(), gofakeit.Int())
	fakeSecondPath := fmt.Sprintf("%s/%s_%d.txt", fakeBase, gofakeit.LoremIpsumWord(), gofakeit.Int())
	t.Cleanup(func() {
		err := os.RemoveAll(fakeBase)
		require.NoError(t, err)
	})
	ctxfakePath := func() context.Context {
		return context.WithValue(context.Background(), pathKey, contentsData(fakeFirstPath, fakeSecondPath))
	}

	tests := []struct {
		name     string
		repo     Repository
		ctx      context.Context
		url      string
		expected error
	}{
		{
			name:     "success download",
			repo:     fakeRepository(&mockSuccess{}),
			ctx:      ctxfakePath(),
			url:      "https://github.com/owner/repo/tree/branch/directory",
			expected: nil,
		},
		{
			name:     "error extract",
			repo:     fakeRepository(&mockSuccess{}),
			ctx:      context.Background(),
			url:      gofakeit.URL(),
			expected: ErrNotValidURL,
		},
		{
			name:     "error content",
			repo:     fakeRepository(&mockError{}),
			ctx:      context.Background(),
			url:      "https://github.com/owner/repo/tree/branch/directory",
			expected: fmt.Errorf("failed to download: %w", errMockContents),
		},
		{
			name:     "error download file",
			repo:     fakeRepository(&mockError{}),
			ctx:      context.Background(),
			url:      "https://github.com/owner/repo/tree/branch/" + testDownloadFail,
			expected: fmt.Errorf("failed to download: %w", ErrInvalidPathURL),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := fakeNew(test.repo)
			err := g.Download(test.ctx, test.url)
			assert.Equal(t, test.expected, err)
		})
	}
}

func TestAuth(t *testing.T) {
	t.Parallel()
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
			expected: fmt.Errorf("failed to check auth: %w", errMockGetUser),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := fakeNew(test.repo)
			err := g.Auth(context.Background())
			assert.Equal(t, test.expected, err)
		})
	}
}
