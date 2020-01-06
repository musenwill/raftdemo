package error

type HttpError struct {
	httpStatusCode int
	Code           int    `json:"code"`
	Err            string `json:"error"`
}

func (p *HttpError) GetStatusCode() int {
	return p.httpStatusCode
}
