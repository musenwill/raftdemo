package proxy

type Proxy interface {
	AppendEntriesRequestSender(nodeID string) (chan<- AppendEntries, error)

	AppendEntriesRequestReader(nodeID string) (<-chan AppendEntries, error)

	AppendEntriesResponseSender(nodeID string) (chan<- Response, error)

	AppendEntriesResponseReader(nodeID string) (<-chan Response, error)

	VoteRequestSender(nodeID string) (chan<- RequestVote, error)

	VoteRequestReader(nodeID string) (<-chan RequestVote, error)

	VoteResponseSender(nodeID string) (chan<- Response, error)

	VoteResponseReader(nodeID string) (<-chan Response, error)
}
