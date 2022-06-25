package security

import (
	"context"
	"encoding/base64"
)

type BasicAuth struct {
	username string
	password string
}

func (b BasicAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	auth := b.username + ":" + b.password
	enc := base64.StdEncoding.EncodeToString([]byte(auth))
	return map[string]string{
		"authorization": "Basic " + enc,
	}, nil
}

func (b BasicAuth) RequireTransportSecurity() bool {
	return true
}

func NewBasicAuth(name, password string) *BasicAuth {
	return &BasicAuth{
		username: name,
		password: password,
	}
}
