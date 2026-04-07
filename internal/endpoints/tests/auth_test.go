package endpoint_tests

import (
	"avgys-gophermat/internal/endpoints"
	"avgys-gophermat/internal/endpoints/tests/mocks"
	"avgys-gophermat/internal/model/requests"
	"avgys-gophermat/internal/service/auth"
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type authSuite struct {
	suite.Suite
	ctrl     *gomock.Controller
	authMock *mocks.MockAuthService
}

func (s *authSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.authMock = mocks.NewMockAuthService(s.ctrl)
}

func (s *authSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *authSuite) newEndpoints() *endpoints.Endpoints {
	return endpoints.New(s.authMock, nil, nil)
}

func (s *authSuite) TestRegisterOK() {
	s.authMock.EXPECT().
		Register(gomock.Any(), gomock.AssignableToTypeOf(&requests.UserRq{})).
		DoAndReturn(func(ctx context.Context, user *requests.UserRq) (*auth.TokenClaims, error) {
			s.Equal("alice", user.Login)
			s.Equal("secret", user.Password)
			return auth.NewToken(1, user.Login), nil
		})

	reqBody := []byte(`{"login":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(reqBody))
	resp := httptest.NewRecorder()

	s.newEndpoints().Register(resp, req)

	s.Equal(http.StatusOK, resp.Code)
	cookies := resp.Result().Cookies()
	if s.Len(cookies, 1) {
		s.Equal(auth.CookieName, cookies[0].Name)
	}
}

func (s *authSuite) TestRegisterBadJSON() {
	reqBody := []byte(`{"login":"alice"`) // invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(reqBody))
	resp := httptest.NewRecorder()

	s.newEndpoints().Register(resp, req)

	s.Equal(http.StatusBadRequest, resp.Code)
}

func (s *authSuite) TestRegisterConflict() {
	s.authMock.EXPECT().
		Register(gomock.Any(), gomock.AssignableToTypeOf(&requests.UserRq{})).
		Return(nil, auth.ErrUserAlreadyExists)

	reqBody := []byte(`{"login":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(reqBody))
	resp := httptest.NewRecorder()

	s.newEndpoints().Register(resp, req)

	s.Equal(http.StatusConflict, resp.Code)
}

func (s *authSuite) TestRegisterInternalError() {
	s.authMock.EXPECT().
		Register(gomock.Any(), gomock.AssignableToTypeOf(&requests.UserRq{})).
		Return(nil, errors.New("boom"))

	reqBody := []byte(`{"login":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(reqBody))
	resp := httptest.NewRecorder()

	s.newEndpoints().Register(resp, req)

	s.Equal(http.StatusInternalServerError, resp.Code)
}

func (s *authSuite) TestLoginOK() {
	s.authMock.EXPECT().
		Login(gomock.Any(), gomock.AssignableToTypeOf(&requests.UserRq{})).
		DoAndReturn(func(ctx context.Context, user *requests.UserRq) (*auth.TokenClaims, error) {
			s.Equal("alice", user.Login)
			s.Equal("secret", user.Password)
			return auth.NewToken(1, user.Login), nil
		})

	reqBody := []byte(`{"login":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(reqBody))
	resp := httptest.NewRecorder()

	s.newEndpoints().Login(resp, req)

	s.Equal(http.StatusOK, resp.Code)
	cookies := resp.Result().Cookies()
	if s.Len(cookies, 1) {
		s.Equal(auth.CookieName, cookies[0].Name)
	}
}

func (s *authSuite) TestLoginBadJSON() {
	reqBody := []byte(`{"login":"alice"`) // invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(reqBody))
	resp := httptest.NewRecorder()

	s.newEndpoints().Login(resp, req)

	s.Equal(http.StatusBadRequest, resp.Code)
}

func (s *authSuite) TestLoginUnauthorized() {
	s.authMock.EXPECT().
		Login(gomock.Any(), gomock.AssignableToTypeOf(&requests.UserRq{})).
		Return(nil, auth.ErrUnauthorized)

	reqBody := []byte(`{"login":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(reqBody))
	resp := httptest.NewRecorder()

	s.newEndpoints().Login(resp, req)

	s.Equal(http.StatusUnauthorized, resp.Code)
}

func (s *authSuite) TestLoginInternalError() {
	s.authMock.EXPECT().
		Login(gomock.Any(), gomock.AssignableToTypeOf(&requests.UserRq{})).
		Return(nil, errors.New("boom"))

	reqBody := []byte(`{"login":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(reqBody))
	resp := httptest.NewRecorder()

	s.newEndpoints().Login(resp, req)

	s.Equal(http.StatusInternalServerError, resp.Code)
}

func TestAuthSuite(t *testing.T) {
	suite.Run(t, new(authSuite))
}
