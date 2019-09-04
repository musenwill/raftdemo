package fsm

import (
	"io"
	"os"

	"github.com/musenwill/raftdemo/proxy"
)

type Committer interface {
	Commit(log proxy.Log) error
}

type FileCommitter struct {
	filePath string
	writer   io.Writer
}

func NewFileCommitter(filePath string) (*FileCommitter, error) {
	writer, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return &FileCommitter{filePath: filePath, writer: writer}, nil
}

func (p *FileCommitter) implCommitterInterface() {
	var _ Committer = &FileCommitter{}
}

func (p *FileCommitter) Commit(log proxy.Log) error {
	_, err := p.writer.Write([]byte(log.Command))
	return err
}
