package committer

import (
	"os"
)

type FileCommitter struct {
	filePath string
	writer   *os.File
}

func NewFileCommitter(filePath string) (*FileCommitter, error) {
	writer, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &FileCommitter{filePath: filePath, writer: writer}, nil
}

func (c *FileCommitter) implCommitterInterface() {
	var _ Committer = &FileCommitter{}
}

func (c *FileCommitter) Commit(data []byte) error {
	_, err := c.writer.Write(data)
	return err
}

func (c *FileCommitter) Data() ([]byte, error) {
	stat, err := c.writer.Stat()
	if err != nil {
		return nil, err
	}
	b := make([]byte, int(stat.Size()))

	_, err = c.writer.ReadAt(b, 0)
	return b, err
}

func (c *FileCommitter) Close() error {
	return c.writer.Close()
}
