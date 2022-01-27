//+build windows

package ssh_server

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
)

func createPty(shell string, connection ssh.Channel) (*os.File, error) {
	return nil, fmt.Errorf("creation of pty unsupported")
}

// setWinsize sets the size of the given pty.
func setWinsize(t *os.File, w, h uint32) error {
	return fmt.Errorf("set-win-size unsupported")
}
