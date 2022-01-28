// +build !windows
// NOTE: pty.Start() is not supported in Windows

package ssh_server

import (
	"github.com/creack/pty"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
)

func createPty(shell string, connection ssh.Channel) (*os.File, error) {
	if shell == "" {
		shell = os.Getenv("SHELL")
	}
	if shell == "" {
		shell = "sh"
	}
	// Fire up bash for this session
	sh := exec.Command(shell)

	// Prepare teardown function
	closer := func() {
		connection.SendRequest("exit-status", false, ssh.Marshal(exitStatusMsg{
			Status: 0,
		}))
		connection.Close()
		_, err := sh.Process.Wait()
		if err != nil {
			log.Printf("Failed to exit bash (%s)", err)
		}
		log.Printf("Session closed")
	}

	// Allocate a terminal for this channel
	log.Print("Creating pty...")
	shf, err := pty.Start(sh)
	if err != nil {
		log.Printf("Could not start pty (%s)", err)
		closer()
		return nil, errors.Errorf("could not start pty (%s)", err)
	}

	// pipe session to bash and visa-versa
	var once sync.Once
	go func() {
		io.Copy(connection, shf)
		once.Do(closer)
	}()
	go func() {
		io.Copy(shf, connection)
		once.Do(closer)
	}()
	return shf, nil
}

// setWinsize sets the size of the given pty.
func setWinsize(t *os.File, w, h uint32) error {
	return pty.Setsize(t, &pty.Winsize{Rows: uint16(h), Cols: uint16(w)})
}
