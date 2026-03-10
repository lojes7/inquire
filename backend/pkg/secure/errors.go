package secure

import "errors"

type MyError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"err"`
}

func (e *MyError) Error() string {
	return e.Err.Error()
}

// Unwrap 实现 Go 1.13 错误链解包接口，允许 errors.Is/As 穿透检查底层错误
func (e *MyError) Unwrap() error {
	return e.Err
}

// Wrap 包装一个 MyError 对象，包含错误码、消息和原始错误
func Wrap(code int, msg string, err error) *MyError {
	return &MyError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}

// Unwrap 尝试将一个 error 类型的错误转换为 MyError 类型
// 如果成功则返回 MyError 对象和 true，否则返回 nil
func Unwrap(err error) *MyError {
	var MyErr *MyError

	if errors.As(err, &MyErr) {
		return MyErr
	}
	return nil
}
