package stats

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-mastodon"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

type Collector struct {
	client     *mastodon.Client
	db         *sqlx.DB
	sampleRate time.Duration
	done       chan struct{}
}

func NewCollector(client *mastodon.Client, db *sqlx.DB) *Collector {
	return &Collector{
		db:         db,
		client:     client,
		sampleRate: time.Second * 3,
	}
}

func (c *Collector) Collect(tagNames []string) error {
	// timelineFeed := client.ServerFeed(ctx, c.client)

	c.done = make(chan struct{}, 1)
	tick := time.NewTicker(c.sampleRate)
	defer tick.Stop()

	for {
		defer func() { c.done <- struct{}{} }()
		logrus.Info("collector run starting")

		// for _, tag := range tagNames {
		// 	items, err := timelineFeed(tag)

		// 	if err != nil {
		// 		logrus.Error("error collecting data: %v\n", err)
		// 		return err
		// 	}

		// 	out, err := json.MarshalIndent(&items, "", "  ")

		// 	if err != nil {
		// 		logrus.Error("error marshalling JSON: %v\n", err)
		// 		return err
		// 	}

		// 	fmt.Println(string(out))
		// }

		select {
		case <-tick.C:
			// proceed through next iteration
		case <-c.done:
			logrus.Debug("collector shutting down now..")
			return nil
		}
	}
}

func (c *Collector) Start(tagNames []string) {
	logrus.Infof("collector started. will refresh every %s", c.sampleRate)
	go c.Collect(tagNames)
}

func (c *Collector) Stop() {
	c.done <- struct{}{}
	timer := time.NewTimer(time.Second)

	select {
	case <-c.done:
		logrus.Debug("collector finished")
	case <-timer.C:
		logrus.Debug("shutting down now")
	}
}
