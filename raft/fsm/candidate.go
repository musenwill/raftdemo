package fsm

import (
	"math/rand"
	"time"

	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/raft"
	"go.uber.org/zap"
)

type Candidate struct {
	node raft.NodeInstance
	cfg  *raft.Config

	leaving chan bool
	logger  *zap.Logger
}

func NewCandidate(node raft.NodeInstance, cfg *raft.Config, logger *zap.Logger) *Candidate {
	rand.Seed(time.Now().UnixNano())
	return &Candidate{
		node:    node,
		cfg:     cfg,
		leaving: make(chan bool),
		logger:  logger.With(zap.String("state", "follower")),
	}
}

func (c *Candidate) Enter() {
	c.node.SetVoteFor(c.node.GetNodeID())

	go func() {
		rndTime := rand.Int63n(c.cfg.ElectionRandom.Nanoseconds())
		ticker := time.NewTicker(time.Duration(rndTime))
		for {
			select {
			case <-c.leaving:
				return
			case <-ticker.C:
				rndTime := rand.Int63n(c.cfg.ElectionRandom.Nanoseconds())
				ticker.Reset(time.Duration(rndTime) + c.cfg.CampaignTimeout)
				c.OnTimeout()
			}
		}
	}()
}

func (c *Candidate) Leave() {
	close(c.leaving)
}

func (c *Candidate) OnAppendEntries(param model.AppendEntries) model.Response {
	if c.node.CASTerm(param.Term) < 0 {
		return model.Response{Term: c.node.GetTerm(), Success: false}
	}

	c.node.SwitchStateTo(model.StateRole_Follower)
	return c.node.OnAppendEntries(param)
}

func (c *Candidate) OnRequestVote(param model.RequestVote) model.Response {
	if c.node.CASTerm(param.Term) <= 0 {
		return model.Response{Term: c.node.GetTerm(), Success: false}
	}

	c.node.SwitchStateTo(model.StateRole_Follower)
	return c.node.OnRequestVote(param)
}

func (c *Candidate) OnTimeout() {
	c.node.IncreaseTerm()
	votesC := make(chan struct{})

	go func() {
		count := 0
		if count >= int(c.cfg.Nodes)/2 {
			c.node.SwitchStateTo(model.StateRole_Leader)
			return
		}
		for {
			select {
			case <-c.leaving:
				return
			case <-votesC:
				count++
				if count >= int(c.cfg.Nodes)/2 {
					c.node.SwitchStateTo(model.StateRole_Leader)
					return
				}
			}
		}
	}()

	term := c.node.GetTerm()
	lastEntry := c.node.GetLastEntry()
	request := &model.RequestVote{
		Term:         c.node.GetTerm(),
		LastLogIndex: lastEntry.Id,
		LastLogTerm:  lastEntry.Term,
		CandidateID:  c.node.GetNodeID(),
	}

	getRequestF := func(nodeID string) (interface{}, error) {
		return *request, nil
	}

	handleResponseF := func(nodeID string, response model.Response) {
		if response.Term > term {
			c.node.SwitchStateTo(model.StateRole_Follower)
		}
		if response.Success {
			c.logger.Info("receive vote", zap.String("voter", nodeID))
			votesC <- struct{}{}
		}
	}

	c.node.Broadcast("election campaign", c.leaving, getRequestF, handleResponseF)
}

func (c *Candidate) State() model.StateRole {
	return model.StateRole_Candidate
}
