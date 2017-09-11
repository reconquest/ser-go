package ser

import (
	"errors"
	"io"
	"testing"

	"github.com/reconquest/hierr-go"
	"github.com/stretchr/testify/assert"
)

func TestError_HierarchicalError_ReturnsHierarchicalRepresentation(t *testing.T) {
	test := assert.New(t)

	err := Push(
		"1",
		Push(
			"2",
			Push(
				"c",
				"d",
			),
			Push(
				"a",
				Push("a-1"),
				Push("a-2"),
			),
		),
		Push(
			"5",
			errors.New("6"),
		),
		Push(
			"7",
			hierr.Errorf(errors.New("10"), "9"),
		),
	)

	test.Equal(`1
├─ 2
│  ├─ c
│  │  └─ d
│  │
│  └─ a
│     ├─ a-1
│     └─ a-2
│
├─ 5
│  └─ 6
│
└─ 7
   └─ 9
      └─ 10`, err.HierarchicalError())
}

func TestError_LinearError_ReturnsLinearRepresentation(t *testing.T) {
	test := assert.New(t)

	err := Error{
		Message: "1",
		Reason: []hierr.Reason{
			Error{
				Message: "2",
				Reason: []hierr.Reason{
					Error{
						Message: "a",
						Reason:  "b",
					},
					Error{
						Message: "c",
						Reason:  "d",
					},
				},
			},
			Error{
				Message: "5",
				Reason:  errors.New("6"),
			},
			Error{
				Message: "7",
				Reason:  hierr.Errorf(errors.New("10"), "9"),
			},
		},
	}

	test.Equal(`1: 2: a: b; c: d; 5: 6; 7: 9: 10`, err.LinearError())
}

func TestError_Push_AddsReasonItem(t *testing.T) {
	test := assert.New(t)

	err := Error{
		Message: "1",
		Reason:  "2",
	}
	err.Push("3", "4")

	test.EqualValues(Error{
		Message: "1",
		Reason: []hierr.Reason{
			"2", "3", "4",
		},
	}, err)
}

func TestError_Serialize(t *testing.T) {
	test := assert.New(t)

	err := Errorf("1", "2%s", "3")
	test.Equal("23: 1", err.Serialize(Linear))
	test.Equal("23\n└─ 1", err.Serialize(Hierarchical))
}

func TestSerializeError(t *testing.T) {
	test := assert.New(t)

	test.Equal("1: EOF", SerializeError(hierr.Errorf(io.EOF, "1"), Linear))
	test.Equal("1\n└─ EOF", SerializeError(hierr.Errorf(io.EOF, "1"), Hierarchical))
}
