package gerrit_test

import (
	"github.com/att-comdev/jarvis-connector/gerrit"
	"net/http"
	"testing"
)

type basicAuthServiceMock struct {
	authenticateFn func(req *http.Request) error
}

func (mock basicAuthServiceMock) Authenticate(req *http.Request) error {
	return mock.authenticateFn(req)
}

func TestServiceMockability(t *testing.T) {
	serviceMock := basicAuthServiceMock{}
	serviceMock.authenticateFn = func(req *http.Request) error {
		return nil
	}

	var request http.Request
	result := serviceMock.Authenticate(&request)
	expected := error(nil)

	if result != expected {
		t.Error("mock was not called, expected mocked service to return nil error")
	}
}

func TestNewBasicAuth(t *testing.T) {
	credentials := "Jarvis:Landry"
	paddedCredentials := " Jarvis:Landry "
	b64EncodedCredentials := "SmFydmlzOkxhbmRyeQ=="
	emptyCredentials := ""
	test1 := gerrit.NewBasicAuth(credentials)
	test2 := gerrit.NewBasicAuth(paddedCredentials)
	test3 := gerrit.NewBasicAuth(emptyCredentials)

	if test1.EncodedBasicAuth != b64EncodedCredentials {
		t.Errorf("Base64 Encoding of %s was %s, expected: %s",
			credentials, test1.EncodedBasicAuth, b64EncodedCredentials)
	}

	if test2.EncodedBasicAuth != b64EncodedCredentials {
		t.Errorf("Base64 Encoding of %s was %s, expected: %s",
			credentials, test2.EncodedBasicAuth, b64EncodedCredentials)
	}

	if test3.EncodedBasicAuth != "" {
		t.Errorf("Base64 Encoding of %s was %s, expected: empty string",
			emptyCredentials, test3.EncodedBasicAuth)
	}
}

func TestAuthenticate(t *testing.T) {
	test1 := gerrit.BasicAuthImpl{
		EncodedBasicAuth: "SmFydmlzOkxhbmRyeQ==",
	}
	request, err := http.NewRequest(http.MethodGet, "gerrit.com", nil)

	err = test1.Authenticate(request)

	if err != nil {
		t.Errorf("Received error from Authenicate function: %v", err)
	} else if request.Header.Get("Authorization") == "" {
		t.Errorf("Authorization header not found")
	} else if request.Header.Get("Authorization") != "Basic SmFydmlzOkxhbmRyeQ==" {
		t.Errorf("Authorization header was %s, expected: Basic SmFydmlzOkxhbmRyeQ==",
			request.Header.Get("Authorization"))
	}
}
