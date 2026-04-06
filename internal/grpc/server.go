package grpc

import (
	"context"
	"errors"
	"net"
	"net/url"

	pb "github.com/iolshn04/go-musthave-shortened-url/internal/grpc/proto"
	"github.com/iolshn04/go-musthave-shortened-url/internal/middlewares"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedShortenerServiceServer

	Service *service.ShortenerService
	Repo    repository.Repository
	BaseURL string
	Logger  *zap.Logger
}

// ===== HANDLERS =====

func (s *Server) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	userID, ok := middlewares.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	id, err := s.Service.Shorten(ctx, userID, req.GetUrl())
	if err != nil {
		var e repository.ErrAlreadyExistsWithID
		if errors.As(err, &e) {
			full, _ := url.JoinPath(s.BaseURL, e.ExistingID)
			return &pb.URLShortenResponse{Result: full}, nil
		}

		s.Logger.Error("shorten failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	full, _ := url.JoinPath(s.BaseURL, id)
	return &pb.URLShortenResponse{Result: full}, nil
}

func (s *Server) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	original, err := s.Service.GetOriginal(ctx, req.GetId())

	switch {
	case errors.Is(err, repository.ErrNotFound):
		return nil, status.Error(codes.NotFound, "not found")

	case errors.Is(err, repository.ErrDeleted):
		return nil, status.Error(codes.FailedPrecondition, "deleted")

	case err != nil:
		s.Logger.Error("expand failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.URLExpandResponse{Result: original}, nil
}

func (s *Server) ListUserURLs(ctx context.Context, _ *pb.ListUserURLsRequest) (*pb.UserURLsResponse, error) {
	userID, ok := middlewares.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	urls, err := s.Repo.GetByUser(ctx, userID)
	if err != nil {
		return &pb.UserURLsResponse{}, nil
	}

	resp := make([]*pb.URLData, 0, len(urls))

	for _, u := range urls {
		full, _ := url.JoinPath(s.BaseURL, u.ShortURL)

		resp = append(resp, &pb.URLData{
			ShortUrl:    full,
			OriginalUrl: u.OriginalURL,
		})
	}

	return &pb.UserURLsResponse{Url: resp}, nil
}

func AuthInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}

		auth := md.Get("authorization")
		if len(auth) == 0 {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}

		if secret != "" && auth[0] != secret {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, middlewares.UserIDKey, auth[0])

		return handler(ctx, req)
	}
}

func TrustedSubnetInterceptor(subnet string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		if info.FullMethod == "/shortener.ShortenerService/GetStats" {

			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Error(codes.PermissionDenied, "forbidden")
			}

			ip := ""
			if v := md.Get("x-real-ip"); len(v) > 0 {
				ip = v[0]
			}

			if !isIPTrustedIP(ip, subnet) {
				return nil, status.Error(codes.PermissionDenied, "forbidden")
			}
		}

		return handler(ctx, req)
	}
}

func isIPTrustedIP(ipStr, subnet string) bool {
	if subnet == "" {
		return false
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return false
	}

	return ipNet.Contains(ip)
}
