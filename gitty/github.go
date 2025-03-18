package gitty

import (
	"context"
	"net/http"

	"github.com/google/go-github/v67/github"
	"github.com/worlpaker/gitty/gitty/token"
)

// GitHub represents a GitHub repository with specific attributes.
type GitHub struct {
	Client Client
	Owner  string
	Repo   string
	Ref    *github.RepositoryContentGetOptions
	Path   string
}

// service represents a GitHub client that interacts with the GitHub API.
type service struct {
	client *github.Client
}

// newClient creates a new authenticated GitHub client using a provided access token, if any.
func newClient() *github.Client {
	c := github.NewClient(nil)
	if token.Get() == "" {
		return c
	}

	return c.WithAuthToken(token.Get())
}

// Client defines methods for interacting with [go-github] API.
//
// [go-github]: https://github.com/google/go-github
type Client interface {
	Get(url string) (resp *http.Response, err error)
	GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error)
	RateLimit(ctx context.Context) (*github.RateLimits, *github.Response, error)
	GetUser(ctx context.Context, user string) (*github.User, *github.Response, error)
}

// Ensure service implements the Client interface.
var _ Client = (*service)(nil)

// Get issues a GET to the specified URL. If the response is one of the
// following redirect codes, Get follows the redirect after calling the
// [Client.CheckRedirect] function:
//
//	301 (Moved Permanently)
//	302 (Found)
//	303 (See Other)
//	307 (Temporary Redirect)
//	308 (Permanent Redirect)
//
// An error is returned if the [Client.CheckRedirect] function fails
// or if there was an HTTP protocol error. A non-2xx response doesn't
// cause an error. Any returned error will be of type [*url.Error]. The
// url.Error value's Timeout method will report true if the request
// timed out.
//
// When err is nil, resp always contains a non-nil resp.Body.
// Caller should close resp.Body when done reading from it.
func (s *service) Get(url string) (resp *http.Response, err error) {
	return s.client.Client().Get(url)
}

// GetContents can return either the metadata and content of a single file
// (when path references a file) or the metadata of all the files and/or
// subdirectories of a directory (when path references a directory). To make it
// easy to distinguish between both result types and to mimic the API as much
// as possible, both result types will be returned but only one will contain a
// value and the other will be nil.
//
// Due to an auth vulnerability issue in the GitHub v3 API, ".." is not allowed
// to appear anywhere in the "path" or this method will return an error.
//
// GitHub API docs: https://docs.github.com/rest/repos/contents#get-repository-content
//
//meta:operation GET /repos/{owner}/{repo}/contents/{path}
func (s *service) GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (fileContent *github.RepositoryContent, directoryContent []*github.RepositoryContent, resp *github.Response, err error) {
	return s.client.Repositories.GetContents(ctx, owner, repo, path, opts)
}

// RateLimit returns the rate limits for the current client.
//
// GitHub API docs: https://docs.github.com/rest/rate-limit/rate-limit#get-rate-limit-status-for-the-authenticated-user
//
//meta:operation GET /rate_limit
func (s *service) RateLimit(ctx context.Context) (*github.RateLimits, *github.Response, error) {
	return s.client.RateLimit.Get(ctx)
}

// GetUser fetches a user. Passing the empty string will fetch the authenticated
// user.
//
// GitHub API docs: https://docs.github.com/rest/users/users#get-a-user
//
// GitHub API docs: https://docs.github.com/rest/users/users#get-the-authenticated-user
//
//meta:operation GET /user
//meta:operation GET /users/{username}
func (s *service) GetUser(ctx context.Context, user string) (*github.User, *github.Response, error) {
	return s.client.Users.Get(ctx, user)
}
