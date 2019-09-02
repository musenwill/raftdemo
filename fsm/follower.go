package fsm

type Follower struct {
	*Server  // embed server
	leaderID string
}

func NewFollower(s *Server, config Config) *Follower {
	return &Follower{s, ""}
}

func (p *Follower) implStateInterface() {
	var _ State = &Follower{}
}

func (p *Follower) enterState() {

}

func (p *Follower) onAppendEntries(param AppendEntries) Response {
	return Response{}
}

func (p *Follower) onRequestVote(param RequestVote) Response {
	return Response{}
}

func (p *Follower) timeout() {

}
