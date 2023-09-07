package orm

import (
	"github.com/sirupsen/logrus"
)

var (
	exists   = "ExistsError"
	notFound = "NotFoundError"
)

type BaseError struct {
	Name  string
	Table string
	Id    string
	Msg   string
	Level logrus.Level
}

func (e BaseError) Error() string {
	return e.Msg
}

func ExistsError(table, id string, level logrus.Level) BaseError {
	return newBaseError(exists, table, id, "The data is exists", level)
}

func (e BaseError) IsExistsError() bool {
	return e.Name == exists
}

func NotFoundError(table, id string, level logrus.Level) BaseError {
	return newBaseError(notFound, table, id, "The data not found", level)
}

func (e BaseError) IsNotFoundError() bool {
	return e.Name == notFound
}

func newBaseError(name, table, id, msg string, level logrus.Level) BaseError {
	return BaseError{
		Name:  name,
		Table: table,
		Id:    id,
		Msg:   msg,
		Level: level,
	}
}
