package logger

import "fmt"

type ZaloError struct {
	Code    int
	Message string
}

func (e ZaloError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("error #%d: %s", e.Code, e.Message)
	}
	return e.Message
}
