/*
Copyright © 2022 Ivan Moscoso
*/
package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ivan3bx/proma/client"
	"github.com/ivan3bx/proma/stats"
	"github.com/jmoiron/sqlx"
	"github.com/mattn/go-mastodon"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var tagNames []string
var webServer bool

// collectCmd represents the collect command
var collectCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collects and aggregates tagged posts over time",
	Long: `Collects posts for the provided tag(s) and aggregates statistics over time.

Example:

Collect posts tagged with '#outage', every 2 minutes on 'mastodon.social'
proma collect -t outage -i 2 -s mastodon.social
`,
	PreRun: anonymousClientAllowed,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			clients []*mastodon.Client
			db      *sqlx.DB
			c       *stats.Collector
			w       *stats.Server
		)

		db = initDB()

		for _, s := range getTargetServers(cmd) {
			clients = append(clients, client.NewAnonymousClient(s))
		}

		c = stats.NewCollector(clients, db)

		if webServer {
			// start collector in the background
			c.Start(cmd.Context(), tagNames)

			// start web server
			w = stats.NewServer(cmd.Context(), db)
			w.Start()

			waitForInterrupt(cmd.Context(), func() {
				c.Stop()
				if w != nil {
					w.Shutdown()
				}
			})

		} else {
			c.Collect(cmd.Context(), tagNames)
		}
	},
}

func init() {
	rootCmd.AddCommand(collectCmd)
	collectCmd.Flags().StringSlice("server", []string{}, "servers to collect from (comma delimited)")
	collectCmd.Flags().StringSliceVarP(&tagNames, "tags", "t", []string{}, "tag names")
	collectCmd.Flags().BoolVar(&webServer, "http", false, "display stats page (http://localhost:8080/)")
}

func initDB() *sqlx.DB {
	var (
		db  *sqlx.DB
		err error
	)

	if _, err := os.Stat("proma.db"); err == nil {
		return sqlx.MustOpen("sqlite3", "proma.db")
	}

	if db, err = sqlx.Open("sqlite3", "proma.db"); err != nil {
		panic(err)
	}

	schema := `
		CREATE TABLE tags (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		);

		CREATE UNIQUE INDEX idx_tags_name ON tags (name);

		CREATE TABLE posts (
			id INTEGER PRIMARY KEY,
			post_id TEXT NOT NULL,
			account_id TEXT NOT NULL,
			server TEXT NOT NULL,
			uri TEXT NOT NULL,
			lang TEXT DEFAULT 'en' NOT NULL,
			content_html TEXT,
			content_text TEXT,
			created_at TEXT
		);

		CREATE UNIQUE INDEX idx_posts_uri ON posts (uri);

		CREATE TABLE posts_tags (
			post_id INTEGER NOT NULL,
			tag_id INTEGER NOT NULL,
			PRIMARY KEY (post_id, tag_id)
		);

		CREATE VIRTUAL TABLE content_index USING FTS5 (
			post_id,
			content
		);
	`

	db.MustExec(schema)

	return db
}

// getTargetServers returns a list of servers if specified as a flag
// or els edefaults to the root command's default "server" value.
func getTargetServers(cmd *cobra.Command) []string {
	var (
		servers []string
		err     error
	)
	servers, err = cmd.Flags().GetStringSlice("server")

	if err != nil {
		log.Fatal(err)
	}

	if len(servers) == 0 {
		servers = []string{serverName}
	}
	return servers
}

// waitForInterrupt will block until either user interrupt is detected,
// or the provided context is marked Done(). It will then invoke the completion func.
func waitForInterrupt(ctx context.Context, complete func()) {
	userInterrupt := make(chan os.Signal, 1)
	signal.Notify(userInterrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-userInterrupt:
	case <-ctx.Done():
	}

	complete()
}
