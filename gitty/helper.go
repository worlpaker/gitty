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
	ErrNotValidURL    = errors.New("url must starts with https://github.com/ or github.com/")
	ErrNotValidFormat = errors.New("url format must be https://github.com/owner/repo/tree/branch/directory")
)

// getGitHubRepo parses and extracts the repository path from a GitHub URL.
func getGitHubRepo(url string) (string, error) {
	prefixes := []string{hPrefix, prefix}
	for _, pref := range prefixes {
		if path, ok := strings.CutPrefix(url, pref); ok {
			return validate(path)
		}
	}
	return "", ErrNotValidURL
}

// validate checks if the URL has a valid format.
func validate(s string) (string, error) {
	// Valid format example is: https://github.com/owner/repo/tree/branch/directory
	// After the domain, the expected format is: owner/repo/tree/branch/directory
	if strings.Count(s, "/") < 4 {
		return "", ErrNotValidFormat
	}
	return s, nil
}

// saveFile saves the content of the file at the specified path.
func saveFile(base, path string, body io.Reader) error {
	p, err := exactPath(base, path)
	if err != nil {
		return err
	}
	fmt.Println("Saving:", p)

	if errMkdir := os.MkdirAll(filepath.Dir(p), os.ModePerm); errMkdir != nil {
		return errMkdir
	}

	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, body); err != nil {
		return err
	}

	return nil
}

// exactPath removes unnecessary directories from the given path.
func exactPath(base, path string) (string, error) {
	relPath, err := filepath.Rel(base, path)
	if err != nil {
		return "", err
	}

	return filepath.Join(filepath.Base(base), relPath), nil
}
