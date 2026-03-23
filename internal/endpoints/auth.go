package endpoints

import (
	"avgys-gophermat/internal/logger"
	"avgys-gophermat/internal/model"
	"avgys-gophermat/internal/service/auth"
	httphelper "avgys-gophermat/internal/shared/http"
	"errors"
	"net/http"
)

func (e *Endpoints) Register(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	traceLogger := logger.Endpoint(ctx, "Register")

	// claims, err := auth.GetFromContext(ctx)

	// if err != nil {
	// 	traceLogger.Err(err).Send()
	// 	err = httphelper.NewError(http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	// 	httphelper.WriteError(w, r, err, traceLogger)
	// 	return
	// }

	var user model.UserApi

	if err := getJSONBody(r, &user); err != nil {
		httphelper.WriteError(w, r, err, traceLogger)
		return
	}

	token, err := e.AuthService.Register(ctx, &user)
	if err != nil {
		if errors.Is(err, auth.ErrUserAlreadyExists) {
			err = httphelper.NewError(err.Error(), http.StatusConflict)
		}

		httphelper.WriteError(w, r, err, traceLogger)
		return
	}

	token.InjectCookie(w)
	httphelper.WriteResponse(w, nil, http.StatusOK)
}

func (e *Endpoints) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.Endpoint(ctx, "Register")

	var user model.UserApi

	if err := getJSONBody(r, &user); err != nil {
		httphelper.WriteError(w, r, err, traceLogger)
		return
	}

	token, err := e.AuthService.Login(ctx, &user)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			err = httphelper.NewError(http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}

		httphelper.WriteError(w, r, err, traceLogger)
		return
	}

	token.InjectCookie(w)
	httphelper.WriteResponse(w, nil, http.StatusOK)
}
