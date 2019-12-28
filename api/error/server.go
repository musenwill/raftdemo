package error

func Success(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 200,
		Code:           0,
		Err:            e.Error(),
	}
}

func UnknownError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 500,
		Code:           1,
		Err:            e.Error(),
	}
}

func ServerError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 500,
		Code:           2,
		Err:            e.Error(),
	}
}

func UnimplementedError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 501,
		Code:           3,
		Err:            e.Error(),
	}
}

func BadGatewayError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 502,
		Code:           4,
		Err:            e.Error(),
	}
}

func UnavailableError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 503,
		Code:           5,
		Err:            e.Error(),
	}
}
