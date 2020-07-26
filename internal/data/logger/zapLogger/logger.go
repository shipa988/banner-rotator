package zapLogger

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/shipa988/banner_rotator/internal/data/logger"
	util "github.com/shipa988/banner_rotator/pkg/request-util"
	"io"
	"strings"
	"time"
)

var _ logger.Logger = (*Logger)(nil)
var sb strings.Builder

type Logger struct {
	logger  *zerolog.Logger
	isDebug bool
}

func NewLogger(logWriter io.Writer, isDebug bool) (*Logger, error) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logger := zerolog.New(logWriter).With().Timestamp().Logger()
	return &Logger{logger: &logger, isDebug: isDebug}, nil
}

func (l Logger) Print(v ...interface{}) {
	if l.isDebug {
		if len(v) == 6 && v[0] == "sql" {
			v2, _ := v[2].(int64)
			dr := time.Duration(v2)
			l.logger.Log().Msgf("type: %v; stack:%v; sqlDuration: %v; query: %v; param: %v; return:%v", v[0], v[1], dr, v[3], v[4], v[5])
			return
		}
		for _, mess := range v {
			str, _ := mess.(string)
			sb.WriteString(" " + str + ";")
		}
		l.logger.Log().Msg(sb.String())
		sb.Reset()
	}
}

func (l Logger) log(ctx context.Context, message string, args ...interface{}) {
	l.logger.Log().Str("Request id", util.GetRequestID(ctx)).Msgf(message, args...)
}

func (l *Logger) Log(ctx context.Context, message interface{}, args ...interface{}) {
	switch mess := message.(type) {
	case error:
		if l.isDebug {
			err, ok := errors.Cause(mess).(stackTracer)
			if ok {
				st := err.StackTrace()
				l.log(ctx, fmt.Sprintf("%+v", st), args...)
				return
			}
		}
		l.log(ctx, mess.Error(), args...)
	case string:
		l.log(ctx, mess, args...)
	default:
		l.log(ctx, fmt.Sprintf("debug message %v has unknown type %v", message, mess), args...)
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}
