package fsm

import (
	"math/rand"
	"time"

	"github.com/musenwill/raftdemo/log"
	"github.com/musenwill/raftdemo/model"
	"github.com/musenwill/raftdemo/proxy"
	"github.com/musenwill/raftdemo/raft"
	"go.uber.org/zap"
)

type Candidate struct {
	node raft.NodeInstance
	cfg  *raft.Config

	leaving chan bool

	nodes []string
	proxy proxy.Proxy

	logger log.Logger
}

func NewCandidate(node raft.NodeInstance, nodes []string, proxy proxy.Proxy, cfg *raft.Config, logger log.Logger) *Candidate {
	rand.Seed(time.Now().UnixNano())
	return &Candidate{
		node:    node,
		nodes:   nodes,
		proxy:   proxy,
		cfg:     cfg,
		leaving: make(chan bool),
		logger:  *logger.With(zap.String("state", "follower")),
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
	if c.node.CompareAndSetTerm(param.Term) < 0 {
		return model.Response{Term: c.node.GetTerm(), Success: false}
	}

	c.node.SwitchStateTo(model.StateRole_Follower)
	return c.node.OnAppendEntries(param)
}

func (c *Candidate) OnRequestVote(param model.RequestVote) model.Response {
	if c.node.CompareAndSetTerm(param.Term) <= 0 {
		return model.Response{Term: c.node.GetTerm(), Success: false}
	}

	c.node.SwitchStateTo(model.StateRole_Follower)
	return c.node.OnRequestVote(param)
}

func (c *Candidate) OnTimeout() {
	votesC := make(chan struct{})

	go func() {
		count := 0
		for {
			select {
			case <-c.leaving:
				return
			case <-votesC:
				count++
				if count >= len(c.nodes)/2 {
					c.node.SwitchStateTo(model.StateRole_Leader)
					return
				}
			}
		}
	}()

	go func() {
		for _, n := range c.nodes {
			nodeID := n
			select {
			case <-c.leaving:
				break
			default:
			}
			go func() {
				response, err := c.proxy.Send(nodeID, model.RequestVote{
					Term:         c.node.GetTerm(),
					LastLogIndex: c.node.GetLastLogIndex(),
					LastLogTerm:  c.node.GetLastLogTerm(),
					CandidateID:  c.node.GetNodeID(),
				}, c.leaving)
				if err != nil {
					c.logger.Errorf("append entries, %w", err)
					return
				}
				if response.Term > c.node.GetTerm() {
					c.node.SwitchStateTo(model.StateRole_Follower)
				}
				if response.Success {
					c.logger.Info("receive vote", zap.String("voter", nodeID))
					votesC <- struct{}{}
				}
			}()
		}
	}()
}

func (c *Candidate) State() model.StateRole {
	return model.StateRole_Candidate
}
