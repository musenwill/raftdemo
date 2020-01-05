package rest

import "errors"

type AddNodeForm struct {
	Host string `json:"host"`
}

type AddLogForm struct {
	RequestID string `json:"request_id"`
	Command   string `json:"command"`
}

type UpdateNodeForm struct {
	Timeout bool `json:"timeout"`
	Sleep   bool `json:"sleep"`
}

func (p *UpdateNodeForm) Valid() error {
	var filedCount int
	if p.Timeout {
		filedCount++
	}
	if p.Sleep {
		filedCount++
	}
	if filedCount > 1 {
		return errors.New("can not specify one field once")
	}
	return nil
}
