package secure

type MyError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"err"`
}

func (e *MyError) Error() string {
	return e.Err.Error()
}
