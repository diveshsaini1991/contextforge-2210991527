package utils

import "testing"

// TestIsEven tests the IsEven function
func TestIsEven(t *testing.T) {
	tests := []struct {
		input int
		want  bool
	}{
		{2, true},
		{3, false},
		{0, true},
		{-2, true},
		{-3, false},
	}

	for _, tt := range tests {
		if got := IsEven(tt.input); got != tt.want {
			t.Errorf("IsEven(%d) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// Note: FormatNumber, IsOdd, Max, Min, and Abs have NO tests
// This creates coverage gaps
