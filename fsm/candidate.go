package fsm

type Candidate struct {
	*Server // embed server
}

func NewCandidate(s *Server, config Config) *Candidate {
	return &Candidate{s}
}

func (p *Candidate) implStateInterface() {
	var _ State = &Candidate{}
}

func (p *Candidate) enterState() {

}

func (p *Candidate) onAppendEntries(param AppendEntries) Response {
	return Response{}
}

func (p *Candidate) onRequestVote(param RequestVote) Response {
	return Response{}
}

func (p *Candidate) timeout() {

}
