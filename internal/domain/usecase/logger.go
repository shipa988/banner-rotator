package usecase

import "context"

type Logger interface {
	Log(ctx context.Context,message string,args...interface{})
}
