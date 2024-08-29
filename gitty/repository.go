package gitty

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v64/github"
	"github.com/worlpaker/gitty/gitty/token"
)

const (
	// baseRateLimit represents the number of unauthenticated requests limited per hour.
	baseRateLimit = 60
	// downloadLimit represents the number of seconds limited per download request.
	// If it takes more than downloadLimit seconds, it returns ErrTookTooLong.
	downloadLimit = 60
)

var (
	ErrTookTooLong    = errors.New("took more than 60 seconds to download contents")
	ErrInvalidPathURL = errors.New("invalid url or path")
)

// Repository defines methods for interacting with GitHub.
type Repository interface {
	extract(url string) error
	downloadContents(ctx context.Context) error
	contents(ctx context.Context, wg *sync.WaitGroup, path string, errCh chan error)
	downloadFile(url, path string) error
	clientStatus(ctx context.Context) error
	clientAuth(ctx context.Context) error
}

// Ensure GitHub implements the Repository interface.
var _ Repository = &GitHub{}

// repository creates a GitHub repository with default values.
func repository(c *github.Client) Repository {
	return &GitHub{
		Client: &service{
			client: c,
		},
		Owner: "",
		Repo:  "",
		Ref:   "",
		Path:  "",
	}
}

// extract parses a GitHub URL and extracts the owner, repository name, reference,
// and path from it. It sets these values in the GitHub struct.
func (g *GitHub) extract(url string) error {
	s, err := getGitHubRepo(url)
	if err != nil {
		return err
	}

	sep := "/"
	strs := strings.Split(s, sep)
	g.Owner = strs[0]
	g.Repo = strs[1]
	g.Ref = strs[3]
	g.Path = strings.Join(strs[4:], sep)

	return nil
}

// downloadContents downloads the contents concurrently.
func (g *GitHub) downloadContents(ctx context.Context) error {
	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)

	wg.Add(1)
	go g.contents(ctx, wg, g.Path, errCh)

	go func() {
		defer func() {
			wg.Wait()
			close(errCh)
		}()
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.Canceled) {
			return context.Canceled
		}
		return ErrTookTooLong
	}

	return nil
}

// contents retrieves the contents of the GitHub directory path. It recursively
// collects subdirectories, if any. It downloads files concurrently.
func (g *GitHub) contents(ctx context.Context, wg *sync.WaitGroup, path string, errCh chan error) {
	defer wg.Done()

	opts := &github.RepositoryContentGetOptions{
		Ref: g.Ref,
	}
	fileContent, directoryContent, _, err := g.Client.GetContents(ctx, g.Owner, g.Repo, path, opts)
	if err != nil {
		errCh <- err
		return
	}

	// If the URL points to a file, only the file is downloaded.
	if len(directoryContent) == 0 && fileContent != nil {
		if err := g.downloadFile(fileContent.GetDownloadURL(), fileContent.GetPath()); err != nil {
			errCh <- err
			return
		}
		return
	}

	// Collect all subcontents of subdirectories.
	for _, content := range directoryContent {
		wg.Add(1)
		go func(content *github.RepositoryContent) {
			defer wg.Done()
			switch content.GetType() {
			case "file":
				// Download the file directly.
				if err := g.downloadFile(content.GetDownloadURL(), content.GetPath()); err != nil {
					errCh <- err
					return
				}
			case "dir":
				// Recursively get the files of the content.
				wg.Add(1)
				go g.contents(ctx, wg, content.GetPath(), errCh)
			}
		}(content)
	}
}

// downloadFile downloads a file from the given URL and saves it.
func (g *GitHub) downloadFile(url, path string) error {
	if url == "" || path == "" {
		return ErrInvalidPathURL
	}

	fmt.Println("Downloading:", path)
	resp, err := g.Client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return saveFile(g.Path, path, resp.Body)
}

// clientStatus reports the status of the client, the remaining hourly
// rate limit, and the time at which the current rate limit will reset.
// This function does not reduce the rate limit. It can be used freely.
func (g *GitHub) clientStatus(ctx context.Context) error {
	auth := "NOT Authorized"
	rate, _, err := g.Client.RateLimit(ctx)
	if err != nil {
		return err
	}

	if token.Get() != "" && rate.Core.Limit > baseRateLimit {
		auth = "Authorized"
	}

	reset := time.Until(rate.Core.Reset.Time).Minutes()
	fmt.Printf("Status: %v | Remaining rate limit: %v | Reset in: %.0f mins \n", auth, rate.Core.Remaining, reset)

	return nil
}

// clientAuth reports the authenticated username, if applicable.
// This function reduces the rate limit for each request.
func (g *GitHub) clientAuth(ctx context.Context) error {
	u, _, err := g.Client.GetUser(ctx, "")
	if err != nil {
		return err
	}
	fmt.Printf("Authenticated as @%s \n", u.GetLogin())

	return nil
}
