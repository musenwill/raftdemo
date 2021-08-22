package error

func DataConflictError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 400,
		Code:           120,
		Err:            e.Error(),
	}
}

func DataNotFoundError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 404,
		Code:           121,
		Err:            e.Error(),
	}
}

func DatExceedError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 400,
		Code:           122,
		Err:            e.Error(),
	}
}
