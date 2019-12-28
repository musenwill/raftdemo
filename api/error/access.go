package error

func RequestError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 400,
		Code:           110,
		Err:            e.Error(),
	}
}

func RequestTooFrequentError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 429,
		Code:           111,
		Err:            e.Error(),
	}
}

func ForbiddenError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 403,
		Code:           112,
		Err:            e.Error(),
	}
}

func RequestTimeoutError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 408,
		Code:           113,
		Err:            e.Error(),
	}
}

func RequestOverloadError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 413,
		Code:           114,
		Err:            e.Error(),
	}
}

func RequestUnauthorizedError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 401,
		Code:           115,
		Err:            e.Error(),
	}
}

func RequestNotAllowedError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 405,
		Code:           116,
		Err:            e.Error(),
	}
}

func NeedLoginError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 401,
		Code:           130,
		Err:            e.Error(),
	}
}

func TokenExpiredError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 401,
		Code:           131,
		Err:            e.Error(),
	}
}

func KickedOutError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 401,
		Code:           132,
		Err:            e.Error(),
	}
}

func PasswordError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 401,
		Code:           133,
		Err:            e.Error(),
	}
}

func VerifyCodeError(e error) *HttpError {
	return &HttpError{
		httpStatusCode: 401,
		Code:           134,
		Err:            e.Error(),
	}
}
