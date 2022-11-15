package client

import "testing"

func TestServerURL(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "success",
			input:    "mastodon.social",
			expected: "https://mastodon.social",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := serverURL(tc.input)
			if actual != tc.expected {
				t.Errorf("expected URL to match (%v / %v)\n", tc.expected, actual)
			}
		})
	}
}
