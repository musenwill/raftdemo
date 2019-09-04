package config

type Node struct {
	ID string
}

type Config struct {
	Nodes        []Node
	Timeout      int // append entries timeout, ms
	MaxReplicate int
}

func (p *Config) Len() int {
	return len(p.Nodes)
}
