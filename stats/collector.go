package stats

import (
	"context"
	"database/sql"
	"time"

	"github.com/ivan3bx/proma/client"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-mastodon"
	_ "github.com/mattn/go-sqlite3"

	log "github.com/sirupsen/logrus"
)

type Collector struct {
	clients    []*mastodon.Client
	db         *sqlx.DB
	sampleRate time.Duration
	stop       chan struct{}
}

func NewCollector(clients []*mastodon.Client, db *sqlx.DB) *Collector {
	return &Collector{
		db:         db,
		clients:    clients,
		sampleRate: time.Minute * 1,
	}
}

// Collect performs a single collection of timeline for the provided tags
// and imports it to the database configured on the collector.
// It returns an error returned by the server or nil if successful.
func (c *Collector) Collect(ctx context.Context, tagNames []string) error {
	for _, cl := range c.clients {
		log.Info("collecting from server: ", cl.Config.Server)
		timelineFeed := client.ServerFeed(ctx, cl)

		for _, tag := range tagNames {
			items, err := timelineFeed(tag)

			if err != nil {
				log.Errorf("error collecting data: %v\n", err)
				return err
			}

			for _, item := range items {
				{
					var exists bool

					err := c.db.Get(&exists, "SELECT 1 FROM posts WHERE uri = ?", item.URI)

					if err != nil && err != sql.ErrNoRows {
						return err
					}

					if exists {
						log.Debug("skipping row")
						continue
					}

					postRes := sqlx.MustExec(c.db, `
				INSERT INTO posts (
					post_id,
					account_id,
					server,
					uri,
					lang,
					content_html,
					created_at
				) VALUES (
					?, ?, ?, ?, ?, ?, ?
				);`,
						item.ID,
						item.Account.ID,
						cl.Config.Server,
						item.URI,
						coalesceString("en", item.Language),
						item.Content,
						item.CreatedAt,
					)

					postID, err := postRes.LastInsertId()

					if err != nil {
						return err
					}

					log.Debug("inserted post")

					for _, tag := range item.Tags {
						sqlx.MustExec(c.db, `INSERT OR IGNORE INTO tags (name) VALUES (?);`, tag.Name)

						sqlx.MustExec(c.db, `
					INSERT INTO posts_tags (
						post_id,
						tag_id
					) VALUES (
						?, (SELECT id FROM tags WHERE name = ?)
					);`, postID, tag.Name)
					}
				}
			}
		}
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

		defer tick.Stop()
		defer cancel()

		for {
			log.Debug("collector run starting")

			if err := c.Collect(ctx, tagNames); err != nil {
				log.Errorf("collector failed with error: %v", err)
				tick.Stop()
				c.Stop()
			}

			select {
			case <-tick.C:
				// proceed through next iteration
			case <-c.stop:
				log.Debug("collector shutting down..")
				c.stop <- struct{}{} // signals back to Stop()
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

// Report generates a list of posts matching one or more of the provided tagNames.
// It returns an error if the underlying SQL query fails.
func (c *Collector) Report(ctx context.Context, tagNames []string) ([]*Status, error) {
	var results []*Status

	query, args, err := sqlx.In(`
		SELECT
		created_at, uri, lang, content_html as content,
		(
			SELECT group_concat(tt.name)
			FROM posts_tags ptt
			JOIN tags tt ON tt.id = ptt.tag_id
			WHERE post_id = p.id
			ORDER BY tt.name
		) tag_list
		FROM
			posts p
		INNER JOIN
			posts_tags pt ON pt.post_id = p.id
		INNER JOIN
			tags tag ON tag.id = pt.tag_id
		WHERE
			tag.name IN (?)
		AND
			created_at > date('now', '-2 days')
		ORDER BY created_at DESC;
	`, tagNames)

	if err != nil {
		return nil, err
	}

	query = c.db.Rebind(query)

	if err := c.db.Select(&results, query, args...); err != nil {
		return nil, err
	}
	return results, nil
}

func coalesceString(defaultVal, val string) string {
	if val == "" {
		return defaultVal
	}
	return val
}
