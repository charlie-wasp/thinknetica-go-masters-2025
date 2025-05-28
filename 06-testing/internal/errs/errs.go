package errs

type ErrBadRequest struct {
	msg string
}

func (e ErrBadRequest) Error() string {
	return e.msg
}

func NewErrBadRequest(msg string) ErrBadRequest {
	return ErrBadRequest{msg: msg}
}
