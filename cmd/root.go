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

	"github.com/mattn/go-mastodon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
)

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
	v          *viper.Viper
	serverName string
	cfgFile    string
	verbose    bool
	client     *mastodon.Client
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.proma.json)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")
	rootCmd.PersistentFlags().StringVarP(&serverName, "server", "s", "mastodon.social", "server name")

	cobra.OnInitialize(initConfig)
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
		debugf("creating config file: %v", v.ConfigFileUsed())
		cobra.CheckErr(v.SafeWriteConfig())
	}

	debugf("reading config file: %v", v.ConfigFileUsed())

	if len(v.AllSettings()) > 0 && !rootCmd.Flags().Changed("server") {
		serverName = maps.Keys(v.AllSettings())[0]
		debug("using default serverName: ", serverName)
	}

	if v.InConfig(serverName) {
		configValues := v.GetStringMapString(serverName)

		clientConfig := &mastodon.Config{
			Server:       configValues["server"],
			ClientID:     configValues["clientid"],
			ClientSecret: configValues["clientsecret"],
			AccessToken:  configValues["accesstoken"],
		}

		client = mastodon.NewClient(clientConfig)
	} else {
		fmt.Printf("Credentials missing for server '%s'\n\n", serverName)
	}
}

func debug(a ...any) {
	if verbose {
		fmt.Println(a...)
	}
}

func debugf(format string, a ...any) {
	if verbose {
		debug(fmt.Sprintf(format, a...))
	}
}
