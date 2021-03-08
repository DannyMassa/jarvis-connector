package gerrit

import (
	"net/http"
)

type Authenticator interface {
	// Authenticate adds an authentication header to an outgoing request.
	Authenticate(req *http.Request) error
}

// BasicAuth adds the "Basic Authorization" header to an outgoing request.
type BasicAuthImpl struct {
	// Base64 encoded user:secret string.
	EncodedBasicAuth string
}

func (b BasicAuthImpl) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Basic "+ b.EncodedBasicAuth)
	return nil
}
