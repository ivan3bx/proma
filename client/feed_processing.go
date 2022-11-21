package client

import (
	"context"
	"encoding/json"
	"os"

	"github.com/mattn/go-mastodon"
	log "github.com/sirupsen/logrus"
)

type TagTimeline func(tag string) ([]*mastodon.Status, error)

// ServerFeed returns a TagTimeline using the provided client.
func ServerFeed(ctx context.Context, client *mastodon.Client) TagTimeline {
	return func(tag string) ([]*mastodon.Status, error) {
		log.Debugf("fetching timeline for tag: '%v'", tag)
		return client.GetTimelineHashtag(ctx, tag, false, nil)
	}
}

// FileFeed returns a TagTimeline using the provided filename source
func FileFeed(filename string) TagTimeline {
	return func(tag string) ([]*mastodon.Status, error) {
		f, _ := os.Open(filename)
		data := []*mastodon.Status{}
		err := json.NewDecoder(f).Decode(&data)
		return data, err
	}
}
