package logger

import "context"

type Logger interface {
	Log(ctx context.Context, message interface{}, args ...interface{})
	Print(v ...interface{})
}
