package stats

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	db  *sqlx.DB
	web *http.Server
}

func NewServer(ctx context.Context, db *sqlx.DB) *Server {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.GET("/", currentStats)

	return &Server{
		db: db,
		web: &http.Server{
			Addr:    "127.0.0.1:8080",
			Handler: e,
			BaseContext: func(l net.Listener) context.Context {
				return ctx
			},
		},
	}
}

func (s *Server) Start() {
	log.Infof("http server is available at http://%s/\n", s.web.Addr)

	go func() {
		if err := s.web.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			log.Info("http server stopped")
		} else {
			log.Errorf("http server stopped with error: %v\n", err)
		}
	}()
}

func (s *Server) Shutdown() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))

	go func() {
		for range ctx.Done() {
			cancel()
		}
	}()

	log.Debug("http server shutting down")
	s.web.Shutdown(ctx)
}

func currentStats(c *gin.Context) {
	c.JSON(200, gin.H{"stats": "ok"})
}
