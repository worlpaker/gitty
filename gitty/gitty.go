package gitty

import (
	"context"
	"fmt"
	"time"
)

// Git represents repository attributes.
type Git struct {
	repo Repository
}

// Gitty defines methods for interacting with cmd.
type Gitty interface {
	Status(ctx context.Context) error
	Auth(ctx context.Context) error
	Download(ctx context.Context, url string) error
}

// Ensure Git implements the Gitty interface.
var _ Gitty = (*Git)(nil)

// New creates a new Gitty.
func New() Gitty {
	client := newClient()
	r := repository(client)
	return &Git{
		repo: r,
	}
}

// Status reports the status of the client.
func (g *Git) Status(ctx context.Context) error {
	return g.repo.status(ctx)
}

// Auth reports the authenticated username.
func (g *Git) Auth(ctx context.Context) error {
	return g.repo.auth(ctx)
}

// Download downloads the contents from the given URL. It extracts the URL,
// collects the contents, and downloads files concurrently.
func (g *Git) Download(ctx context.Context, url string) error {
	fmt.Println("Downloading:", url)
	start := time.Now()

	if err := g.repo.extract(url); err != nil {
		return err
	}

	if err := g.repo.download(ctx); err != nil {
		return err
	}

	fmt.Println("Download Completed")
	fmt.Println(time.Since(start))

	return nil
}
