package calculator_test

import (
	"testing"

	calculator "github.com/asiafrolova/Multi-user-calculator/agent_service/pkg"
)

func TestParseOK(t *testing.T) {
	testCases := []struct {
		input calculator.SimpleExpression
		arg1  float64
		arg2  float64
		err   error
	}{
		{calculator.SimpleExpression{Arg1: "1.0", Arg2: "2.0"}, 1.0, 2.0, nil},
		{calculator.SimpleExpression{Arg1: "1.5", Arg2: "2.67"}, 1.5, 2.67, nil},
		{calculator.SimpleExpression{Arg1: "-1", Arg2: "0"}, -1.0, 0, nil},
		{calculator.SimpleExpression{Arg1: "1", Arg2: "1"}, 1, 1, nil},
	}
	for _, tc := range testCases {
		got1, got2, gotErr := tc.input.ParseArg()
		if gotErr != nil {
			t.Errorf("ParseString(%v) returned unexpected error: %v", tc.input, gotErr)
		}
		if got1 != tc.arg1 || got2 != tc.arg2 {
			t.Errorf("ParseString(%v) = %v %v, want %v %v", tc.input, got1, got2, tc.arg1, tc.arg2)
		}
	}
}

func TestCalcOK(t *testing.T) {
	testCases := []struct {
		input calculator.SimpleExpression
		want  float64
	}{
		{calculator.SimpleExpression{Arg1: "1.0", Arg2: "2.0", Operation: "+"}, 3.0},
		{calculator.SimpleExpression{Arg1: "40.0", Arg2: "20.0", Operation: "-"}, 20.0},
		{calculator.SimpleExpression{Arg1: "15.0", Arg2: "2.0", Operation: "*"}, 30.0},
		{calculator.SimpleExpression{Arg1: "1.5", Arg2: "2.0", Operation: "+"}, 3.5},
		{calculator.SimpleExpression{Arg1: "1.53", Arg2: "0.47", Operation: "+"}, 2.0},
		{calculator.SimpleExpression{Arg1: "1", Arg2: "1", Operation: "/"}, 1.0},
	}

	for _, tc := range testCases {
		err := tc.input.Calc()
		if err != nil {
			t.Errorf("Calc(%v) returned unexpected error: %v", tc.input, err)
		}
		if tc.input.Result != tc.want {
			t.Errorf("Calc(%v) = %v, want %v", tc.input, tc.input.Result, tc.want)
		}
	}
}
