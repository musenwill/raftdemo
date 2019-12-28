package error

func ParamError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 400,
		Code:           100,
		Err:            e.Error(),
	}
}

func FormatError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 415,
		Code:           101,
		Err:            e.Error(),
	}
}
