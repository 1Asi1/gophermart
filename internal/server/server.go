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
		MaxConn:         10,
		MaxConnLifeTime: 30 * time.Second,
		MaxConnIdleTime: 30 * time.Second,
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
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	errs.Go(func() error {
		if err = httpServer.ListenAndServe(); err != nil {
			log.Fatalf("http.ListenAndServe: %v", err)
			return err
		}
		return nil
	})

	<-ctx.Done()

	l.Info().Msg("shutting down gracefully")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Perform application shutdown with a maximum timeout of 5 seconds.
	if err := httpServer.Shutdown(timeoutCtx); err != nil {
		l.Err(err).Msg(err.Error())
	}
}
