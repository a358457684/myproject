package util

import (
	"errors"
	"fmt"
)

func WrapErr(err error, msg string) error {
	if err == nil {
		return errors.New(msg)
	}
	return fmt.Errorf("%s::%w", msg, err)
}

func UnwrapErr(err error) (error, string) {
	subErr := errors.Unwrap(err)
	if subErr == nil {
		return err, err.Error()
	}
	i := len(err.Error()) - len(subErr.Error())
	return subErr, err.Error()[:i-2]
}
