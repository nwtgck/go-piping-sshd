package cmd

import (
	"fmt"
	"github.com/nwtgck/go-piping-sshd/piping_util"
	"github.com/pkg/errors"
)

const (
	yamuxFlagLongName = "yamux"
)

const yamuxMimeType = "application/yamux"

func generatePaths(args []string) (string, string, error) {
	var clientToServerPath string
	var serverToClientPath string

	switch len(args) {
	case 1:
		// NOTE: "cs": from client-host to server-host
		clientToServerPath = fmt.Sprintf("%s/cs", args[0])
		// NOTE: "sc": from server-host to client-host
		serverToClientPath = fmt.Sprintf("%s/sc", args[0])
	case 2:
		clientToServerPath = args[0]
		serverToClientPath = args[1]
	default:
		return "", "", errors.New("the number of paths should be one or two")
	}
	return clientToServerPath, serverToClientPath, nil
}

func headersWithYamux(headers []piping_util.KeyValue) []piping_util.KeyValue {
	return append(headers, piping_util.KeyValue{Key: "Content-Type", Value: yamuxMimeType})
}
