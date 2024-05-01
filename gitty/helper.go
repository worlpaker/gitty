package gitty

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	hPrefix = "https://github.com/"
	prefix  = "github.com/"
)

var (
	ErrNotValidURL       = errors.New("url must starts with https://github.com/ or github.com/")
	ErrNotValidURLFormat = errors.New("url format must be: https://github.com/owner/repo/tree/branch/directory")
)

// getGitHubRepo parses and extracts the repository path from a GitHub URL.
func getGitHubRepo(url string) (string, error) {
	h, hFound := strings.CutPrefix(url, hPrefix)
	p, found := strings.CutPrefix(url, prefix)

	switch {
	case hFound:
		if isInvalidFormat(h) {
			return "", ErrNotValidURLFormat
		}
		return h, nil
	case found:
		if isInvalidFormat(p) {
			return "", ErrNotValidURLFormat
		}
		return p, nil
	default:
		return "", ErrNotValidURL
	}
}

// isInvalidFormat checks if the URL has an invalid format.
func isInvalidFormat(s string) bool {
	// Valid format example is: https://github.com/owner/repo/tree/branch/directory
	// After the domain, the expected format is: owner/repo/tree/branch/directory
	return strings.Count(s, "/") < 4
}

// saveFile saves the content of the file at the specified path.
func saveFile(base, path string, body io.Reader) error {
	const perm = 0755
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	p, err := exactPath(base, path)
	if err != nil {
		return err
	}
	fmt.Println("Saving:", p)

	if err := os.MkdirAll(filepath.Dir(p), perm); err != nil {
		return err
	}

	return os.WriteFile(p, data, perm)
}

// exactPath removes unnecessary directories from the given path.
func exactPath(base, path string) (string, error) {
	basePath := filepath.Base(base)
	relPath, err := filepath.Rel(base, path)
	if err != nil {
		return "", err
	}

	return filepath.Join(basePath, relPath), nil
}
