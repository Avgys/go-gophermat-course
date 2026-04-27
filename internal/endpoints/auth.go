package endpoints

import (
	"avgys-gophermat/internal/logger"
	"avgys-gophermat/internal/model/requests"
	"avgys-gophermat/internal/service/auth"
	httphelper "avgys-gophermat/internal/shared/http"
	"context"
	"net/http"

	"github.com/rs/zerolog"
)

func (e *Endpoints) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.FromContext(ctx)

	e.AuthorizeUser(w, r, e.AuthService.Register, traceLogger)
}

func (e *Endpoints) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	traceLogger := logger.FromContext(ctx)

	e.AuthorizeUser(w, r, e.AuthService.Login, traceLogger)
}

type getTokenFn func(ctx context.Context, user *requests.UserRq) (*auth.TokenClaims, error)

func (e *Endpoints) AuthorizeUser(w http.ResponseWriter, r *http.Request, getToken getTokenFn, traceLogger *zerolog.Logger) {

	var user requests.UserRq
	err := httphelper.GetJSONBody(r, &user)
	if httphelper.HandleErr(w, r, err, traceLogger) {
		return
	}

	token, err := getToken(r.Context(), &user)

	if err != nil {

		httphelper.HandleErr(w, r, err, traceLogger)
		return
	}

	if err := token.InjectCookie(w); err != nil {
		httphelper.HandleErr(w, r, err, traceLogger)
		return
	}
	httphelper.WriteResponse(w, nil, http.StatusOK, traceLogger)
}
