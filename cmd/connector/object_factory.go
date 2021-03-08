package main

import (
	"github.com/att-comdev/jarvis-connector/gerrit"
)

// NewGerritChecker creates a server that periodically checks a gerrit
// server for pending checks.
func NewGerritChecker(server gerrit.ServerServiceInterface) (GerritCheckerServiceImpl, error) {
	gc := GerritCheckerServiceImpl{
		Server: server,
		todo:   make(chan *gerrit.PendingChecksInfo, 5),
	}

	go gc.pendingLoop()
	return gc, nil
}
