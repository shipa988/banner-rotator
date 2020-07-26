package grpcservice

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/shipa988/banner_rotator/cmd/rotator/internal/domain/usecase"
	"github.com/shipa988/banner_rotator/internal/data/logger"
	util "github.com/shipa988/banner_rotator/pkg/request-util"
)

var _ BannerRotatorServiceServer = (*GRPCServer)(nil)

const pageURLCookie = "page_url"
const defaultTimer = time.Second * 5

var headers = []string{
	"Cookie",
}

type GRPCServer struct {
	logger   logger.Logger
	wg       *sync.WaitGroup
	rotator  usecase.Rotator
	server   *grpc.Server
	gwserver *http.Server
}

func (s *GRPCServer) SubscribeOnEvents(req *StatRequest, srv BannerRotatorService_SubscribeOnEventsServer) error {
	page_url := util.GetAuthorizationToken(srv.Context())
	stats := []*Stat{}

	stop := false
	for !stop {
		select {
		case <-time.After(defaultTimer):
			//s.lock.RLock()
			slots, err := s.rotator.GetPageStat(page_url)
			if err != nil {
				return status.Error(codes.Aborted, err.Error())
			}
			for slot, banners := range slots {
				for banner, events := range banners {
					for group, event := range events {
						stat := &Stat{
							PageUrl:          page_url,
							SlotId:           uint64(slot.InnerID),
							BannerId:         uint64(banner.InnerID),
							GroupDescription: group.Description,
							ClickCount:       uint64(event.Clicks),
							ShowCount:        uint64(event.Shows),
						}
						stats = append(stats, stat)
					}
				}
			}
			//s.lock.RUnlock()
			resp := &StatResponse{Time: ptypes.TimestampNow(), Stat: stats}
			if err := srv.Send(resp); err != nil {
				return status.Error(codes.Aborted, err.Error())
			}
			if err := srv.Send(resp); err != nil {
				return status.Error(codes.Aborted, err.Error())
				stop = true
			}
		case <-srv.Context().Done():
			s.logger.Log(srv.Context(), "stats listener disconnected")
			stop = true
		}
	}

	return nil
}

func (s *GRPCServer) RegisterSlot(ctx context.Context, req *RegisterSlotRequest) (*empty.Empty, error) {
	page_url := util.GetAuthorizationToken(ctx)
	err := s.rotator.AddSlot(page_url, uint(req.GetSlotId()), req.GetSlotDescription())
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) RegisterBanner(ctx context.Context, req *RegisterBannerRequest) (*empty.Empty, error) {
	page_url := util.GetAuthorizationToken(ctx)
	err := s.rotator.AddBannerToSlot(page_url, uint(req.GetSlotId()), uint(req.GetBannerId()), req.GetBannerDescription())
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) DeleteBanner(ctx context.Context, req *DeleteBannerRequest) (*empty.Empty, error) {
	page_url := util.GetAuthorizationToken(ctx)
	err := s.rotator.DeleteBannerFromSlot(page_url, uint(req.GetSlotId()), uint(req.GetBannerId()))
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) DeleteSlot(ctx context.Context, req *DeleteSlotRequest) (*empty.Empty, error) {
	page_url := util.GetAuthorizationToken(ctx)
	err := s.rotator.DeleteSlot(page_url, uint(req.GetSlotId()))
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) DeleteAllSlots(ctx context.Context, req *DeleteAllSlotsRequest) (*empty.Empty, error) {
	page_url := util.GetAuthorizationToken(ctx)
	err := s.rotator.DeleteAllSlots(page_url)
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) DeleteAllBanners(ctx context.Context, req *DeleteAllBannersRequest) (*empty.Empty, error) {
	page_url := util.GetAuthorizationToken(ctx)
	err := s.rotator.DeleteAllBannersFormSlot(page_url, uint(req.GetSlotId()))
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) ClickEvent(ctx context.Context, req *ClickRequest) (*empty.Empty, error) {
	fmt.Println(req)
	page_url := util.GetAuthorizationToken(ctx)
	err := s.rotator.ClickByBanner(page_url, uint(req.GetSlotId()), uint(req.GetBannerId()), uint(req.GetUserAge()), req.GetUserSex())
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) GetNextBanner(ctx context.Context, req *GetNextBannerRequest) (*GetNextBannerResponse, error) {
	page_url := util.GetAuthorizationToken(ctx)
	banner, err := s.rotator.GetNextBanner(page_url, uint(req.GetSlotId()), uint(req.GetUserAge()), req.GetUserSex())
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	resp := GetNextBannerResponse{BannerId: uint64(banner)}
	return &resp, nil
}

func NewGRPCServer(wg *sync.WaitGroup, logger logger.Logger, rotator usecase.Rotator) *GRPCServer {
	return &GRPCServer{
		logger:  logger,
		wg:      wg,
		rotator: rotator}
}

func (s *GRPCServer) ServeGW(addr string, addrgw string) {
	defer s.wg.Done()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.logger.Log(ctx, "starting grpc gateway server at %v", addrgw)

	mux := runtime.NewServeMux(
		runtime.WithMetadata(injectHeadersIntoMetadata),
	)
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := RegisterBannerRotatorServiceHandlerFromEndpoint(ctx, mux, addr, opts)
	if err != nil {
		s.logger.Log(ctx, errors.Wrapf(err, "can't register gateway from grpc endpoint at addr %v", addr))
		return
	}
	s.gwserver = &http.Server{
		Addr:    addrgw,
		Handler: mux,
	}

	if err := s.gwserver.ListenAndServe(); err != http.ErrServerClosed {
		s.logger.Log(ctx, errors.Wrapf(err, "can't start  grpc gateway server at %v", addrgw))
	}
}

func (s *GRPCServer) StopGWServe() {
	ctx := context.Background()
	s.logger.Log(ctx, "stopping grpc gw server")
	defer s.logger.Log(ctx, "grpc gw stopped")
	if s.gwserver == nil {
		s.logger.Log(ctx, "grpc gw server is nil")
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := s.gwserver.Shutdown(ctx); err != nil {
		s.logger.Log(ctx, "can't stop grpc gw server with error: %v", err)
	}
}

func (s *GRPCServer) PrepareListener(addr string) net.Listener {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Log(nil, "failed to listen: %v", err)
	}
	return lis
}

func (s *GRPCServer) Serve(listener net.Listener) {
	var opts []grpc.ServerOption
	streamingChain := grpc_middleware.ChainStreamServer(s.authstream)
	opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(s.log, s.auth)))
	opts = append(opts, grpc.StreamInterceptor(streamingChain))

	s.server = grpc.NewServer(opts...)
	RegisterBannerRotatorServiceServer(s.server, s)
	if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
		s.logger.Log(nil, "can't start grpc server at %v", listener.Addr().String())
	}
}

func (s *GRPCServer) StopServe() {
	ctx := context.Background()
	s.logger.Log(ctx, "stopping grpc server")
	defer s.logger.Log(ctx, "grpc server stopped")

	s.server.GracefulStop()
}
func (s *GRPCServer) auth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	page_url := ""

	newCtx, err, done := s.authorization(ctx, page_url)
	if done {
		return nil, err
	}
	return handler(newCtx, req)
}

func (s *GRPCServer) authstream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	page_url := ""

	newCtx, err, done := s.authorization(ss.Context(), page_url)
	if done {
		return err
	}
	wrapped := grpc_middleware.WrapServerStream(ss)
	wrapped.WrappedContext = newCtx
	return handler(srv, wrapped)
}

func (s *GRPCServer) authorization(ctx context.Context, page_url string) (context.Context, error, bool) {
	//auth middleware in learning purpose it is setting Cookie header to value:page_url-this value using across all endpoints in grpc server as AuthToken
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if cookies, ok := md["cookie"]; ok {
			for _, cookie := range cookies {
				dict := strings.SplitN(strings.SplitN(cookie, ";", 2)[0], "=", 2)
				if dict[0] == pageURLCookie {
					//read first match
					page_url = dict[1]
					break
				}
			}
		}
		if page_url == "" {
			return nil, status.Error(codes.NotFound, "page_url cookie is not set"), true
		}
	}

	newCtx := util.SetAutorizationToken(ctx, page_url)
	return newCtx, nil, false
}

func (s *GRPCServer) log(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	start := time.Now()

	clientIP := "unknown"
	if p, ok := peer.FromContext(ctx); ok {
		clientIP = p.Addr.String()
	}

	useragent := ""
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ua, ok := md["user-agent"]; ok {
			useragent = strings.Join(ua, ",")
		}
	}
	ri := util.NewHTTPReqInfo(clientIP, start, info.FullMethod, "", "proto3", useragent)

	newctx := util.SetRequestID(ctx)
	h, err := handler(newctx, req)
	// after executing rpc
	st, _ := status.FromError(err)
	ri.Code = st.Code().String()
	ri.Latency = time.Since(start)
	//logging
	s.logRequest(newctx, ri)
	return h, err
}

func (cs *GRPCServer) logRequest(ctx context.Context, ri *util.HTTPReqInfo) {
	cs.logger.Log(ctx, "%s [%s] %s %s %s %s %s [%s]", ri.IP, ri.Start, ri.Method, ri.Path, ri.Httpver, ri.Code, ri.Latency, ri.Useragent)
}

func injectHeadersIntoMetadata(ctx context.Context, req *http.Request) metadata.MD {
	fmt.Println(req)
	pairs := make([]string, 0, len(headers))
	for _, h := range headers {
		if v := req.Header.Get(h); len(v) > 0 {
			pairs = append(pairs, h, v)
		}
	}
	return metadata.Pairs(pairs...)
}
