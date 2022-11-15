package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattn/go-mastodon"
	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
)

func RegisterNewClient(serverName string) (*mastodon.Client, error) {
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

		log.Debugf("listening for auth response on port %v", listenerPort)
		if err := http.Serve(listener, mux); err != nil && err != http.ErrServerClosed {
			log.Fatalf("authentication listener failed to start: %s\n", err)
		}
	}()

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	startTimeout(done)

	app, err := mastodon.RegisterApp(context.Background(), &mastodon.AppConfig{
		Server:       serverURL(serverName),
		ClientName:   "Links From Bookmarks",
		Scopes:       "read:bookmarks read:favourites",
		Website:      "https://github.com/ivan3bx/proma",
		RedirectURIs: fmt.Sprintf("http://%s/auth", listenerHost),
	})

	if err != nil {
		return nil, err
	}

	log.Debugf("client-id    : %s", app.ClientID)
	log.Debugf("client-secret: %s", app.ClientSecret)

	if err := browser.OpenURL(app.AuthURI); err != nil {
		return nil, err
	}

	<-done

	log.Debug("listener stopped")

	if authCode == "" {
		return nil, errors.New("auth code was not present, or was blank")
	}

	// Create mastodon client
	client := mastodon.NewClient(&mastodon.Config{
		Server:       serverURL(serverName),
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
	})

	if err = client.AuthenticateToken(context.Background(), authCode, app.RedirectURI); err != nil {
		return nil, err
	}

	log.Infof("authenticated to %s\n", client.Config.Server)

	return client, nil
}

func startTimeout(done chan<- os.Signal) {
	timer := time.NewTimer(time.Second * 60)

	go func() {
		<-timer.C
		log.Debug("timeout exceeded. canceling authentication")
		done <- os.Interrupt
	}()
}

func serverURL(serverName string) string {
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
