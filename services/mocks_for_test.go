package services_test

import (
	"github.com/att-comdev/jarvis-connector/services"
	"github.com/att-comdev/jarvis-connector/types"
	"net/url"
)

type serverServiceMock struct {
	postPathFn    func(pathing string, headers []types.Header, content []byte) ([]byte, error)
	getPathFn     func(pathing string, headers []types.Header) ([]byte, error)
	getFn         func(u *url.URL) ([]byte, error)
	initFn        func(url url.URL, authenticator services.Authenticator, testPath string)
	getURLFn      func() url.URL
	getRepoRootFn func() string
}

func (s serverServiceMock) GetPath(pathing string, headers []types.Header) ([]byte, error) {
	return s.getPathFn(pathing, headers)
}

func (s serverServiceMock) PostPath(pathing string, headers []types.Header, content []byte) ([]byte, error) {
	return s.postPathFn(pathing, headers, content)
}

func (s serverServiceMock) Get(u *url.URL) ([]byte, error) {
	return s.getFn(u)
}

func (s serverServiceMock) Init(url url.URL, authenticator services.Authenticator, testPath string) {
	s.initFn(url, authenticator, testPath)
}

func (s serverServiceMock) GetURL() url.URL {
	return s.getURLFn()
}

func (s serverServiceMock) GetRepoRoot() string {
	return s.getRepoRootFn()
}