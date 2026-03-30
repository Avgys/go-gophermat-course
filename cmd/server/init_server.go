package server

import (
	"avgys-gophermat/internal/config"
	"avgys-gophermat/internal/db"
	"avgys-gophermat/internal/endpoints"
	"avgys-gophermat/internal/processor"
	"avgys-gophermat/internal/repository"
	"avgys-gophermat/internal/router"
	"avgys-gophermat/internal/service/accrualclient"
	"avgys-gophermat/internal/service/auth"
	"avgys-gophermat/internal/service/orders"
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

	//Db
	dbConnection, err := db.NewDB(done, &db.Config{ConnectionString: cfg.DBConnectionString})

	if err != nil {
		err = fmt.Errorf("error initializing db: %w", err)
		return nil, err
	}

	//Repos
	authRepo := repository.NewAuthRepository(dbConnection)
	orderRepo := repository.NewOrderRepository(dbConnection)

	// Services
	authService := auth.NewAuthService(authRepo)
	accrualService := accrualclient.NewAccrualService(done, cfg)
	orderService := orders.NewOrderService(orderRepo, accrualService)

	//Background processors
	processor.NewProcessor(done, orderService, accrualService, traceLogger)

	h := endpoints.New(authService, orderService)

	return h, nil
}
