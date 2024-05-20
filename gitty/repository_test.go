package gitty

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/go-github/v61/github"
	"github.com/stretchr/testify/assert"
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
		Ref:    "",
		Path:   "",
	}
}

func (m *mockSuccess) Get(url string) (resp *http.Response, err error) {
	resp = &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte("test data"))),
	}
	return
}

func (m *mockError) Get(url string) (resp *http.Response, err error) {
	return &http.Response{}, errMockGet
}

// pStr returns a pointer to the provided string.
func pStr(s string) *string {
	return &s
}

// contenstsData for testing Contents.
var contentsData = []*github.RepositoryContent{
	{
		Type:        pStr("file"),
		Path:        pStr(gofakeit.LoremIpsumWord()),
		DownloadURL: pStr(gofakeit.LoremIpsumWord()),
	},
	{
		Type:        pStr("file"),
		Path:        pStr(gofakeit.LoremIpsumWord()),
		DownloadURL: pStr(gofakeit.LoremIpsumWord()),
	},
	{
		Type: pStr("dir"),
		Path: pStr("dir"),
	},
}

// testDownloadFailData for testing Download failure.
var testDownloadFailData = []*github.RepositoryContent{
	{Type: pStr("file")},
	{Type: pStr("file")},
}

// Contents test paths.
const (
	testFileOnly     = "testFileOnly"
	testDownloadFail = "testDownloadFail"
	testContentFail  = "testContentFail"
)

func (m *mockSuccess) GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error) {
	switch path {
	case testFileOnly:
		return contentsData[0], nil, nil, nil
	case "dir":
		return nil, contentsData[:len(contentsData)-1], nil, nil
	}

	return nil, contentsData, nil, nil
}

func (m *mockError) GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error) {
	if path == testDownloadFail {
		return testDownloadFailData[0], nil, nil, nil
	}
	if path == testContentFail {
		return nil, testDownloadFailData, nil, nil
	}
	if ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, nil, nil, context.Canceled
		}
		return nil, nil, nil, ErrTookTooLong
	}
	return nil, nil, nil, errMockContents
}

func (m *mockSuccess) RateLimit(ctx context.Context) (*github.RateLimits, *github.Response, error) {
	r := &github.RateLimits{
		Core: &github.Rate{
			Limit:     5000,
			Remaining: 50,
			Reset:     github.Timestamp{Time: time.Now().Add(1 * time.Hour)},
		},
	}
	return r, nil, nil
}

func (m *mockError) RateLimit(ctx context.Context) (*github.RateLimits, *github.Response, error) {
	return nil, nil, errMockRateLimit
}

func (m *mockSuccess) GetUser(ctx context.Context, user string) (*github.User, *github.Response, error) {
	u := &github.User{
		Name: &user,
	}
	return u, nil, nil
}

func (m *mockError) GetUser(ctx context.Context, user string) (*github.User, *github.Response, error) {
	return nil, nil, errMockGetUser
}

func TestRepository(t *testing.T) {
	c := github.NewClient(nil)
	actual := repository(c)
	expected := &GitHub{
		Client: &service{
			client: c,
		},
		Owner: "",
		Repo:  "",
		Ref:   "",
		Path:  "",
	}
	assert.Equal(t, expected, actual)
}

func TestExtract(t *testing.T) {
	c := &mockSuccess{}
	r := fakeRepository(c)
	reset := func() {
		r.(*GitHub).Owner = ""
		r.(*GitHub).Repo = ""
		r.(*GitHub).Ref = ""
		r.(*GitHub).Path = ""

	}
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
				Ref:   "branch",
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
				Ref:   "branch",
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
				Ref:   "branch",
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
				Ref:   "",
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
				Ref:   "",
				Path:  "",
			},
			expectedErr: ErrNotValidURL,
		},
	}

	for _, test := range tests {
		err := r.extract(test.url)
		t.Run(test.name, func(t *testing.T) {
			t.Cleanup(reset)
			assert.Equal(t, test.expectedErr, err)
			assert.Equal(t, test.expected.Owner, r.(*GitHub).Owner)
			assert.Equal(t, test.expected.Repo, r.(*GitHub).Repo)
			assert.Equal(t, test.expected.Ref, r.(*GitHub).Ref)
			assert.Equal(t, test.expected.Path, r.(*GitHub).Path)
		})
	}
}

func TestDownloadFile(t *testing.T) {
	// Discard output during tests.
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	fakePath := gofakeit.LoremIpsumWord()
	// Don't use gofakeit.FileExtension(), it might create "zip", "rar".
	fakeExt := "go"
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
			path:     fmt.Sprintf("%s/%s.%s", fakePath, fakePath, fakeExt),
			expected: nil,
		},
		{
			name:     "error download file",
			repo:     fakeRepository(&mockError{}),
			url:      gofakeit.URL(),
			path:     fmt.Sprintf("%s/%s.%s", fakePath, fakePath, fakeExt),
			expected: errMockGet,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.repo.downloadFile(test.url, test.path)
			assert.Equal(t, test.expected, err)
			if test.expected == nil {
				err := os.Remove(test.path)
				assert.Nil(t, err)
				err = os.RemoveAll(filepath.Dir(test.path))
				assert.Nil(t, err)
			}
		})
	}
}

// deleteTestFiles removes test files after the test.
func deleteTestFiles(t *testing.T, paths ...string) {
	t.Helper()
	for _, path := range paths {
		err := os.RemoveAll(path)
		assert.Nil(t, err)
	}
}

func TestDownloadContents(t *testing.T) {
	// Discard output during tests.
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

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
			ctx:      context.Background(),
			path:     "directory",
			expected: nil,
		},
		{
			name:     "error get contents",
			repo:     fakeRepository(&mockError{}),
			ctx:      context.Background(),
			expected: errMockContents,
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
			if test.expected == nil {
				t.Cleanup(func() {
					paths := []string{contentsData[0].GetPath(), contentsData[1].GetPath()}
					deleteTestFiles(t, paths...)
				})
			}

			err := test.repo.downloadContents(test.ctx)
			assert.Equal(t, test.expected, err)
		})
	}
}

func TestContents(t *testing.T) {
	// Discard output during tests.
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	tests := []struct {
		name      string
		repo      Repository
		path      string
		filePaths []string
		expected  error
	}{
		{
			name:      "successfully get contents with directories",
			repo:      fakeRepository(&mockSuccess{}),
			path:      "directory",
			filePaths: []string{contentsData[0].GetPath(), contentsData[1].GetPath()},
			expected:  nil,
		},
		{
			name:      "successfully get contents file only",
			repo:      fakeRepository(&mockSuccess{}),
			path:      testFileOnly,
			filePaths: []string{contentsData[0].GetPath()},
			expected:  nil,
		},
		{
			name:     "error contents",
			repo:     fakeRepository(&mockError{}),
			expected: errMockContents,
		},
		{
			name:     "error downloading file",
			path:     testContentFail,
			repo:     fakeRepository(&mockError{}),
			expected: ErrInvalidPathURL,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expected == nil {
				t.Cleanup(func() {
					deleteTestFiles(t, test.filePaths...)
				})
			}

			wg := &sync.WaitGroup{}
			errCh := make(chan error, 1)

			wg.Add(1)
			go test.repo.contents(context.Background(), wg, test.path, errCh)
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
			expectedErr: errMockRateLimit,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.auth {
				err := os.Setenv(tokenKey, gofakeit.LoremIpsumWord())
				assert.Nil(t, err)
				t.Cleanup(func() {
					err := os.Unsetenv(tokenKey)
					assert.Nil(t, err)
				})
			}
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := test.repo.clientStatus(context.Background())
			assert.Equal(t, test.expectedErr, err)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			assert.Nil(t, err)
			actual := buf.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestClientAuth(t *testing.T) {
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
			name:     "success auth",
			repo:     fakeRepository(&mockSuccess{}),
			expected: nil,
		},
		{
			name:     "error auth",
			repo:     fakeRepository(&mockError{}),
			expected: errMockGetUser,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.repo.clientAuth(context.Background())
			assert.Equal(t, test.expected, err)
		})
	}
}
