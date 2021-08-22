package committer

type ImmCommiter struct {
	data []byte
}

func NewImmCommitter() *ImmCommiter {
	return &ImmCommiter{
		data: make([]byte, 0),
	}
}

func (c *ImmCommiter) implCommitterInterface() {
	var _ Committer = &ImmCommiter{}
}

func (c *ImmCommiter) Commit(data []byte) error {
	c.data = append(c.data, data...)
	return nil
}

func (c *ImmCommiter) Data() ([]byte, error) {
	data := make([]byte, len(c.data))
	copy(data, c.data)
	return data, nil
}

func (c *ImmCommiter) Close() error {
	return nil
}
