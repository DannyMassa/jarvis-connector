package gerrit

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
)

// NewServer creates a Gerrit ServerImpl for the given URL.
func NewServer(u url.URL) ServerImpl {
	g := ServerImpl{
		URL: u,
	}

	g.Client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return nil
	}

	return g
}

// NewBasicAuth creates a BasicAuth authenticator. |who| should be a "user:secret" string.
func NewBasicAuth(who string) BasicAuthImpl {
	auth := strings.TrimSpace(who)
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(auth)))
	base64.StdEncoding.Encode(encoded, []byte(auth))
	return BasicAuthImpl{
		EncodedBasicAuth: string(encoded),
	}
}
