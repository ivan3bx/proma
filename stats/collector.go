package stats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ivan3bx/proma/client"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-mastodon"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type Collector struct {
	client     *mastodon.Client
	db         *sqlx.DB
	sampleRate time.Duration
	stop       chan struct{}
}

func NewCollector(client *mastodon.Client, db *sqlx.DB) *Collector {
	return &Collector{
		db:         db,
		client:     client,
		sampleRate: time.Minute * 1,
	}
}

// Collect performs a single collection of timeline for the provided tags
// and imports it to the database configured on the collector.
// It returns an error returned by the server or nil if successful.
func (c *Collector) Collect(ctx context.Context, tagNames []string) error {
	timelineFeed := client.ServerFeed(ctx, c.client)

	for _, tag := range tagNames {
		items, err := timelineFeed(tag)

		if err != nil {
			log.Error("error collecting data: %v\n", err)
			return err
		}

		out, err := json.MarshalIndent(&items, "", "  ")

		if err != nil {
			log.Error("error marshalling JSON: %v\n", err)
			return err
		}

		fmt.Println(string(out))
	}

	return nil
}

// Start will run this collector in a loop. It will shut down
// when Stop() is called, or if Collect() function returns an error.
func (c *Collector) Start(ctx context.Context, tagNames []string) {
	log.Infof("collector started. will refresh every %s", c.sampleRate)
	c.stop = make(chan struct{}, 1)

	go func() {
		ctx, cancel := context.WithCancel(ctx)
		tick := time.NewTicker(c.sampleRate)

		defer func() {
			tick.Stop()
			c.stop <- struct{}{} // signals back to Stop()
		}()

		for {
			log.Debug("collector run starting")

			if err := c.Collect(ctx, tagNames); err != nil {
				log.Errorf("collector failed with error: %v", err)
				c.stop <- struct{}{}
			}

			select {
			case <-tick.C:
				// proceed through next iteration
			case <-c.stop:
				log.Debug("collector shutting down..")
				cancel()
				return
			}
		}
	}()
}

// Stop will shutdown the collector, waiting a predetermined time for a graceful
// shutdown, after which it will return.
func (c *Collector) Stop() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))

	c.stop <- struct{}{} // signal main routine to stop

	select {
	case <-c.stop: // main routine signaled back
		log.Debug("collector finished")
	case <-ctx.Done():
		log.Debug("shutting down now")
	}

	cancel()
}
