package app

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/shipa988/banner_rotator/internal/domain/usecase"
	"github.com/shipa988/banner_rotator/pkg/request-util"
	"io"
)

var _ usecase.Logger = (*Logger)(nil)

type Logger struct {
	logger *zerolog.Logger
}

func (l Logger) Print(v ...interface{}) {
	l.logger.Log().Msgf("",v)
}

func (l Logger) Log(ctx context.Context, message string, args ...interface{}) {
	l.logger.Log().Str("Request id", request_util.GetRequestID(ctx)).Msgf(message, args...)
}

func NewLogger(logWriter io.Writer) (*Logger, error) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(logWriter).With().Timestamp().Logger()
	return &Logger{logger: &logger}, nil
}
