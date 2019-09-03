package fsm

type nodeIndex struct {
	nextIndex, matchIndex int64
}

type Leader struct {
	*Server // embed server
	nIndex  map[string]nodeIndex
}

func NewLeader(s *Server, config *Config) *Leader {
	nIndex := make(map[string]nodeIndex)
	for _, n := range config.Nodes {
		nIndex[n.ID] = nodeIndex{}
	}
	return &Leader{s, nIndex}
}

func (p *Leader) implStateInterface() {
	var _ State = &Leader{}
}

func (p *Leader) enterState() {

}

func (p *Leader) onAppendEntries(param AppendEntries) Response {
	return Response{}
}

func (p *Leader) onRequestVote(param RequestVote) Response {
	return Response{}
}

func (p *Leader) timeout() {

}
