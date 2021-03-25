package util

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"time"
)

// (base: https://stackoverflow.com/a/34668130/2885946)
func UrlJoin(s string, p ...string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(append([]string{u.Path}, p...)...)
	return u.String(), nil
}

// Generate HTTP client
func CreateHttpClient(insecure bool, writeBufSize int, readBufSize int) *http.Client {
	// Set insecure or not
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: insecure},
		WriteBufferSize:   writeBufSize,
		ReadBufferSize:    readBufSize,
		ForceAttemptHTTP2: true,
	}
	return &http.Client{Transport: tr}
}

// Set default resolver for HTTP client
func CreateDialContext(dnsServer string) func(ctx context.Context, network, address string) (net.Conn, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, "udp", dnsServer)
		},
	}

	// Resolver for HTTP
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{
			Timeout:  time.Millisecond * time.Duration(10000),
			Resolver: resolver,
		}
		return d.DialContext(ctx, network, address)
	}
}

type combinedError struct {
	e1 error
	e2 error
}

func (e combinedError) Error() string {
	return fmt.Sprintf("%v and %v", e.e1, e.e2)
}

func CombineErrors(e1 error, e2 error) error {
	if e1 == nil {
		return e2
	}
	if e2 == nil {
		return e1
	}
	return &combinedError{e1: e1, e2: e2}
}
