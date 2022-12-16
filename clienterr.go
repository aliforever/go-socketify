package socketify

type UpdateError struct {
	Update []byte
	Error  error
	Extra  []string
}

func newUpdateError(update []byte, err error, extra ...string) UpdateError {
	return UpdateError{
		Update: update,
		Error:  err,
		Extra:  extra,
	}
}
