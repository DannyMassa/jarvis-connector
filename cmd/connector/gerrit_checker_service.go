package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/att-comdev/jarvis-connector/gerrit"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type GerritCheckerServiceInterface interface {
	ListCheckers() ([]*gerrit.CheckerInfo, error)
	PostChecker(repo, prefix string, update bool, blocking bool) (*gerrit.CheckerInfo, error)
	Serve()
}

// gerritChecker run formatting checks against a gerrit server.
type GerritCheckerServiceImpl struct {
	Server gerrit.ServerServiceInterface
	todo   chan *gerrit.PendingChecksInfo
}

// ListCheckers returns all the checkers for our scheme.
func (gc *GerritCheckerServiceImpl) ListCheckers() ([]*gerrit.CheckerInfo, error) {
	c, err := gc.Server.GetPath("a/plugins/checks/checkers/")
	if err != nil {
		log.Fatalf("ListCheckers: %v", err)
	}

	var out []*gerrit.CheckerInfo
	if err := gerrit.Unmarshal(c, &out); err != nil {
		return nil, err
	}

	filtered := out[:0]
	for _, o := range out {
		if !strings.HasPrefix(o.UUID, checkerScheme+":") {
			continue
		}
		if _, ok := checkerPrefix(o.UUID); !ok {
			continue
		}

		filtered = append(filtered, o)
	}
	return filtered, nil
}

// PostChecker creates or changes a checker. It sets up a checker on
// the given repo, for the given prefix.
func (gc *GerritCheckerServiceImpl) PostChecker(repo, prefix string, update bool, blocking bool) (*gerrit.CheckerInfo, error) {
	hash := sha1.New()       //nolint
	hash.Write([]byte(repo)) //nolint
	var blockingList []string

	// If the blocking flag is set to true, register the checker as a blocking checker
	if blocking {
		blockingList = append(blockingList, "STATE_NOT_PASSING")
	}

	uuid := fmt.Sprintf("%s:%s-%x", checkerScheme, prefix, hash.Sum(nil))
	in := gerrit.CheckerInput{
		UUID:        uuid,
		Name:        prefix,
		Description: "NewServer Checker that blocks.",
		URL:         "",
		Repository:  repo,
		Status:      "ENABLED",
		Blocking:    blockingList,
		Query:       "status:open",
	}

	body, err := json.Marshal(&in)
	log.Printf("body: %v", body)
	if err != nil {
		return nil, err
	}

	path := "a/plugins/checks/checkers/"
	if update {
		path += uuid
	}
	content, err := gc.Server.PostPath(path, "application/json", body)
	if err != nil {
		return nil, err
	}

	out := gerrit.CheckerInfo{}
	if err := gerrit.Unmarshal(content, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// Serve runs the serve loop, dispatching for checks that need it.
func (gc *GerritCheckerServiceImpl) Serve() {
	for p := range gc.todo {
		// TODO: parallelism?.
		if err := gc.executeCheck(p); err != nil {
			log.Printf("executeCheck(%v): %v", p, err)
		}
	}
}

// checkChange checks a (change, patchset) for correct formatting in the given prefix. It returns
// a list of complaints, or the errIrrelevant error if there is nothing to do.
func (gc *GerritCheckerServiceImpl) checkChange(uuid string, repository string, changeID string, psID int, prefix string) ([]string, string, error) {
	log.Printf("checkChange(%s, %d, %q)", changeID, psID, prefix)

	data := TektonListenerPayload{
		RepoRoot:       GerritURL,
		Project:        repository,
		ChangeNumber:   changeID,
		PatchSetNumber: psID,
		CheckerUUID:    uuid,
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
	req.Header.Set("X-Jarvis", "create")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var msgs []string
	msgs = append(msgs, fmt.Sprintf("%s", "Job has been submitted to tekton")) //nolint
	var details string //nolint
	details = ""
	return msgs, details, nil
}

// executeCheck executes the pending checks specified in the argument.
func (gc *GerritCheckerServiceImpl) executeCheck(pc *gerrit.PendingChecksInfo) error {
	log.Println("checking", pc)

	repository := pc.PatchSet.Repository
	changeID := strconv.Itoa(pc.PatchSet.ChangeNumber)
	psID := pc.PatchSet.PatchSetID
	for uuid := range pc.PendingChecks {
		now := gerrit.Timestamp(time.Now())
		checkInput := gerrit.CheckInput{
			CheckerUUID: uuid,
			State:       statusRunning.String(),
			Message:     "Jarvis about to submit job to tekton",
			Started:     &now,
		}
		log.Printf("posted %s", &checkInput)
		_, err := gc.Server.PostCheck(
			changeID, psID, &checkInput)
		if err != nil {
			return err
		}

		var status StatusService
		msg := ""
		url := ""
		lang, ok := checkerPrefix(uuid)
		if !ok {
			return fmt.Errorf("uuid %q had unknown prefix", uuid)
		} else {
			msgs, details, err := gc.checkChange(uuid, repository, changeID, psID, lang)
			if err == errIrrelevant {
				status = statusIrrelevant
			} else if err != nil {
				status = statusFail
				log.Printf("failed in attempt to schedule checkChange(%s, %s, %d, %q): %v", uuid, changeID, psID, lang, err)
			} else if len(msgs) != 0 {
				status = statusSuccessful
			} else {
				status = statusFail
				log.Printf("message empty for checkChange(%s, %s, %d, %q): %v", uuid, changeID, psID, lang, err)
			}
			url = details
			msg = strings.Join(msgs, ", ")
			if len(msg) > 1000 {
				msg = msg[:995] + "..."
			}
		}

		log.Printf("status %s for lang %s on %v", status, lang, pc.PatchSet)
		checkInput = gerrit.CheckInput{
			CheckerUUID: uuid,
			State:       status.String(),
			Message:     msg,
			URL:         url,
			Started:     &gerrit.Timestamp{},
		}
		log.Printf("posted %s", &checkInput)

		if _, err := gc.Server.PostCheck(changeID, psID, &checkInput); err != nil {
			return err
		}
	}
	return nil
}

// pendingLoop periodically contacts gerrit to find new checks to
// execute. It should be executed in a goroutine.
func (gc *GerritCheckerServiceImpl) pendingLoop() {
	for {
		// TODO: real rate limiting.
		time.Sleep(10 * time.Second)

		pending, err := gc.Server.PendingChecksByScheme(checkerScheme)
		if err != nil {
			log.Printf("PendingChecksByScheme: %v", err)
			return
		}

		err = gc.Server.HandleSubmissions()
		if err != nil {
			log.Printf("HandleSubmissions: %v", err)
			return
		}

		if len(pending) == 0 {
			log.Printf("no pending checks")
		}

		for _, pc := range pending {
			select {
			case gc.todo <- pc:
			default:
				log.Println("too busy; dropping pending check.")
			}
		}
	}
}
