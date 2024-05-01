//go:build !windows && !darwin && !linux

package script

import (
	"fmt"
	"runtime"
)

// Run returns an error for unsupported os.
func Run(key, value string) error {
	return fmt.Errorf("failed to config token: %v is currently unsupported", runtime.GOOS)
}
