package ufwb

import (
	"testing"
)

func TestNewExpression(t *testing.T) {
	var tests = []struct {
		input string
		want  Expression
	}{
		{input: "", want: nil},
		{input: "-1", want: ConstExpression(-1)},
		{input: "0", want: ConstExpression(0)},
		{input: "1", want: ConstExpression(1)},
		{input: "unlimited", want: StringExpression("unlimited")},
		{input: "blah", want: StringExpression("blah")},
	}

	for _, test := range tests {
		got := NewExpression(test.input)
		if got != test.want {
			t.Errorf("NewExpression(%q) = %v, want %v", test.input, got, test.want)
		}
	}

}
