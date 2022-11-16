/*
Copyright © 2022 Ivan Moscoso
*/
package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ivan3bx/proma/stats"
	"github.com/jmoiron/sqlx"
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
			db *sqlx.DB
			c  *stats.Collector
			w  *stats.Server
		)

		db = sqlx.MustConnect("sqlite3", "proma.db")
		c = stats.NewCollector(mClient, db)

		if webServer {
			w = stats.NewServer(cmd.Context(), db)
			w.Start()
		}

		c.Start(tagNames)

		waitForInterrupt(cmd.Context(), func() {
			c.Stop()
			if w != nil {
				w.Shutdown()
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(collectCmd)
	collectCmd.Flags().StringSliceVarP(&tagNames, "tags", "t", []string{}, "tag names")
	collectCmd.Flags().BoolVar(&webServer, "http", false, "display stats page (http://localhost:8080/)")
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
