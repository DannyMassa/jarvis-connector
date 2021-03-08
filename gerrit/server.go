// Copyright 2019 Google Inc. All rights reserved.
// Copyright 2021 AT&T Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gerrit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
)

const (
	JarvisMergeHashtag string = "jarvis-merge"
)

type ServerServiceInterface interface {
	GetPath(p string) ([]byte, error)
	Do(req *http.Request) (*http.Response, error)
	Get(u *url.URL) ([]byte, error)
	PostPath(p string, contentType string, content []byte) ([]byte, error)
	PendingChecksByScheme(scheme string) ([]*PendingChecksInfo, error)
	PendingChecks(checkerUUID string) ([]*PendingChecksInfo, error)
	PostCheck(changeID string, psID int, input *CheckInput) (*CheckInfo, error)
	HandleSubmissions() error
}

// ServerImpl represents a single Gerrit host.
type ServerImpl struct {
	UserAgent string
	URL       url.URL
	Client    http.Client

	// Issue trace requests.
	Debug bool

	Authenticator Authenticator
}

var (
	ServiceService ServerServiceInterface = ServerImpl{}
)

// GetPath runs a Get on the given path.
func (service ServerImpl) GetPath(p string) ([]byte, error) {
	u := service.URL
	u.Path = path.Join(u.Path, p)
	if strings.HasSuffix(p, "/") && !strings.HasSuffix(u.Path, "/") {
		// Ugh.
		u.Path += "/"
	}
	return service.Get(&u)
}

// Do runs a HTTP request against the remote server.
func (service ServerImpl) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", service.UserAgent)
	if service.Authenticator != nil {
		if err := service.Authenticator.Authenticate(req); err != nil {
			return nil, err
		}
	}

	if service.Debug {
		if req.URL.RawQuery != "" {
			req.URL.RawQuery += "&trace=0x1"
		} else {
			req.URL.RawQuery += "trace=0x1"
		}
	}
	return service.Client.Do(req)
}

// Get runs a HTTP GET request on the given URL.
func (service ServerImpl) Get(u *url.URL) ([]byte, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	rep, err := service.Do(req)
	if err != nil {
		return nil, err
	}
	if rep.StatusCode/100 != 2 {
		return nil, fmt.Errorf("get %s: status %d", u.String(), rep.StatusCode)
	}

	defer rep.Body.Close()
	return ioutil.ReadAll(rep.Body)
}

// PostPath posts the given data onto a path.
func (service ServerImpl) PostPath(p string, contentType string, content []byte) ([]byte, error) {
	u := service.URL
	u.Path = path.Join(u.Path, p)
	if strings.HasSuffix(p, "/") && !strings.HasSuffix(u.Path, "/") {
		// Ugh.
		u.Path += "/"
	}
	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	rep, err := service.Do(req)
	if err != nil {
		return nil, err
	}
	if rep.StatusCode/100 != 2 {
		return nil, fmt.Errorf("post %s service: status %d", u.String(), rep.StatusCode)
	}

	defer rep.Body.Close()
	return ioutil.ReadAll(rep.Body)
}

// PendingChecksByScheme queries Gerrit
func (service ServerImpl) PendingChecksByScheme(scheme string) ([]*PendingChecksInfo, error) {
	u := service.URL

	// The trailing '/' handling is really annoying.
	u.Path = path.Join(u.Path, "a/plugins/checks/checks.pending/") + "/"

	q := "scheme:" + scheme
	u.RawQuery = "query=" + q
	content, err := service.Get(&u)
	if err != nil {
		return nil, err
	}

	var out []*PendingChecksInfo
	if err := Unmarshal(content, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// PendingChecks returns the checks pending for the given checker.
func (service ServerImpl) PendingChecks(checkerUUID string) ([]*PendingChecksInfo, error) {
	u := service.URL

	// The trailing '/' handling is really annoying.
	u.Path = path.Join(u.Path, "a/plugins/checks/checks.pending/") + "/"

	q := "checker:" + checkerUUID
	u.RawQuery = "query=" + url.QueryEscape(q)

	content, err := service.Get(&u)
	if err != nil {
		return nil, err
	}

	var out []*PendingChecksInfo
	if err := Unmarshal(content, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// PostCheck posts a single check result onto a change.
func (service ServerImpl) PostCheck(changeID string, psID int, input *CheckInput) (*CheckInfo, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	res, err := service.PostPath(fmt.Sprintf("a/changes/%service/revisions/%d/checks/", changeID, psID),
		"application/json", body)
	if err != nil {
		return nil, err
	}

	var out CheckInfo
	if err := Unmarshal(res, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (service ServerImpl) HandleSubmissions() error {
	u := service.URL

	u.Path = path.Join(u.Path, "a/changes/?o=SUBMITTABLE&q=is:open/") + "/"
	content, err := service.Get(&u)
	if err != nil {
		return err
	}

	var out []*PendingSubmitInfo
	if err := Unmarshal(content, &out); err != nil {
		return err
	}

	if err = service.FilterPatchsets(out); err != nil {
		return err
	}

	for _, obj := range out {
		if err = service.PostHashtag(obj); err != nil {
			return err
		}

		if err = service.CallPipeline(obj); err != nil {
			return err
		}
	}


	return nil
}

type HashtagPayload struct {
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
}

func (service ServerImpl) PostHashtag(patchset *PendingSubmitInfo) error {
	u := service.URL
	u.Path = path.Join(u.Path, fmt.Sprintf("a/changes/%s/hashtags", patchset.ChangeId)) + "/"
	hashtagPayload := HashtagPayload{
		Add: []string{JarvisMergeHashtag},
		Remove: []string{},
	}
	body, err := json.Marshal(hashtagPayload)
	if err != nil {
		return err
	}
	_, err = service.PostPath(u.Path, "application/json", body)
	if err != nil {
		return err
	}

	return nil
}

func (service ServerImpl) FilterPatchsets(patchsets []*PendingSubmitInfo) error {
	for i, obj := range patchsets {
		// Ignore merge conflicts, patchsets without required labels, and patchsets currently being handled by Jarvis
		if obj.Mergeable == false || obj.Subittable == false || contains(obj.Hashtags, JarvisMergeHashtag) {
			patchsets = append(patchsets[:i], patchsets[i+1:]...)
		}
	}
	return nil
}

type TektonMergePayload struct {
	RepoRoot       string `json:"repoRoot"`
	Project        string `json:"project"`
	ChangeNumber   string `json:"changeNumber"`
}

func (service ServerImpl) CallPipeline(patchset *PendingSubmitInfo) error {
	EventListenerURL := "http://el-jarvis-system.jarvis-system.svc.cluster.local:8080/"
	data := TektonMergePayload{
		RepoRoot:       "http://gerrit.jarvis.local/",
		Project:        patchset.Project,
		ChangeNumber:   patchset.ChangeId,
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest("POST", EventListenerURL, body)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Jarvis", "merge")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	return nil
}

func contains(list []string, element string) bool {
	for _, obj := range list {
		if obj == element {
			return true
		}
	}

	return false
}
