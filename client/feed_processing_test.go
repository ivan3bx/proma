package client

import (
	"testing"

	"github.com/mattn/go-mastodon"
)

func TestFetchItems(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		assertFunc func(*testing.T, []*mastodon.Status, error)
	}{
		{
			name:  "tagged items",
			input: "testfiles/tag_timeline.json",
			assertFunc: func(t *testing.T, items []*mastodon.Status, err error) {
				if err != nil {
					t.Error(err)
				}
				if len(items) < 2 {
					t.Error("Expected 2 items, was ", len(items))
				}
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			items, err := FileFeed(tc.input)("ignore_")
			tc.assertFunc(t, items, err)
		})
	}
}
