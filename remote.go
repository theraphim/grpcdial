package grpcdial

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func Conn(ctx context.Context, remote string, popts ...grpc.DialOption) (*grpc.ClientConn, error) {
	remote = strings.TrimSpace(remote)
	if remote == "" {
		return nil, errNoURL{}
	}
	do, err := parseURL(remote)
	if err != nil {
		_, port, err := net.SplitHostPort(remote)
		if err != nil {
			return nil, err
		}
		if port == "443" {
			do.tls = true
		}
		do.remote = remote
	}
	var opts []grpc.DialOption
	if do.tls {
		opts = []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))}
	} else {
		opts = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	}
	if len(popts) != 0 {
		opts = append(opts, popts...)
	}
	return grpc.NewClient(do.remote, opts...)
}

type errNoURL struct{}

func (errNoURL) Error() string { return "not an url" }

type dialOpts struct {
	remote string
	tls    bool
}

func parseURL(remote string) (result dialOpts, err error) {
	destURL, err := url.Parse(remote)
	if err != nil {
		return result, err
	}

	switch destURL.Scheme {
	case "unix":
		result.remote = remote
	case "https":
		port := destURL.Port()
		if port == "" {
			port = "443"
		}
		result.remote = destURL.Hostname() + ":" + port
		result.tls = true
	case "http":
		port := destURL.Port()
		if port == "" {
			port = "80"
		}
		result.remote = destURL.Hostname() + ":" + port
	default:
		return result, errNoURL{}
	}
	return result, nil
}

type simpleMapOption struct {
	values map[string]string
}

func (s simpleMapOption) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	if len(s.values) == 0 {
		return nil, nil
	}
	return s.values, nil
}

func (s simpleMapOption) RequireTransportSecurity() bool {
	return false
}

func WithMapOption(value map[string]string) credentials.PerRPCCredentials {
	return simpleMapOption{value}
}

func WithAccessOption(access string) credentials.PerRPCCredentials {
	if access == "" {
		return nil
	}
	return WithMapOption(map[string]string{
		"access": access,
	})
}

func MaybeWithAccessTokenOptions(accessToken string, opts ...grpc.DialOption) []grpc.DialOption {
	if accessToken != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(WithAccessOption(accessToken)))
	}
	return opts
}
