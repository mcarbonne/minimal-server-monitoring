package errdef

type errorCode int

const (
	notFound errorCode = iota
)

type apiError struct {
	error
	code errorCode
}

func (ctx apiError) Code() errorCode {
	return ctx.code
}

type apiErrorInterface interface {
	Code() errorCode
}

func is(err error, code errorCode) bool {
	itf, ok := (err).(apiErrorInterface)
	return ok && itf.Code() == code
}

func IsErrNotFound(err error) bool {
	return is(err, notFound)
}

func errorBuilder(err error, code errorCode) error {
	if err == nil || is(err, code) {
		return err
	}
	return apiError{
		error: err,
		code:  code}
}

func ErrorNotFound(err error) error {
	return errorBuilder(err, notFound)
}
