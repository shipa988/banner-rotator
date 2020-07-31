//go:generate  protoc -I ../../../../api/ api.proto --go_out=plugins=grpc:../../../../api --grpc-gateway_out=logtostderr=true:../../../../api
package grpcservice

import (
	"context"
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

	api "github.com/shipa988/banner_rotator/cmd/rotator/api"
	"github.com/shipa988/banner_rotator/cmd/rotator/internal/domain/usecase"
	"github.com/shipa988/banner_rotator/internal/data/logger"
	util "github.com/shipa988/banner_rotator/pkg/request-util"
)

var _ api.BannerRotatorServiceServer = (*GRPCServer)(nil)

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

func (s *GRPCServer) SubscribeOnEvents(req *api.StatRequest, srv api.BannerRotatorService_SubscribeOnEventsServer) error {
	pageURL := util.GetAuthorizationToken(srv.Context())
	stats := []*api.Stat{}

	stop := false
	for !stop {
		select {
		case <-time.After(defaultTimer):
			slots, err := s.rotator.GetPageStat(pageURL)
			if err != nil {
				return status.Error(codes.Aborted, err.Error())
			}
			for slot, banners := range slots {
				for banner, events := range banners {
					for group, event := range events {
						stat := &api.Stat{
							PageUrl:          pageURL,
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
			resp := &api.StatResponse{Time: ptypes.TimestampNow(), Stat: stats}
			if err := srv.Send(resp); err != nil {
				s.logger.Log(srv.Context(), errors.Wrap(err, "stream send error"))
				stop = true
			}
		case <-srv.Context().Done():
			s.logger.Log(srv.Context(), "stats listener disconnected")
			stop = true
		}
	}

	return nil
}

func (s *GRPCServer) RegisterSlot(ctx context.Context, req *api.RegisterSlotRequest) (*empty.Empty, error) {
	pageURL := util.GetAuthorizationToken(ctx)
	err := s.rotator.AddSlot(pageURL, uint(req.GetSlotId()), req.GetSlotDescription())
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) RegisterBanner(ctx context.Context, req *api.RegisterBannerRequest) (*empty.Empty, error) {
	pageURL := util.GetAuthorizationToken(ctx)
	err := s.rotator.AddBannerToSlot(pageURL, uint(req.GetSlotId()), uint(req.GetBannerId()), req.GetBannerDescription())
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) DeleteBanner(ctx context.Context, req *api.DeleteBannerRequest) (*empty.Empty, error) {
	pageURL := util.GetAuthorizationToken(ctx)
	err := s.rotator.DeleteBannerFromSlot(pageURL, uint(req.GetSlotId()), uint(req.GetBannerId()))
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) DeleteSlot(ctx context.Context, req *api.DeleteSlotRequest) (*empty.Empty, error) {
	pageURL := util.GetAuthorizationToken(ctx)
	err := s.rotator.DeleteSlot(pageURL, uint(req.GetSlotId()))
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) DeleteAllSlots(ctx context.Context, req *api.DeleteAllSlotsRequest) (*empty.Empty, error) {
	pageURL := util.GetAuthorizationToken(ctx)
	err := s.rotator.DeleteAllSlots(pageURL)
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) DeleteAllBanners(ctx context.Context, req *api.DeleteAllBannersRequest) (*empty.Empty, error) {
	pageURL := util.GetAuthorizationToken(ctx)
	err := s.rotator.DeleteAllBannersFormSlot(pageURL, uint(req.GetSlotId()))
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) ClickEvent(ctx context.Context, req *api.ClickRequest) (*empty.Empty, error) {
	pageURL := util.GetAuthorizationToken(ctx)
	err := s.rotator.ClickByBanner(pageURL, uint(req.GetSlotId()), uint(req.GetBannerId()), uint(req.GetUserAge()), req.GetUserSex())
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	return &empty.Empty{}, nil
}

func (s *GRPCServer) GetNextBanner(ctx context.Context, req *api.GetNextBannerRequest) (*api.GetNextBannerResponse, error) {
	pageURL := util.GetAuthorizationToken(ctx)
	banner, err := s.rotator.GetNextBanner(pageURL, uint(req.GetSlotId()), uint(req.GetUserAge()), req.GetUserSex())
	if err != nil {
		s.logger.Log(ctx, err)
		return nil, status.Error(codes.Aborted, err.Error())
	}
	s.logger.Log(ctx, "success")
	resp := api.GetNextBannerResponse{BannerId: uint64(banner)}
	return &resp, nil
}

func NewGRPCServer(wg *sync.WaitGroup, logger logger.Logger, rotator usecase.Rotator) *GRPCServer {
	return &GRPCServer{
		logger:  logger,
		wg:      wg,
		rotator: rotator}
}

func (s *GRPCServer) ServeGW(addr string, addrgw string) error {
	defer s.wg.Done()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.logger.Log(ctx, "starting grpc gateway server at %v", addrgw)

	mux := runtime.NewServeMux(
		runtime.WithMetadata(injectHeadersIntoMetadata),
	)
	opts := []grpc.DialOption{grpc.WithInsecure()}

	err := api.RegisterBannerRotatorServiceHandlerFromEndpoint(ctx, mux, addr, opts)
	if err != nil {
		return errors.Wrapf(err, "can't register gateway from grpc endpoint at addr %v", addr)
	}
	s.gwserver = &http.Server{
		Addr:    addrgw,
		Handler: mux,
	}

	if err := s.gwserver.ListenAndServe(); err != http.ErrServerClosed {
		return errors.Wrapf(err, "can't start  grpc gateway server at %v", addrgw)
	}
	return nil
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
		s.logger.Log(context.TODO(), "failed to listen: %v", err)
	}
	return lis
}

func (s *GRPCServer) Serve(listener net.Listener) error {
	s.logger.Log(context.Background(), "starting grpc server at %v", listener.Addr())
	var opts []grpc.ServerOption
	streamingChain := grpc_middleware.ChainStreamServer(s.authstream)
	opts = append(opts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(s.log, s.auth)))
	opts = append(opts, grpc.StreamInterceptor(streamingChain))

	s.server = grpc.NewServer(opts...)
	api.RegisterBannerRotatorServiceServer(s.server, s)
	if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
		return errors.Wrapf(err, "can't start grpc server at %v", listener.Addr().String())
	}
	return nil
}

func (s *GRPCServer) StopServe() {
	ctx := context.Background()
	s.logger.Log(ctx, "stopping grpc server")
	defer s.logger.Log(ctx, "grpc server stopped")

	s.server.GracefulStop()
}
func (s *GRPCServer) auth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	pageURL := ""

	newCtx, err := s.authorization(ctx, pageURL)
	if err != nil {
		return nil, err
	}
	return handler(newCtx, req)
}

func (s *GRPCServer) authstream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	pageURL := ""

	newCtx, err := s.authorization(ss.Context(), pageURL)
	if err != nil {
		return err
	}
	wrapped := grpc_middleware.WrapServerStream(ss)
	wrapped.WrappedContext = newCtx
	return handler(srv, wrapped)
}

func (s *GRPCServer) authorization(ctx context.Context, pageURL string) (context.Context, error) {
	//auth middleware in learning purpose it is setting Cookie header to value:pageURL-this value using across all endpoints in grpc server as AuthToken
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if cookies, ok := md["cookie"]; ok {
			for _, cookie := range cookies {
				dict := strings.SplitN(strings.SplitN(cookie, ";", 2)[0], "=", 2)
				if dict[0] == pageURLCookie {
					//read first match
					pageURL = dict[1]
					break
				}
			}
		}
		if pageURL == "" {
			return nil, status.Error(codes.NotFound, "pageURL cookie is not set")
		}
	}

	newCtx := util.SetAutorizationToken(ctx, pageURL)
	return newCtx, nil
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

func (s *GRPCServer) logRequest(ctx context.Context, ri *util.HTTPReqInfo) {
	s.logger.Log(ctx, "%s [%s] %s %s %s %s %s [%s]", ri.IP, ri.Start, ri.Method, ri.Path, ri.Httpver, ri.Code, ri.Latency, ri.Useragent)
}

func injectHeadersIntoMetadata(ctx context.Context, req *http.Request) metadata.MD {
	pairs := make([]string, 0, len(headers))
	for _, h := range headers {
		if v := req.Header.Get(h); len(v) > 0 {
			pairs = append(pairs, h, v)
		}
	}
	return metadata.Pairs(pairs...)
}
