package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/1Asi1/gophermart/internal/config"
	"github.com/1Asi1/gophermart/internal/integration"
	"github.com/1Asi1/gophermart/internal/integration/accrual"
	"github.com/1Asi1/gophermart/internal/repository"
	"github.com/1Asi1/gophermart/internal/service"
	"github.com/1Asi1/gophermart/internal/transport/rest"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

const (
	maxConnDB         = 10
	maxConnLifeTimeDB = 30
	maxConnIdleTimeDB = 30

	ReadTimeoutServer  = 5
	WriteTimeoutServer = 10
	IdleTimeoutServer  = 120
	timeoutShutdown    = 5
)

type Server struct {
	log zerolog.Logger
}

func New() Server {
	out := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05 -0700",
		NoColor:    true,
	}
	l := zerolog.New(out)

	return Server{log: l}
}

func (s *Server) Run() {
	l := s.log.With().Str("server", "Run").Logger()
	cfg := config.New(l)

	st, err := repository.New(repository.Config{
		ConnDSN:         cfg.DBConnDSN,
		MaxConn:         maxConnDB,
		MaxConnLifeTime: maxConnLifeTimeDB * time.Second,
		MaxConnIdleTime: maxConnIdleTimeDB * time.Second,
	})
	if err != nil {
		l.Fatal().Err(err).Msg("repository.New")
	}

	cl := accrual.New(cfg, l)

	sv := service.New(st, cl)

	mg := integration.New(&cl, st, l)
	go func() {
		mg.Sync(context.Background())
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	errs, ctx := errgroup.WithContext(ctx)

	httpServer := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      rest.New(sv, l),
		ReadTimeout:  ReadTimeoutServer * time.Second,
		WriteTimeout: WriteTimeoutServer * time.Second,
		IdleTimeout:  IdleTimeoutServer * time.Second,
	}

	errs.Go(func() error {
		if err = httpServer.ListenAndServe(); err != nil {
			log.Fatalf("http.ListenAndServe: %v", err)
		}
		return nil
	})

	<-ctx.Done()

	l.Info().Msg("shutting down gracefully")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeoutShutdown*time.Second)
	defer cancel()

	// Perform application shutdown with a maximum timeout of 5 seconds.
	if err := httpServer.Shutdown(timeoutCtx); err != nil {
		l.Err(err).Msg(err.Error())
	}
}
