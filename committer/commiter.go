package committer

import (
	"io"
	"os"

	"github.com/musenwill/raftdemo/model"
)

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

func (p *FileCommitter) Commit(log model.Log) error {
	_, err := p.writer.Write([]byte(log.Command))
	return err
}
