package server

import (
	"avgys-gophermat/internal/config"
	"avgys-gophermat/internal/handler"
	"avgys-gophermat/internal/repository"
	"avgys-gophermat/internal/router"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
)

func GetServer(done context.Context, traceLogger *zerolog.Logger) (*http.Server, error) {

	cfg, err := config.GetConfig(os.Args[1:], traceLogger)

	if err != nil {
		return nil, err
	}

	h, err := prepareDI(done, cfg, traceLogger)

	if err != nil {
		return nil, err
	}

	r := router.NewRouter(h)

	srv := &http.Server{
		Addr:    cfg.AppAddr,
		Handler: r,
	}

	return srv, nil
}

func prepareDI(done context.Context, cfg *config.Config, traceLogger *zerolog.Logger) (*handler.Handlers, error) {

	store, err := repository.NewRepository(done, cfg, traceLogger)

	if err != nil {
		err = fmt.Errorf("error initializing repository: %w", err)
		return nil, err
	}

	// shortifier := service.NewShortifier(done, generator, store, &cfg.RedirectDomain)
	h := handler.NewHandlers(store)

	return h, nil
}
