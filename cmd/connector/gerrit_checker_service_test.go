package main_test

import (
	"encoding/json"
	"github.com/att-comdev/jarvis-connector/cmd/connector"
	"github.com/att-comdev/jarvis-connector/gerrit"
	"net/http"
	"net/url"
	"testing"
)

type gerritCheckerServiceMock struct {
	listCheckersFn func() ([]*gerrit.CheckerInfo, error)
	postCheckerFn func(repo, prefix string, update bool, blocking bool) (*gerrit.CheckerInfo, error)
	serveFn func()
}

func (mock gerritCheckerServiceMock) ListCheckers() ([]*gerrit.CheckerInfo, error) {
	return mock.listCheckersFn()
}

func (mock gerritCheckerServiceMock) PostChecker(repo, prefix string, update bool, blocking bool) (*gerrit.CheckerInfo, error) {
	return mock.postCheckerFn(repo, prefix, update, blocking)
}

func (mock gerritCheckerServiceMock) Serve() {}

type serverMock struct {
	GetPathFn func(p string) ([]byte, error)
	DoFn func(req *http.Request) (*http.Response, error)
	GetFn func(u *url.URL) ([]byte, error)
	PostPathFn func(p string, contentType string, content []byte) ([]byte, error)
	PendingChecksBySchemeFn func(scheme string) ([]*gerrit.PendingChecksInfo, error)
	PendingChecksFn func(checkerUUID string) ([]*gerrit.PendingChecksInfo, error)
	PostCheckFn func(changeID string, psID int, input *gerrit.CheckInput) (*gerrit.CheckInfo, error)
	HandleSubmissionsFn func() error
}

func (mock serverMock) GetPath(p string) ([]byte, error) {
	return mock.GetPathFn(p)
}

func (mock serverMock) Do(req *http.Request) (*http.Response, error) {
	return mock.DoFn(req)
}

func (mock serverMock) Get(u *url.URL) ([]byte, error) {
	return mock.GetFn(u)
}

func (mock serverMock) PostPath(p string, contentType string, content []byte) ([]byte, error) {
	return mock.PostPathFn(p, contentType, content)
}

func (mock serverMock) PendingChecksByScheme(checkerUUID string) ([]*gerrit.PendingChecksInfo, error) {
	return mock.PendingChecksBySchemeFn(checkerUUID)
}

func (mock serverMock) PendingChecks(checkerUUID string) ([]*gerrit.PendingChecksInfo, error) {
	return mock.PendingChecksFn(checkerUUID)
}

func (mock serverMock) PostCheck(changeID string, psID int, input *gerrit.CheckInput) (*gerrit.CheckInfo, error) {
	return mock.PostCheckFn(changeID, psID, input)
}

func (mock serverMock) HandleSubmissions() error {
	return mock.HandleSubmissionsFn()
}

func TestGerritCheckerServiceMockability(t *testing.T) {
	serviceMock := gerritCheckerServiceMock{}

	serviceMock.listCheckersFn = func() ([]*gerrit.CheckerInfo, error) {
		return nil, nil
	}
	serviceMock.postCheckerFn = func(repo, prefix string, update bool, blocking bool) (*gerrit.CheckerInfo, error) {
		return nil, nil
	}
	serviceMock.serveFn = func() {}

	l1, l2 := serviceMock.ListCheckers()
	l3, l4 := serviceMock.PostChecker("repo", "prefix", false, false)
	// If not mocked, this will not terminate
	serviceMock.Serve()

	if l1 != nil || l2 != nil || l3 != nil || l4 != nil {
		t.Error("mock was not called, expected mocked service to return nil error")
	}
}

func TestListCheckers(t *testing.T) {
	// Build desired response
	checkerReturnString := ")]}'" +
		"[{" +
		"\"uuid\":\"jarvis:jarvispipeline-c3dfad99656b3d59b2c7ef8bc044cb4ce9f534b8\"," +
		"\"name\":\"jarvispipeline\"," +
		"\"description\":\"check source code formatting.\"," +
		"\"repository\":\"ausf\"," +
		"\"status\":\"ENABLED\"," +
		"\"blocking\":[]," +
		"\"query\":\"status:open\"," +
		"\"created\":\"2021-02-23 16:54:54.000000000\"," +
		"\"updated\":\"2021-02-23 16:54:54.000000000\"" +
	"}]\n"
	var expected []*gerrit.CheckerInfo
	err := json.Unmarshal([]byte(checkerReturnString[4:]), &expected)

	if err != nil {
		t.Errorf("Error setting up test: %v", err)
	}

	// Create a Gerrit Checker Service with a mocked out Server
	serverServiceMock := &serverMock{}
	serverServiceMock.GetPathFn = func (p string) ([]byte, error) {
		return []byte(checkerReturnString), nil
	}
	gerrit.ServiceService = serverServiceMock
	gc, _ := main.NewGerritChecker(serverServiceMock)
	gc.Server = serverServiceMock

	// run function under test
	actual, err := gc.ListCheckers()

	// Asserts
	if err != nil {
		t.Errorf("ListCheckers method returned error: %v", err)
	}

	if expected[0] != nil && actual[0] != nil && expected[0].UUID != actual[0].UUID { //TODO DeepEquals?
		t.Errorf("ListCheckers did not return expected checker \nexpected: %v, \nactual: %v", expected, actual)
	}
}

func TestPostChecker(t *testing.T) {
	// build desired result
	checkerReturnString := ")]}'" + "{" +
		"\"uuid\":\"this-value-is-generated-in-the-method\"," +
		"\"name\":\"jarvis-pipeline\"," +
		"\"description\":\"mic check one two.\"," +
		"\"repository\":\"ausf\"," +
		"\"status\":\"ENABLED\"," +
		"\"blocking\":[]," +
		"\"query\":\"status:open\"" +
	"}"
	var expected *gerrit.CheckerInfo
	err := json.Unmarshal([]byte(checkerReturnString[4:]), &expected)

	if err != nil {
		t.Errorf("Error setting up test: %v", err)
	}
}

func TestServe(t *testing.T) {

}
