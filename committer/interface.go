package committer

type Committer interface {
	Commit(data []byte) error
	Data() ([]byte, error)
	Close() error
}
