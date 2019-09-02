package fsm

type Node struct {
	ID string
}

type Config struct {
	Nodes   []Node
	Timeout int // append entries timeout, ms
}
