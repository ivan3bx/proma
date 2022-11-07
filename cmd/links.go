/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mattn/go-mastodon"
	"github.com/spf13/cobra"
)

type LinkRef struct {
	AccountName string `json:"profileName"`
	AccountURL  string `json:"profileURL"`
	URL         string `json:"URL"`
	LinkRef     string `json:"linkRef"`
}

var linksCmd = &cobra.Command{
	Use:   "links",
	Short: "Extract links from any saved bookmarks",
	Long: `
Collects links embedded in the content of your saved bookmarks.`,
	PreRun: requireClient,
	Run: func(cmd *cobra.Command, args []string) {
		st, err := client.GetBookmarks(cmd.Context(), &mastodon.Pagination{Limit: limit})
		cobra.CheckErr(err)

		outputLinks(st)
	},
}

var limit int64

func init() {
	rootCmd.AddCommand(linksCmd)
	linksCmd.PersistentFlags().Int64Var(&limit, "limit", 10, "Limit the number of bookmarks to search for links.")
}

func outputLinks(status []*mastodon.Status) {
	refs := []LinkRef{}

	for _, entry := range status {
		html := entry.Content
		origin := parseOrigin(entry.Account.URL)

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))

		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing content: %v", err)
		}

		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			if href, ok := s.Attr("href"); ok {

				if strings.HasPrefix(href, origin) {
					debug("skipping internal href: ", href)
					return
				}

				refs = append(refs, LinkRef{
					URL:         entry.URL,
					LinkRef:     href,
					AccountName: entry.Account.Username,
					AccountURL:  entry.Account.URL,
				})
			}
		})

	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(refs)
}

func parseOrigin(accountURL string) string {
	url, _ := url.Parse(accountURL)
	return fmt.Sprintf("%s://%s", url.Scheme, url.Hostname())
}
