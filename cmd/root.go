package cmd

import (
	"fmt"
	"github.com/hashicorp/yamux"
	"github.com/nwtgck/go-piping-sshd/piping_util"
	"github.com/nwtgck/go-piping-sshd/priv_key"
	"github.com/nwtgck/go-piping-sshd/ssh_server"
	"github.com/nwtgck/go-piping-sshd/util"
	"github.com/nwtgck/go-piping-sshd/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	ServerUrlEnvName = "PIPING_SERVER"
)

var serverUrl string
var insecure bool
var dnsServer string
var showsVersion bool
var headerKeyValueStrs []string
var httpWriteBufSize int
var httpReadBufSize int
var sshYamux bool
var sshUser string
var sshPassword string
var sshShell string

func init() {
	cobra.OnInitialize()
	defaultServer, ok := os.LookupEnv(ServerUrlEnvName)
	if !ok {
		defaultServer = "https://ppng.io"
	}
	RootCmd.PersistentFlags().StringVarP(&serverUrl, "server", "s", defaultServer, "Piping Server URL")
	RootCmd.PersistentFlags().StringVar(&dnsServer, "dns-server", "", "DNS server (e.g. 1.1.1.1:53)")
	// NOTE: --insecure, -k is inspired by curl
	RootCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure server connections when using SSL")
	RootCmd.PersistentFlags().StringArrayVarP(&headerKeyValueStrs, "header", "H", []string{}, "HTTP header")
	RootCmd.PersistentFlags().IntVarP(&httpWriteBufSize, "http-write-buf-size", "", 4096, "HTTP write-buffer size in bytes")
	RootCmd.PersistentFlags().IntVarP(&httpReadBufSize, "http-read-buf-size", "", 4096, "HTTP read-buffer size in bytes")
	RootCmd.Flags().BoolVarP(&showsVersion, "version", "v", false, "show version")
	RootCmd.Flags().StringVarP(&sshUser, "user", "u", "", "SSH user name")
	RootCmd.Flags().StringVarP(&sshPassword, "password", "p", "", "SSH user password")
	RootCmd.Flags().StringVarP(&sshShell, "shell", "", "", "Shell")
	RootCmd.Flags().BoolVarP(&sshYamux, "yamux", "", false, "Multiplex connection by yamux")
}

var RootCmd = &cobra.Command{
	Use:          os.Args[0],
	Short:        "piping-sshd",
	Long:         "SSH server from anywhere with Piping Server",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if showsVersion {
			fmt.Println(version.Version)
			return nil
		}

		clientToServerPath, serverToClientPath, err := generatePaths(args)
		if err != nil {
			return err
		}
		headers, err := piping_util.ParseKeyValueStrings(headerKeyValueStrs)
		if err != nil {
			return err
		}
		httpClient := util.CreateHttpClient(insecure, httpWriteBufSize, httpReadBufSize)
		if dnsServer != "" {
			// Set DNS resolver
			httpClient.Transport.(*http.Transport).DialContext = util.CreateDialContext(dnsServer)
		}
		serverToClientUrl, err := util.UrlJoin(serverUrl, serverToClientPath)
		if err != nil {
			return err
		}
		clientToServerUrl, err := util.UrlJoin(serverUrl, clientToServerPath)
		if err != nil {
			return err
		}
		// Print hint
		sshPrintHintForClientHost(clientToServerUrl, serverToClientUrl, clientToServerPath, serverToClientPath)

		// (base: https://gist.github.com/jpillora/b480fde82bff51a06238)
		sshConfig := &ssh.ServerConfig{
			//Define a function to run when a client attempts a password login
			PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
				// Should use constant-time compare (or better, salt+hash) in a production setting.
				if (sshUser == "" || c.User() == sshUser) && string(pass) == sshPassword {
					return nil, nil
				}
				return nil, fmt.Errorf("password rejected for %q", c.User())
			},
			// No auth when password is empty
			NoClientAuth: sshPassword == "",
		}
		// TODO: specify priv_key by flags
		pri, err := ssh.ParsePrivateKey([]byte(priv_key.PrivateKeyPem))
		if err != nil {
			return err
		}
		sshConfig.AddHostKey(pri)

		// If not using multiplexer
		if !sshYamux {
			duplex, err := piping_util.DuplexConnect(httpClient, headers, serverToClientUrl, clientToServerUrl)
			if err != nil {
				return err
			}
			// Before use, a handshake must be performed on the incoming net.Conn.
			sshConn, chans, reqs, err := ssh.NewServerConn(util.NewDuplexConn(duplex), sshConfig)
			if err != nil {
				log.Printf("Failed to handshake (%s)", err)
				return nil
			}

			log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
			// Discard all global out-of-band Requests
			go ssh.DiscardRequests(reqs)
			// Accept all channels
			ssh_server.HandleChannels(sshShell, chans)
		}

		// If yamux is enabled
		if sshYamux {
			fmt.Println("[INFO] Multiplexing with yamux")
			return sshHandleWithYamux(sshConfig, httpClient, headers, clientToServerUrl, serverToClientUrl)
		}
		return nil
	},
}

func sshPrintHintForClientHost(clientToServerUrl string, serverToClientUrl string, clientToServerPath string, serverToClientPath string) {
	clientHostPort := 2022
	if !sshYamux {
		fmt.Println("=== Client host (socat + curl) ===")
		fmt.Printf(
			"  socat TCP-LISTEN:%d,reuseaddr 'EXEC:curl -NsS %s!!EXEC:curl -NsST - %s'\n",
			clientHostPort,
			strings.Replace(serverToClientUrl, ":", "\\:", -1),
			strings.Replace(clientToServerUrl, ":", "\\:", -1),
		)
	}
	flags := ""
	if sshYamux {
		flags += fmt.Sprintf("--%s ", yamuxFlagLongName)
	}
	fmt.Println("=== Client host (piping-tunnel) ===")
	fmt.Printf(
		"  piping-tunnel -s %s client -p %d %s%s %s\n",
		serverUrl,
		clientHostPort,
		flags,
		clientToServerPath,
		serverToClientPath,
	)
	fmt.Println("=== SSH client ===")
	userAndHost := "localhost"
	if sshUser != "" {
		userAndHost = sshUser + "@localhost"
	}
	fmt.Printf("  ssh-keygen -R [localhost]:%d; ssh -p %d %s\n", clientHostPort, clientHostPort, userAndHost)
}

func sshHandleWithYamux(sshConfig *ssh.ServerConfig, httpClient *http.Client, headers []piping_util.KeyValue, clientToServerUrl string, serverToClientUrl string) error {
	var duplex io.ReadWriteCloser
	duplex, err := piping_util.DuplexConnectWithHandlers(
		func(body io.Reader) (*http.Response, error) {
			return piping_util.PipingSend(httpClient, headersWithYamux(headers), serverToClientUrl, body)
		},
		func() (*http.Response, error) {
			res, err := piping_util.PipingGet(httpClient, headers, clientToServerUrl)
			if err != nil {
				return nil, err
			}
			contentTypes := res.Header.Values("Content-Type")
			// NOTE: No Content-Type is for curl user
			// NOTE: application/octet-stream is for compatibility
			if !(len(contentTypes) == 0 || contentTypes[0] == yamuxMimeType || contentTypes[0] == "application/octet-stream") {
				return nil, errors.Errorf("invalid content-type: %s", contentTypes)
			}
			return res, nil
		},
	)
	yamuxSession, err := yamux.Server(duplex, nil)
	if err != nil {
		return err
	}
	for {
		yamuxStream, err := yamuxSession.Accept()
		if err != nil {
			return err
		}
		go func() {
			// (base: https://gist.github.com/jpillora/b480fde82bff51a06238)
			// Before use, a handshake must be performed on the incoming net.Conn.
			sshConn, chans, reqs, err := ssh.NewServerConn(yamuxStream, sshConfig)
			if err != nil {
				log.Printf("Failed to handshake (%s)", err)
				return
			}
			log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
			// Discard all global out-of-band Requests
			go ssh.DiscardRequests(reqs)
			// Accept all channels
			ssh_server.HandleChannels(sshShell, chans)
		}()
	}
}
