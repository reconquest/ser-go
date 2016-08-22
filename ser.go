package ser

import (
	"fmt"
	"strings"

	"github.com/reconquest/hierr"
)

type (
	Mode int
)

const (
	Linear       Mode = iota
	Hierarchical Mode = iota
)

type Error struct {
	Message string
	Nested  interface{}
}

// HierarchicalError returns hierarchical representation of errors
func (err Error) HierarchicalError() string {
	return hierr.String(
		hierr.Error{
			Message: err.Message,
			Nested:  err.Nested,
		},
	)
}

// LinearError returns linear representation of errors
func (err Error) LinearError() string {
	return linearalize(err)
}

// Error implements error interface and returns hierarchical representation.
func (err Error) Error() string {
	return err.HierarchicalError()
}

// Push specified nested errors into given error.
func (err *Error) Push(nested ...interface{}) {
	children, ok := err.Nested.([]interface{})
	if !ok {
		children = []interface{}{}
		if err.Nested != nil {
			children = append(children, err.Nested)
		}
	}

	for _, item := range nested {
		children = append(children, interface{}(item))
	}

	err.Nested = children
}

// Push specified nested errors into specified error and return new error.
func Push(top interface{}, nested ...interface{}) Error {
	err, ok := top.(Error)
	if !ok {
		return Error{Message: fmt.Sprint(top), Nested: nested}
	}

	err.Push(nested...)

	return err
}

// Serialize given error using specified mode.
func (err Error) Serialize(mode Mode) string {
	switch mode {
	case Linear:
		return err.LinearError()
	case Hierarchical:
		return err.HierarchicalError()
	default:
		return ""
	}
}

// SerializeError returns string representation of specified error in specified
// mode format.
func SerializeError(err error, mode Mode) string {
	switch mode {
	case Linear:
		return linearalize(err)
	case Hierarchical:
		return hierr.String(err)
	default:
		return err.Error()
	}
}

func Errorf(err interface{}, format string, arg ...interface{}) Error {
	return Error{
		Message: fmt.Sprintf(format, arg...),
		Nested:  interface{}(err),
	}
}

func (err Error) GetNested() []interface{} {
	if errs, ok := err.Nested.([]interface{}); ok {
		return errs
	}

	return []interface{}{err.Nested}
}

func linearalize(err error) string {
	var nested interface{}
	var message string

	if packageError, ok := err.(Error); ok {
		nested = packageError.Nested
		message = packageError.Message
	} else if hierrError, ok := err.(hierr.Error); ok {
		nested = hierrError.Nested
		message = hierrError.Message
	} else {
		return err.Error()
	}

	linear := fmt.Sprint(nested)
	switch typed := nested.(type) {
	case error:
		linear = linearalize(typed)

	case []interface{}:
		linearItems := []string{}

		for _, nestedItem := range typed {
			linearItem := fmt.Sprint(nestedItem)
			switch part := nestedItem.(type) {
			case error:
				linearItem = linearalize(part)

			case string:
				linearItem = part
			}

			linearItems = append(
				linearItems,
				linearItem,
			)
		}

		linear = strings.Join(linearItems, "; ")
	}

	return message + ": " + linear
}
