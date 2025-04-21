package orkestrator_test

import (
	"reflect"
	"testing"

	"github.com/asiafrolova/Final_task/orkestrator_service/pkg/orkestrator"
)

func TestTokenizeOK(t *testing.T) {
	testCases := []struct {
		input orkestrator.Expression

		want []string
	}{
		{orkestrator.Expression{Exp: "1+2*3"}, []string{"1", "+", "2", "*", "3"}},
		{orkestrator.Expression{Exp: "(1+2)*3"}, []string{"(", "1", "+", "2", ")", "*", "3"}},
		{orkestrator.Expression{Exp: "-1*(1+2)"}, []string{"-1", "*", "1", "*", "(", "1", "+", "2", ")"}},
		{orkestrator.Expression{Exp: "2*(-3)"}, []string{"2", "*", "(", "-1", "*", "3", ")"}},
		{orkestrator.Expression{Exp: "2*(-3.14)"}, []string{"2", "*", "(", "-1", "*", "3.14", ")"}},
	}

	for _, tc := range testCases {
		got, err := tc.input.TokenizeString()
		if err != nil {
			t.Errorf("TokenizeString(%v) returned unexpected error: %v", tc.input, err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("TokenizeString(%v) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
func TestTokenizeInvalidExpression(t *testing.T) {
	testCases := []struct {
		input orkestrator.Expression
	}{
		{orkestrator.Expression{Exp: "1+2**3"}},
		{orkestrator.Expression{Exp: "(1+2))*3"}},
		{orkestrator.Expression{Exp: "-1.0.9*(1+2)"}},
		{orkestrator.Expression{Exp: "2*(~-3)"}},
		{orkestrator.Expression{Exp: "2*a(-3.14)"}},
		{orkestrator.Expression{Exp: "2-3.14.0)"}},
	}

	for _, tc := range testCases {
		_, err := tc.input.TokenizeString()
		if err == nil {
			t.Errorf("TokenizeString(%v) did not return an error", tc.input)
		} else if err != orkestrator.ErrInvalidExpression {
			t.Errorf("TokenizeString(%v) returned an unexpected error: %v", tc.input, err)
		}
	}
}
func TestCheckExpressionTest(t *testing.T) {

	testCases := []struct {
		input string

		want bool
	}{
		{"1+2", true},
		{"-1-(3*9)", true},
		{"((1-2)*9)", true},
		{"1--2", false},
		{"1s-90", false},
		{"((1-1)))", false},
		{"2-11-", false},
	}

	for _, tc := range testCases {
		got := orkestrator.CheckExpression(tc.input)

		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("TokenizeString(%v) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
