package grpcdial

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func Conn(ctx context.Context, remote string, popts ...grpc.DialOption) (*grpc.ClientConn, error) {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return nil, errNoURL{}
	}
	do, err := parseURL(remote)
	if err != nil {
		do.host, do.port, err = net.SplitHostPort(remote)
		if err != nil {
			return nil, err
		}
		if do.port == "443" {
			do.tls = true
		}
	}
	var opts []grpc.DialOption
	if do.tls {
		opts = []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))}
	} else {
		opts = []grpc.DialOption{grpc.WithInsecure()}
	}
	if len(popts) != 0 {
		opts = append(opts, popts...)
	}
	muxb := do.host + ":" + do.port
	return grpc.DialContext(ctx, muxb, opts...)
}

type errNoURL struct{}

func (errNoURL) Error() string { return "not an url" }

type dialOpts struct {
	host, port string
	tls        bool
}

func parseURL(remote string) (result dialOpts, err error) {
	destURL, err := url.Parse(remote)
	if err != nil {
		return result, err
	}

	result.port = destURL.Port()
	switch destURL.Scheme {
	case "https":
		if result.port == "" {
			result.port = "443"
		}
		result.tls = true
	case "http":
		if result.port == "" {
			result.port = "80"
		}
	default:
		return result, errNoURL{}
	}
	result.host = destURL.Hostname()
	return result, nil
}
