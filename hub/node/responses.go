package node

import "errors"

// failure turns an error the node reported inside its response into an error.
func (x *BoolResponse) failure() error {
	return reportedError(x.ErrorCode, x.ErrorMessage)
}

func (x *ErrorResponse) failure() error {
	return reportedError(x.ErrorCode, x.ErrorMessage)
}

func reportedError(code int32, message string) error {
	if code == 0 {
		return nil
	}
	return errors.New(message)
}
