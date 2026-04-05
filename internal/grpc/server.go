package grpc

import (
	"context"
	"errors"
	"net/url"

	pb "github.com/iolshn04/go-musthave-shortened-url/internal/grpc/proto"
	"github.com/iolshn04/go-musthave-shortened-url/internal/repository"
	"github.com/iolshn04/go-musthave-shortened-url/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	pb.UnimplementedShortenerServiceServer

	Service *service.ShortenerService
	Repo    repository.Repository
	BaseURL string
	Secret  string
}

func (s *Server) getUserID(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no metadata")
	}

	auth := md.Get("authorization")
	if len(auth) == 0 {
		return "", status.Error(codes.Unauthenticated, "no auth header")
	}

	return auth[0], nil
}

func (s *Server) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	id, err := s.Service.Shorten(ctx, userID, req.Url)
	if err != nil {
		var e repository.ErrAlreadyExistsWithID
		if errors.As(err, &e) {
			full, _ := url.JoinPath(s.BaseURL, e.ExistingID)
			return &pb.URLShortenResponse{Result: full}, nil
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	full, _ := url.JoinPath(s.BaseURL, id)
	return &pb.URLShortenResponse{Result: full}, nil
}

func (s *Server) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	original, err := s.Service.GetOriginal(ctx, req.Id)

	if errors.Is(err, repository.ErrNotFound) {
		return nil, status.Error(codes.NotFound, "not found")
	}
	if errors.Is(err, repository.ErrDeleted) {
		return nil, status.Error(codes.FailedPrecondition, "deleted")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.URLExpandResponse{Result: original}, nil
}

func (s *Server) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, err := s.getUserID(ctx)
	if err != nil {
		return nil, err
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
