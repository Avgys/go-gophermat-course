package logger

import (
	"context"
	"io"
	"os"
	"runtime"

	"github.com/rs/zerolog"
)

var funcNameTag string = "func_name_tag"

func NewLogger() (*zerolog.Logger, func(), error) {

	f, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)

	if err != nil {
		return nil, nil, err
	}

	funcName := callerFuncName()

	logger := zerolog.
		New(io.MultiWriter(os.Stderr, f)).
		With().
		Str(funcNameTag, funcName).
		Timestamp().
		Logger()

	return &logger, func() { f.Close() }, nil
}

func NewRequestLogger(ctx context.Context, spanID int64) (*zerolog.Logger, func(), error) {

	log, close, err := NewLogger()

	if err != nil {
		return nil, nil, err
	}

	wrappedLog := log.With().
		Timestamp().
		Ctx(ctx).
		Int64("spanID", spanID).
		Stack().
		Logger()

	return &wrappedLog, close, nil
}

func FromContext(ctx context.Context) *zerolog.Logger {
	funcName := callerFuncName()

	log := FromContext(ctx).With().Str(funcNameTag, funcName).Logger()
	return &log
}

func Middleware(ctx context.Context, name string) *zerolog.Logger {
	log := zerolog.Ctx(ctx).With().Str("middleware", name).Logger()
	return &log
}

func callerFuncName() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "?"
	}
	return runtime.FuncForPC(pc).Name()
}
