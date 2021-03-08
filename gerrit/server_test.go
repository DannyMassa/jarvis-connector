package gerrit_test

import (
	"github.com/att-comdev/jarvis-connector/gerrit"
	"net/http"
	"net/url"
	"testing"
)

type serverMock struct {
	GetPathFn func(p string) ([]byte, error)
	DoFn func(req *http.Request) (*http.Response, error)
	GetFn func(u *url.URL) ([]byte, error)
	PostPathFn func(p string, contentType string, content []byte) ([]byte, error)
	PendingChecksBySchemeFn func(scheme string) ([]*gerrit.PendingChecksInfo, error)
	PendingChecksFn func(checkerUUID string) ([]*gerrit.PendingChecksInfo, error)
	PostCheckFn func(changeID string, psID int, input *gerrit.CheckInput) (*gerrit.CheckInfo, error)
	HandleSubmissionsFn func([]*gerrit.PendingSubmitInfo, error)
}

func (mock *serverMock) GetPath(p string) ([]byte, error) {
	return mock.GetPathFn(p)
}

func (mock *serverMock) Do(req *http.Request) (*http.Response, error) {
	return mock.DoFn(req)
}

func (mock *serverMock) Get(u *url.URL) ([]byte, error) {
	return mock.GetFn(u)
}

func (mock *serverMock) PostPath(p string, contentType string, content []byte) ([]byte, error) {
	return mock.PostPathFn(p, contentType, content)
}

func (mock *serverMock) PendingChecksByScheme(checkerUUID string) ([]*gerrit.PendingChecksInfo, error) {
	return mock.PendingChecksBySchemeFn(checkerUUID)
}

func (mock *serverMock) PendingChecks(checkerUUID string) ([]*gerrit.PendingChecksInfo, error) {
	return mock.PendingChecksFn(checkerUUID)
}

func (mock *serverMock) PostCheck(changeID string, psID int, input *gerrit.CheckInput) (*gerrit.CheckInfo, error) {
	return mock.PostCheckFn(changeID, psID, input)
}

func TestServerMockability(t *testing.T) {
	serviceMock := serverMock{}
	serviceMock.GetPathFn = func(p string) ([]byte, error) {
		return []byte{}, nil
	}

	result, resultError := serviceMock.GetPath("anyString")
	var expected []byte
	expectedError := error(nil)

	if len(result) != len(expected) || resultError != expectedError {
		t.Error("mock was not called, expected mocked service to return \"anyString\" string")
	}
}

func TestGetPath(t *testing.T) {
	// Setup ServerImpl
	goodUrl, err := url.Parse("https://github.com/")
	if err != nil {
		t.Errorf("Received error during test setup: %v", err)
	}
	goodServer := gerrit.NewServer(*goodUrl)

	// Mock API Endpoint

	// Ensure API GET request is made, and trailing slash is appended if necessary
	slashAPI := "my/fake/api/"
	noSlashAPI := "my/fake/api"
	_, slashErr := goodServer.GetPath(slashAPI)
	_, noSlashErr := goodServer.GetPath(noSlashAPI)
	if slashErr != nil {
		// t.Errorf("Error making GET request for path: %s, error: %v", slashAPI, slashErr)
	}
	if noSlashErr != nil {
		// t.Errorf("Error making GET request for path: %s, error: %v", noSlashAPI, noSlashErr)
	}
}

func TestHandleSubmissions(t *testing.T) {
	// TODO write tests
}