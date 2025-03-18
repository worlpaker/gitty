//go:build darwin || linux

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
	arg := fmt.Sprintf("export %s=%s && echo 'export %s=%s' >> ~/.bashrc", key, value, key, value)
	return &Script{
		cmd: exec.Command("bash", "-c", arg),
		msg: msgSaved,
	}
}

// del creates a script to delete a key from the os environment.
func del(key string) *Script {
	arg := fmt.Sprintf("sed -i '/^export %s=/d' ~/.bashrc", key)
	return &Script{
		cmd: exec.Command("bash", "-c", arg),
		msg: msgDeleted,
	}
}

// Run executes a script based on the provided key and value. If the value is empty,
// it deletes the key. Otherwise, it saves into the os environment.
func Run(key, value string) error {
	if value == "" {
		return del(key).execute()
	}
	return save(key, value).execute()
}
