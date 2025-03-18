package gitty

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/go-github/v70/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errMockRateLimit = errors.New("mock ratelimit error")
	errMockGet       = errors.New("mock get error")
	errMockContents  = errors.New("mock contents error")
	errMockGetUser   = errors.New("mock getuser error")
)

type mockSuccess struct{}

type mockError struct{}

type mockClient interface {
	Get(url string) (resp *http.Response, err error)
	GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error)
	RateLimit(ctx context.Context) (*github.RateLimits, *github.Response, error)
	GetUser(ctx context.Context, user string) (*github.User, *github.Response, error)
}

func fakeRepository(c mockClient) Repository {
	return &GitHub{
		Client: c,
		Owner:  "",
		Repo:   "",
		Ref:    nil,
		Path:   "",
	}
}

func (m *mockSuccess) Get(_ string) (resp *http.Response, err error) {
	resp = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte("test data"))),
	}
	return
}

func (m *mockError) Get(_ string) (resp *http.Response, err error) {
	return &http.Response{}, errMockGet
}

// ptr returns a pointer to the provided value.
func ptr[T any](t T) *T {
	return &t
}

type contextPathKey string

const pathKey contextPathKey = "fakepath"

// contenstsData for testing Contents.
func contentsData(firstPath, secondPath string) []*github.RepositoryContent {
	return []*github.RepositoryContent{
		{
			Type:        ptr("file"),
			Path:        ptr(firstPath),
			DownloadURL: ptr(gofakeit.URL()),
		},
		{
			Type:        ptr("file"),
			Path:        ptr(secondPath),
			DownloadURL: ptr(gofakeit.URL()),
		},
		{
			Type: ptr("dir"),
			Path: ptr("dir"),
		},
	}
}

// testDownloadFailData for testing Download failure.
func testDownloadFailData() []*github.RepositoryContent {
	return []*github.RepositoryContent{
		{Type: ptr("file")},
		{Type: ptr("file")},
	}
}

// Contents test paths.
const (
	testFileOnly     = "testFileOnly"
	testDownloadFail = "testDownloadFail"
	testContentFail  = "testContentFail"
)

func (m *mockSuccess) GetContents(ctx context.Context, _, _, path string, _ *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error) {
	fakeData, ok := ctx.Value(pathKey).([]*github.RepositoryContent)
	if !ok {
		fakeData = contentsData("tmp/file_0", "tmp/file_1")
	}

	switch path {
	case testFileOnly:
		return fakeData[0], nil, nil, nil
	case "dir":
		return nil, fakeData[:len(fakeData)-1], nil, nil
	default:
		return nil, fakeData, nil, nil
	}
}

func (m *mockError) GetContents(ctx context.Context, _, _, path string, _ *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error) {
	if path == testDownloadFail {
		return testDownloadFailData()[0], nil, nil, nil
	}
	if path == testContentFail {
		return nil, testDownloadFailData(), nil, nil
	}
	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, nil, nil, context.Canceled
		}
		return nil, nil, nil, ErrTookTooLong
	}
	return nil, nil, nil, errMockContents
}

func (m *mockSuccess) RateLimit(_ context.Context) (*github.RateLimits, *github.Response, error) {
	r := &github.RateLimits{
		Core: &github.Rate{
			Limit:     5000,
			Remaining: 50,
			Reset:     github.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		},
	}
	return r, nil, nil
}

func (m *mockError) RateLimit(_ context.Context) (*github.RateLimits, *github.Response, error) {
	return nil, nil, errMockRateLimit
}

func (m *mockSuccess) GetUser(_ context.Context, user string) (*github.User, *github.Response, error) {
	u := &github.User{
		Name: &user,
	}
	return u, nil, nil
}

func (m *mockError) GetUser(_ context.Context, _ string) (*github.User, *github.Response, error) {
	return nil, nil, errMockGetUser
}

func TestRepository(t *testing.T) {
	t.Parallel()
	c := github.NewClient(nil)
	actual := repository(c)
	expected := &GitHub{
		Client: &service{
			client: c,
		},
		Owner: "",
		Repo:  "",
		Ref:   nil,
		Path:  "",
	}
	assert.Equal(t, expected, actual)
}

func TestExtract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		url         string
		expected    *GitHub
		expectedErr error
	}{
		{
			name: "valid url with https://github.com/",
			url:  "https://github.com/owner/repo/tree/branch/directory",
			expected: &GitHub{
				Owner: "owner",
				Repo:  "repo",
				Ref:   &github.RepositoryContentGetOptions{Ref: "branch"},
				Path:  "directory",
			},
			expectedErr: nil,
		},
		{
			name: "valid url with github.com/",
			url:  "github.com/owner/repo/tree/branch/directory1/directory2",
			expected: &GitHub{
				Owner: "owner",
				Repo:  "repo",
				Ref:   &github.RepositoryContentGetOptions{Ref: "branch"},
				Path:  "directory1/directory2",
			},
			expectedErr: nil,
		},
		{
			name: "valid url file with github.com/",
			url:  "github.com/owner/repo/tree/branch/directory1/directory2/file.txt",
			expected: &GitHub{
				Owner: "owner",
				Repo:  "repo",
				Ref:   &github.RepositoryContentGetOptions{Ref: "branch"},
				Path:  "directory1/directory2/file.txt",
			},
			expectedErr: nil,
		},
		{
			name: "invalid https url",
			url:  "https://gitlab.com/owner/repo/tree/branch/directory",
			expected: &GitHub{
				Owner: "",
				Repo:  "",
				Ref:   nil,
				Path:  "",
			},
			expectedErr: ErrNotValidURL,
		},
		{
			name: "invalid url",
			url:  "gitlab.com/owner/repo/tree/branch/directory",
			expected: &GitHub{
				Owner: "",
				Repo:  "",
				Ref:   nil,
				Path:  "",
			},
			expectedErr: ErrNotValidURL,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			r, ok := fakeRepository(&mockSuccess{}).(*GitHub)
			assert.True(t, ok)
			err := r.extract(test.url)
			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expected.Owner, r.Owner)
			assert.Equal(t, test.expected.Repo, r.Repo)
			assert.Equal(t, test.expected.Ref, r.Ref)
			assert.Equal(t, test.expected.Path, r.Path)
		})
	}
}

func TestGetFile(t *testing.T) {
	t.Parallel()
	fakeBase := fmt.Sprintf("%s_%d", gofakeit.LoremIpsumWord(), gofakeit.Int())
	// Don't use gofakeit.FileExtension(), it might create "zip", "rar".
	fakePath := fmt.Sprintf("%s/%s_%d.go", fakeBase, gofakeit.LoremIpsumWord(), gofakeit.Int())
	t.Cleanup(func() {
		err := os.RemoveAll(fakeBase)
		require.NoError(t, err)
	})
	tests := []struct {
		name     string
		repo     Repository
		url      string
		path     string
		expected error
	}{
		{
			name:     "invalid url path",
			repo:     fakeRepository(&mockSuccess{}),
			url:      "",
			path:     "",
			expected: ErrInvalidPathURL,
		},
		{
			name:     "successfully download file",
			repo:     fakeRepository(&mockSuccess{}),
			url:      gofakeit.URL(),
			path:     fakePath,
			expected: nil,
		},
		{
			name:     "error download file",
			repo:     fakeRepository(&mockError{}),
			url:      gofakeit.URL(),
			path:     fakePath,
			expected: errMockGet,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := test.repo.getFile(test.url, test.path)
			assert.Equal(t, test.expected, err)
		})
	}
}

func TestDownloadContents(t *testing.T) {
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
	ctxTimeOut := func() context.Context {
		ctx, cancel := context.WithTimeout(context.Background(), -10*time.Second)
		defer cancel()
		return ctx
	}
	ctxCancel := func() context.Context {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx
	}
	tests := []struct {
		name     string
		repo     Repository
		path     string
		ctx      context.Context
		expected error
	}{
		{
			name:     "successfully download directories and files",
			repo:     fakeRepository(&mockSuccess{}),
			ctx:      ctxfakePath(),
			path:     "directory",
			expected: nil,
		},
		{
			name:     "error get contents",
			repo:     fakeRepository(&mockError{}),
			ctx:      context.Background(),
			expected: fmt.Errorf("failed to download: %w", errMockContents),
		},
		{
			name:     "error ctx with timeout",
			repo:     fakeRepository(&mockError{}),
			ctx:      ctxTimeOut(),
			expected: ErrTookTooLong,
		},
		{
			name:     "error ctx cancel",
			repo:     fakeRepository(&mockError{}),
			ctx:      ctxCancel(),
			expected: context.Canceled,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := test.repo.download(test.ctx)
			assert.Equal(t, test.expected, err)
		})
	}
}

func TestContents(t *testing.T) {
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
		path     string
		expected error
	}{
		{
			name:     "successfully get contents with directories",
			repo:     fakeRepository(&mockSuccess{}),
			ctx:      ctxfakePath(),
			path:     "directory",
			expected: nil,
		},
		{
			name:     "successfully get contents file only",
			repo:     fakeRepository(&mockSuccess{}),
			ctx:      ctxfakePath(),
			path:     testFileOnly,
			expected: nil,
		},
		{
			name:     "error contents",
			repo:     fakeRepository(&mockError{}),
			ctx:      context.Background(),
			expected: errMockContents,
		},
		{
			name:     "error downloading file",
			path:     testContentFail,
			repo:     fakeRepository(&mockError{}),
			ctx:      context.Background(),
			expected: ErrInvalidPathURL,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			wg := &sync.WaitGroup{}
			errCh := make(chan error, 1)

			wg.Add(1)
			go test.repo.contents(test.ctx, wg, test.path, errCh)
			go func() {
				defer func() {
					wg.Wait()
					close(errCh)
				}()
			}()

			assert.Equal(t, test.expected, <-errCh)
		})
	}
}

func TestClientStatus(t *testing.T) {
	// Must be same as token const key.
	tokenKey := "GH_TOKEN"
	// This test also checks print outputs.
	tests := []struct {
		name        string
		repo        Repository
		auth        bool
		expected    string
		expectedErr error
	}{
		{
			name:        "success with authorized",
			repo:        fakeRepository(&mockSuccess{}),
			auth:        true,
			expected:    fmt.Sprintf("Status: %v | Remaining rate limit: %v | Reset in: %.0f mins \n", "Authorized", 50, float64(60)),
			expectedErr: nil,
		},
		{
			name:        "success with not authorized",
			repo:        fakeRepository(&mockSuccess{}),
			expected:    fmt.Sprintf("Status: %v | Remaining rate limit: %v | Reset in: %.0f mins \n", "NOT Authorized", 50, float64(60)),
			expectedErr: nil,
		},
		{
			name:        "error with not authorized",
			repo:        fakeRepository(&mockError{}),
			expected:    "",
			expectedErr: fmt.Errorf("failed to check status: %w", errMockRateLimit),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.auth {
				t.Setenv(tokenKey, gofakeit.LoremIpsumWord())
				t.Cleanup(func() {
					err := os.Unsetenv(tokenKey)
					require.NoError(t, err)
				})
			}
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := test.repo.status(context.Background())
			assert.Equal(t, test.expectedErr, err)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			require.NoError(t, err)
			actual := buf.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestClientAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		repo     Repository
		expected error
	}{
		{
			name:     "success auth",
			repo:     fakeRepository(&mockSuccess{}),
			expected: nil,
		},
		{
			name:     "error auth",
			repo:     fakeRepository(&mockError{}),
			expected: fmt.Errorf("failed to check auth: %w", errMockGetUser),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := test.repo.auth(context.Background())
			assert.Equal(t, test.expected, err)
		})
	}
}
