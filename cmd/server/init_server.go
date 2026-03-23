package server

import (
	"avgys-gophermat/internal/config"
	"avgys-gophermat/internal/endpoints"
	"avgys-gophermat/internal/repository"
	"avgys-gophermat/internal/router"
	"avgys-gophermat/internal/service/auth"
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

func prepareDI(done context.Context, cfg *config.Config, traceLogger *zerolog.Logger) (*endpoints.Endpoints, error) {

	store, err := repository.CreateRepository(done, cfg)

	if err != nil {
		err = fmt.Errorf("error initializing repository: %w", err)
		return nil, err
	}

	authService := auth.NewAuthService(store)

	// shortifier := service.NewShortifier(done, generator, store, &cfg.RedirectDomain)
	h := endpoints.New(authService)

	return h, nil
}
