package token

import (
	"os"

	"github.com/worlpaker/gitty/gitty/token/script"
)

const key = "GH_TOKEN"

// Get retrieves the GitHub token from the environment variable.
func Get() string {
	return os.Getenv(key)
}

// Set sets the provided GitHub token in the environment variable.
func Set(token string) error {
	return script.Run(key, token)
}

// Unset unsets the GitHub token from the environment variable.
func Unset() error {
	return script.Run(key, "")
}
