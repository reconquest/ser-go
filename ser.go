package ser

import (
	"fmt"
	"strings"

	"github.com/reconquest/hierr-go"
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
	Reason  interface{}
}

// HierarchicalError returns hierarchical representation of errors
func (err Error) HierarchicalError() string {
	return hierr.String(
		hierr.Error{
			Message: err.Message,
			Reason:  err.Reason,
		},
	)
}

func (err Error) GetReasons() []hierr.Reason {
	children, ok := err.Reason.([]hierr.Reason)
	if !ok {
		children = []hierr.Reason{}
		if err.Reason != nil {
			children = append(children, hierr.Reason(err.Reason))
		}
	}

	return children
}

func (err Error) GetMessage() string {
	return err.Message
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
	children := err.GetReasons()

	for _, item := range nested {
		children = append(children, hierr.Reason(item))
	}

	err.Reason = children
}

// Push specified nested errors into specified error and return new error.
func Push(top interface{}, nested ...interface{}) Error {
	err, ok := top.(Error)
	if !ok {
		err = Error{Message: fmt.Sprint(top)}
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
		Reason:  hierr.Reason(err),
	}
}

func linearalize(err error) string {
	var nested interface{}
	var message string

	if hierarchical, ok := err.(hierr.HierarchicalError); ok {
		nested = hierarchical.GetReasons()
		message = hierarchical.GetMessage()
	} else if hierarchical, ok := err.(hierr.Error); ok {
		nested = hierarchical.GetReasons()
		message = hierarchical.GetMessage()
	} else {
		return err.Error()
	}

	linear := fmt.Sprint(nested)
	switch typed := nested.(type) {
	case error:
		linear = linearalize(typed)

	case []hierr.Reason:
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
