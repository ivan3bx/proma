/*
Copyright © 2022 Ivan Moscoso
*/
package cmd

import (
	"bufio"
	"os"

	"github.com/fatih/color"
	"github.com/ivan3bx/proma/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var yellow = color.New(color.FgYellow, color.Bold).SprintFunc()

// authenticateCmd represents the authenticate command
var authenticateCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with a Mastodon server.",
	Long: `Authenticate with a given Mastodon server.
You may store credentials for multiple servers,
but only one server will be used when any commands are run.
	
Examples:
  proma auth -s 'indieweb.social'

  Opens browser to authenticate with the server
  and saves an AccessToken to the config file.
`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Server: %s\n", yellow(defaultServer))
		log.Infof("Re-run this command with '-server' to use a different server.\n\n")
		log.Infof("This will launch a browser window in order to authorize this app.\n")
		log.Infof("Hit <Enter> to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')

		var err error
		c, err := client.RegisterNewClient(defaultServer)
		cobra.CheckErr(err)

		v.Set(defaultServer, c.Config)
		cobra.CheckErr(v.WriteConfig())

	},
}

func init() {
	rootCmd.AddCommand(authenticateCmd)
}

func anonymousClientAllowed(cmd *cobra.Command, args []string) {
	if mClient == nil {
		mClient = client.NewAnonymousClient(defaultServer)
	}
}

func requireClient(cmd *cobra.Command, args []string) {
	if mClient == nil {
		log.Infof("See 'auth -h' to authenticate\n")
		os.Exit(1)
	}
}
