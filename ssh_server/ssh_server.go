// (base: https://gist.github.com/jpillora/b480fde82bff51a06238)

package ssh_server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"github.com/creack/pty"
	"github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"unsafe"
)

func HandleChannels(shell string, chans <-chan ssh.NewChannel) {
	// Service the incoming Channel channel in go routine
	for newChannel := range chans {
		go handleChannel(shell, newChannel)
	}
}

func handleChannel(shell string, newChannel ssh.NewChannel) {
	switch newChannel.ChannelType() {
	case "session":
		handleSession(shell, newChannel)
	case "direct-tcpip":
		handleDirectTcpip(newChannel)
	default:
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", newChannel.ChannelType()))
	}
}

func handleSession(shell string, newChannel ssh.NewChannel) {
	// At this point, we have the opportunity to reject the client's
	// request for another logical connection
	connection, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}

	var shf *os.File = nil

	for req := range requests {
		switch req.Type {
		case "exec":
			var msg struct {
				Command string
			}
			if err := ssh.Unmarshal(req.Payload, &msg); err != nil {
				log.Printf("error in parse message (%v) in exec", err)
				return
			}
			cmdSlice, err := shellwords.Parse(msg.Command)
			if err != nil {
				return
			}
			cmd := exec.Command(cmdSlice[0], cmdSlice[1:]...)
			stdin, err := cmd.StdinPipe()
			if err != nil {
				return
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return
			}
			stderr, err := cmd.StderrPipe()
			if err != nil {
				return
			}
			go io.Copy(stdin, connection)
			go io.Copy(connection, stdout)
			go io.Copy(connection, stderr)
			req.Reply(true, nil)
			var exitCode int
			if err := cmd.Run(); err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				}
			}
			resMsg := struct{ Status uint32 }{Status: uint32(exitCode)}
			connection.SendRequest("exit-status", false, ssh.Marshal(resMsg))
			connection.Close()
		case "shell":
			// We only accept the default shell
			// (i.e. no command in the Payload)
			if len(req.Payload) == 0 {
				req.Reply(true, nil)
			}
		case "pty-req":
			termLen := req.Payload[3]
			w, h := parseDims(req.Payload[termLen+4:])
			shf, err = createPty(shell, connection)
			if err != nil {
				return
			}
			SetWinsize(shf.Fd(), w, h)
			// Responding true (OK) here will let the client
			// know we have a pty ready for input
			req.Reply(true, nil)
		case "window-change":
			w, h := parseDims(req.Payload)
			if shf != nil {
				SetWinsize(shf.Fd(), w, h)
			}
		}
	}
}

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

// (base: https://github.com/peertechde/zodiac/blob/110fdd2dfd27359546c1cd75a9fec5de2882bf42/pkg/server/server.go#L228)
func handleDirectTcpip(newChannel ssh.NewChannel) {
	var msg struct {
		RemoteAddr string
		RemotePort uint32
		SourceAddr string
		SourcePort uint32
	}
	if err := ssh.Unmarshal(newChannel.ExtraData(), &msg); err != nil {
		log.Printf("error in parse message (%v)", err)
		return
	}
	channel, reqs, err := newChannel.Accept()
	if err != nil {
		log.Printf("accept error (%v)", err)
		return
	}
	go ssh.DiscardRequests(reqs)
	raddr := net.JoinHostPort(msg.RemoteAddr, strconv.Itoa(int(msg.RemotePort)))
	conn, err := net.Dial("tcp", raddr)
	if err != nil {
		log.Printf("dial error (%v)", err)
		channel.Close()
		return
	}
	var closeOnce sync.Once
	closer := func() {
		channel.Close()
		conn.Close()
	}
	go func() {
		io.Copy(channel, conn)
		closeOnce.Do(closer)
	}()
	io.Copy(conn, channel)
	closeOnce.Do(closer)
	return
}

// =======================

// parseDims extracts terminal dimensions (width x height) from the provided buffer.
func parseDims(b []byte) (uint32, uint32) {
	w := binary.BigEndian.Uint32(b)
	h := binary.BigEndian.Uint32(b[4:])
	return w, h
}

// ======================

// Winsize stores the Height and Width of a terminal.
type Winsize struct {
	Height uint16
	Width  uint16
	x      uint16 // unused
	y      uint16 // unused
}

// SetWinsize sets the size of the given pty.
func SetWinsize(fd uintptr, w, h uint32) {
	ws := &Winsize{Width: uint16(w), Height: uint16(h)}
	syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSWINSZ), uintptr(unsafe.Pointer(ws)))
}

func GenerateKey() ([]byte, error) {
	var r io.Reader
	r = rand.Reader
	priv, err := rsa.GenerateKey(r, 2048)
	if err != nil {
		return nil, err
	}
	err = priv.Validate()
	if err != nil {
		return nil, err
	}
	b := x509.MarshalPKCS1PrivateKey(priv)
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: b}), nil
}

// Borrowed from https://github.com/creack/termios/blob/master/win/win.go
