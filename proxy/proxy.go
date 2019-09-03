package proxy

import (
	"fmt"

	"github.com/musenwill/raftdemo/config"
)

type appendEntriesEndpoint struct {
	request  chan AppendEntries
	response chan Response
}

type requestVoteEndpoint struct {
	request  chan RequestVote
	response chan Response
}

type endpoint struct {
	appendEntries appendEntriesEndpoint
	requestVote   requestVoteEndpoint
}

type AppendEntriesHandler func(param AppendEntries) Response
type RequestVoteHandler func(param RequestVote) Response

var router map[string]endpoint

func Config(nodes []config.Node) {
	if router == nil {
		router = make(map[string]endpoint)
	}

	nodeSet := make(map[string]bool)
	for _, node := range nodes {
		nodeSet[node.ID] = true
	}

	// clear all none exists nodes
	for k, v := range router {
		if _, ok := nodeSet[k]; !ok {
			close(v.appendEntries.request)
			close(v.appendEntries.response)
			close(v.requestVote.request)
			close(v.requestVote.response)
			delete(router, k)
		}
	}

	// add all new joined nodes
	for k, _ := range nodeSet {
		if _, ok := router[k]; !ok {
			e := endpoint{
				appendEntries: appendEntriesEndpoint{
					request:  make(chan AppendEntries),
					response: make(chan Response),
				},
				requestVote: requestVoteEndpoint{
					request:  make(chan RequestVote),
					response: make(chan Response),
				},
			}
			router[k] = e
		}
	}
}

func SendAppendEntries(nodeID string, request AppendEntries) (Response, error) {
	e, ok := router[nodeID]
	if !ok {
		return Response{}, fmt.Errorf("node with id %v not exist", nodeID)
	}

	e.appendEntries.request <- request
	return <-e.appendEntries.response, nil
}

func SendRequestVote(nodeID string, request RequestVote) (Response, error) {
	e, ok := router[nodeID]
	if !ok {
		return Response{}, fmt.Errorf("node with id %v not exist", nodeID)
	}

	e.requestVote.request <- request
	return <-e.requestVote.response, nil
}

func HandleAppendEntries(nodeID string, h AppendEntriesHandler) error {
	e, ok := router[nodeID]
	if !ok {
		return fmt.Errorf("node with id %v not exist", nodeID)
	}
	e.appendEntries.response <- h(<-e.appendEntries.request)
	return nil
}

func HandleRequestVote(nodeID string, h RequestVoteHandler) error {
	e, ok := router[nodeID]
	if !ok {
		return fmt.Errorf("node with id %v not exist", nodeID)
	}
	e.requestVote.response <- h(<-e.requestVote.request)
	return nil
}
