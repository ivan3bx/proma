/*
Copyright © 2022 Ivan Moscoso

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/mattn/go-mastodon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
)

type LogFormatter struct{}

func (f *LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(fmt.Sprintf("%s\n", entry.Message)), nil
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "proma",
	Short: "Query tool from Mastodon",
	Long: `
proma is a CLI tool for querying data via the Mastodon API.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	v             *viper.Viper
	defaultServer string
	allServers    []string
	cfgFile       string
	verbose       bool
	mClient       *mastodon.Client
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.proma.json)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")
	rootCmd.PersistentFlags().StringSliceVarP(&allServers, "servers", "s", []string{"mastodon.social"}, "server names to check")

	cobra.OnInitialize(initLogging, initConfig)
}

func initLogging() {
	if verbose {
		log.SetFormatter(&log.TextFormatter{
			DisableTimestamp: true,
			PadLevelText:     true,
		})
		log.Info("Debug logs enabled")
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetFormatter(&LogFormatter{})
	}
}

func initConfig() {
	v = viper.NewWithOptions(viper.KeyDelimiter("|"))

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		v.AddConfigPath(home)
		v.SetConfigType("json")
		v.SetConfigName(".proma")
	}

	v.AutomaticEnv()
	err := v.ReadInConfig()

	if err != nil {
		log.Debugf("creating config file: %v", v.ConfigFileUsed())
		cobra.CheckErr(v.SafeWriteConfig())
	}

	log.Debugf("reading config file: %v", v.ConfigFileUsed())

	// default server taken from command flag
	defaultServer = allServers[0]

	// default server is overridden by any previous configuration
	if len(v.AllSettings()) > 0 && !rootCmd.Flags().Changed("servers") {
		defaultServer = maps.Keys(v.AllSettings())[0]
		log.Info("using default serverName: ", defaultServer)
	}

	if v.InConfig(defaultServer) {
		configValues := v.GetStringMapString(defaultServer)

		clientConfig := &mastodon.Config{
			Server:       configValues["server"],
			ClientID:     configValues["clientid"],
			ClientSecret: configValues["clientsecret"],
			AccessToken:  configValues["accesstoken"],
		}

		mClient = mastodon.NewClient(clientConfig)
	} else {
		log.Debugf("credentials missing for default server '%s'", defaultServer)
	}
}
