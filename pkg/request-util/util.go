package requestutil

import (
	"context"
	"net"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	LayoutISO     = "2006-01-02 15:04:05"
	LayoutDateISO = "2006-01-02"
)

const (
	RequestID = contextKey("RequestID")
	PageURL   = contextKey("PageURL")
)

type contextKey string

func GetAuthorizationToken(ctx context.Context) (userID string) {
	if ctx == nil {
		return
	}
	userID, _ = ctx.Value(PageURL).(string)
	return
}

func SetAutorizationToken(ctx context.Context, userID string) context.Context {
	if len(GetAuthorizationToken(ctx)) == 0 {
		return context.WithValue(ctx, PageURL, userID)
	}
	return ctx
}

func GetRequestID(ctx context.Context) (reqID string) {
	if ctx == nil {
		return
	}
	reqID, _ = ctx.Value(RequestID).(string)
	return
}

func SetRequestID(ctx context.Context) context.Context {
	if len(GetRequestID(ctx)) == 0 {
		reqid := uuid.NewV4()
		return context.WithValue(ctx, RequestID, reqid.String())
	}
	return ctx
}

type HTTPReqInfo struct {
	IP, Start, Method, Path, Httpver, Code, Useragent string
	Latency                                           time.Duration
}

func NewHTTPReqInfo(addr string, start time.Time, method, path, httpver string, useragent string) *HTTPReqInfo {
	return &HTTPReqInfo{
		IP:        ipFromHostPort(addr),
		Start:     start.Format(LayoutISO),
		Method:    method,
		Path:      path,
		Httpver:   httpver,
		Useragent: useragent,
	}
}

func ipFromHostPort(hp string) string {
	h, _, err := net.SplitHostPort(hp)
	if err != nil {
		return ""
	}
	if len(h) > 0 && h[0] == '[' {
		return h[1 : len(h)-1]
	}
	return h
}
