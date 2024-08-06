package main

import "testing"

func TestHumanizeSize(t *testing.T) {
	testCases := []struct {
		input    int64
		expected string
	}{
		{0, "0.00B"},
		{1024, "1.00KB"},
		{1048576, "1.00MB"},
		{1073741824, "1.00GB"},
		{109951162777, "102.40GB"},
	}

	for _, tc := range testCases {
		result := HumanizeSize(tc.input)
		if result != tc.expected {
			t.Errorf("Expected %s, but got %s", tc.expected, result)
		}
	}
}
