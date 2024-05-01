//go:build windows

package script

import (
	"fmt"
	"os/exec"
)

var (
	msgSaved   = "Specified token was saved. Please reload your shell."
	msgDeleted = "Specified token was deleted. Please reload your shell."
)

// Script represents a shell script.
type Script struct {
	cmd *exec.Cmd
	msg string
}

// execute runs the shell script.
func (s *Script) execute() error {
	if _, err := s.cmd.CombinedOutput(); err != nil {
		return err
	}
	fmt.Println(s.msg)
	return nil
}

// save creates a script to set an environment variable with a specified key and value.
func save(key, value string) *Script {
	return &Script{
		cmd: exec.Command("cmd", "/c", "setx", key, value),
		msg: msgSaved,
	}
}

// delete creates a script to delete a key from the os environment.
func delete(key string) *Script {
	return &Script{
		cmd: exec.Command("cmd", "/c", "reg delete HKCU\\Environment /F /V", key),
		msg: msgDeleted,
	}
}

// Run executes a script based on the provided key and value. If the value is empty,
// it deletes the key. Otherwise, it saves into the os environment.
func Run(key, value string) error {
	if value == "" {
		return delete(key).execute()
	}
	return save(key, value).execute()
}
