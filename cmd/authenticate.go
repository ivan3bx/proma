/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-mastodon"
	"github.com/pkg/browser"
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
		fmt.Printf("Server: %s\n", yellow(serverName))
		fmt.Printf("Re-run this command with '-server' to use a different server.\n\n")
		fmt.Printf("This will launch a browser window in order to authorize this app.\n")
		fmt.Println("Hit <Enter> to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')

		var err error
		client, err = RegisterNewClient()
		cobra.CheckErr(err)

		v.Set(serverName, client.Config)
		cobra.CheckErr(v.WriteConfig())

	},
}

func init() {
	rootCmd.AddCommand(authenticateCmd)
}

func RegisterNewClient() (*mastodon.Client, error) {
	done := make(chan os.Signal, 1)

	// Start temporary server to capture auth code
	var authCode string

	// Listen on default port
	listener := *newListener()
	listenerPort := listener.Addr().(*net.TCPAddr).Port
	listenerHost := fmt.Sprintf("%s:%v", "localhost", listenerPort)

	go func() {
		mux := http.NewServeMux()

		// Handle client-side redirect to extract 'auth' code, and close window
		mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
			authCode = r.URL.Query().Get("code")
			w.Write([]byte(`
			<html>
				<body>
					<h2>It is safe to close this window..</h2>
				</body>
			</html>
			`))
			done <- os.Interrupt
		})

		debugf("listening for auth response on port %v", listenerPort)
		if err := http.Serve(listener, mux); err != nil && err != http.ErrServerClosed {
			fmt.Printf("authentication listener failed to start: %s\n", err)
			os.Exit(1)
		}
	}()

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	startTimeout(done)

	app, err := mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
		Server:       serverURL(),
		ClientName:   "Links From Bookmarks",
		Scopes:       "read:bookmarks read:favourites",
		Website:      "https://github.com/ivan3bx/proma",
		RedirectURIs: fmt.Sprintf("http://%s/auth", listenerHost),
	})

	if err != nil {
		return nil, err
	}

	debugf("client-id    : %s", app.ClientID)
	debugf("client-secret: %s", app.ClientSecret)

	if err := browser.OpenURL(app.AuthURI); err != nil {
		return nil, err
	}

	<-done

	debug("listener stopped")

	if authCode == "" {
		return nil, errors.New("auth code was not present, or was blank")
	}

	// Create mastodon client
	client := mastodon.NewClient(&mastodon.Config{
		Server:       serverURL(),
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
	})

	if err = client.AuthenticateToken(context.Background(), authCode, app.RedirectURI); err != nil {
		return nil, err
	}

	fmt.Printf("authenticated to %s\n", client.Config.Server)

	return client, nil
}

func startTimeout(done chan<- os.Signal) {
	timer := time.NewTimer(time.Second * 60)

	go func() {
		<-timer.C
		debug("timeout exceeded. canceling authentication")
		done <- os.Interrupt
	}()
}

func serverURL() string {
	return fmt.Sprintf("https://%s", serverName)
}

func newListener() *net.Listener {
	listener, err := net.Listen("tcp", "localhost:3334")

	if err != nil {
		// attempt to use next available port
		listener, err = net.Listen("tcp", "localhost:0")
		if err != nil {
			panic(err)
		}
	}
	return &listener
}

func requireClient(cmd *cobra.Command, args []string) {
	if client == nil {
		fmt.Printf("See 'auth -h' to authenticate\n")
		os.Exit(1)
	}
}
