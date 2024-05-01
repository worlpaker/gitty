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
	Download(ctx context.Context, url string) error
	Auth(ctx context.Context) error
}

// Ensure Git implements the Gitty interface.
var _ Gitty = &Git{}

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
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := g.repo.clientStatus(ctx); err != nil {
		return fmt.Errorf("failed to check status: %v", err)
	}

	return nil
}

// Download downloads the contents of the given URL. It extracts url,
// collects contents, and downloads files concurrently.
func (g *Git) Download(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, downloadLimit*time.Second)
	start := time.Now()
	defer func() {
		cancel()
		fmt.Println(time.Since(start))
	}()

	if err := g.repo.extract(url); err != nil {
		return err
	}

	fmt.Println("Downloading:", url)
	if err := g.repo.downloadContents(ctx); err != nil {
		return fmt.Errorf("failed to download: %v", err)
	}
	fmt.Println("Download Completed")

	return nil
}

// Auth reports the auth status of the user.
func (g *Git) Auth(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := g.repo.clientAuth(ctx); err != nil {
		return fmt.Errorf("failed to check auth: %v", err)
	}

	return nil
}
