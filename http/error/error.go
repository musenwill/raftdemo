package error

type HttpError struct {
	httpStatusCode int
	Code           int    `json:"code"`
	Err            string `json:"error"`
}

func (e *HttpError) GetStatusCode() int {
	return e.httpStatusCode
}
