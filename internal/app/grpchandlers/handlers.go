package grpchandlers

import (
	"context"
	"errors"
	"github.com/egor-zakharov/tiny-url/internal/app/auth"
	"github.com/egor-zakharov/tiny-url/internal/app/config"
	"github.com/egor-zakharov/tiny-url/internal/app/logger"
	"github.com/egor-zakharov/tiny-url/internal/app/service"
	"github.com/egor-zakharov/tiny-url/internal/app/storage"
	pb "github.com/egor-zakharov/tiny-url/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net/url"
)

// ShortenerServer shortener gRPC server
type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer

	service service.Service
	log     *logger.Logger
	auth    *auth.Auth
	config  *config.Config
}

func NewShortenerServer(service service.Service, log *logger.Logger, auth *auth.Auth, config *config.Config) *ShortenerServer {
	return &ShortenerServer{
		service: service,
		log:     log,
		auth:    auth,
		config:  config,
	}
}

func (s *ShortenerServer) Stats(ctx context.Context, _ *pb.StatsRequest) (*pb.StatsResponse, error) {
	response := &pb.StatsResponse{}
	stats, err := s.service.GetStats(ctx)
	if err != nil {
		return response, status.Errorf(codes.InvalidArgument, err.Error())
	}

	response.Urls = int64(stats.Urls)
	response.Users = int64(stats.Users)

	return response, nil
}

func (s *ShortenerServer) GetURL(ctx context.Context, in *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	url, err := s.service.Get(ctx, in.ShortUrl)

	if err != nil {
		if errors.Is(err, storage.ErrDeletedURL) {
			return nil, status.Errorf(codes.DataLoss, err.Error())
		}
		return nil, status.Error(codes.NotFound, storage.ErrNotFound.Error())

	}

	md := metadata.Pairs(
		"Location", url,
	)
	err = grpc.SendHeader(ctx, md)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *ShortenerServer) Auth(ctx context.Context, in *pb.AuthRequest) (*pb.AuthResponse, error) {
	response := &pb.AuthResponse{}
	token, err := auth.BuildToken()
	if err != nil {
		return nil, err
	}
	response.SessionToken = token
	return response, err
}

func (s *ShortenerServer) PostShorten(ctx context.Context, in *pb.PostShortenRequest) (*pb.PostShortenResponse, error) {
	response := &pb.PostShortenResponse{}
	//получаем ID
	ID, err := s.auth.GetIDGrpc(ctx)
	if err != nil {
		s.log.GetLog().Sugar().With("error", err).Error("get ID from token")
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	//валидируем полученное тело
	err = s.service.ValidateURL(in.Url)
	if err != nil {
		s.log.GetLog().Sugar().With("error", err).Error("validate url")
		return nil, status.Error(codes.Internal, err.Error())
	}

	newURL, err := url.Parse(s.config.FlagShortAddr)
	if err != nil {
		s.log.GetLog().Sugar().With("error", err).Error("short addr parse")
		return nil, status.Error(codes.Internal, err.Error())
	}
	//Кодируем и добавляем с сторейдж
	shortURL, err := s.service.Add(ctx, in.Url, ID)
	newURL.Path = shortURL

	if err != nil && !errors.Is(err, storage.ErrConflict) {
		s.log.GetLog().Sugar().With("error", err).Error("add storage")
		return nil, status.Error(codes.Internal, err.Error())
	}
	if errors.Is(err, storage.ErrConflict) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	response.Result = newURL.String()
	return response, nil
}

func (s *ShortenerServer) PostShortenBatch(ctx context.Context, in *pb.PostShortenBatchRequest) (*pb.PostShortenBatchResponse, error) {
	response := &pb.PostShortenBatchResponse{}
	//получаем ID
	ID, err := s.auth.GetIDGrpc(ctx)
	if err != nil {
		s.log.GetLog().Sugar().With("error", err).Error("get ID from token")
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	for _, v := range in.In {
		//валидируем полученное тело`
		err = s.service.ValidateURL(v.OriginalUrl)
		if err != nil {
			s.log.GetLog().Sugar().With("error", err).Error("validate url")
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	urls := make(map[string]string, len(in.In))

	for _, v := range in.In {
		urls[v.CorrelationId] = v.OriginalUrl
	}

	//Кодируем и добавляем с сторейдж
	shortURLs, err := s.service.AddBatch(ctx, urls, ID)
	if err != nil {
		s.log.GetLog().Sugar().With("error", err).Error("add storage")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	newURL, err := url.Parse(s.config.FlagShortAddr)
	if err != nil {
		s.log.GetLog().Sugar().With("error", err).Error("short addr parse")
		return nil, status.Error(codes.Internal, err.Error())
	}

	for corID, shortURL := range shortURLs {
		newURL.Path = shortURL
		response.Out = append(response.Out, &pb.OutShortenBatch{
			CorrelationId: corID,
			ShortUrl:      newURL.String(),
		})
	}
	return response, nil
}

func (s *ShortenerServer) GetAll(ctx context.Context, in *pb.GetAllRequest) (*pb.GetAllResponse, error) {
	response := &pb.GetAllResponse{}
	//получаем ID
	ID, err := s.auth.GetIDGrpc(ctx)
	if err != nil {
		s.log.GetLog().Sugar().With("error", err).Error("get ID from token")
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	urls, err := s.service.GetAll(ctx, ID)
	if err != nil {
		s.log.GetLog().Sugar().With("error", err).Error("service error")
		return nil, status.Error(codes.NotFound, err.Error())
	}

	newURL, err := url.Parse(s.config.FlagShortAddr)
	if err != nil {
		s.log.GetLog().Sugar().With("error", err).Error("short addr parse")
		return nil, status.Error(codes.Internal, err.Error())
	}

	for shortURL, originalURL := range urls {
		newURL.Path = shortURL
		response.Out = append(response.Out, &pb.OutGetAll{
			ShortUrl:    newURL.String(),
			OriginalUrl: originalURL,
		})
	}

	return response, nil
}

func (s *ShortenerServer) DeleteBatch(ctx context.Context, in *pb.DeleteBatchRequest) (*pb.DeleteBatchResponse, error) {
	response := &pb.DeleteBatchResponse{}
	//получаем ID
	ID, err := s.auth.GetIDGrpc(ctx)
	if err != nil {
		s.log.GetLog().Sugar().With("error", err).Error("get ID from token")
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	ch := s.generator(in.ShortUrl)
	_ = s.deleteURL(ch, ID)

	response.Result = "accepted"
	return response, err
}

func (s *ShortenerServer) generator(input []string) chan string {
	ch := make(chan string)
	go func() {
		defer close(ch)
		for _, v := range input {
			ch <- v
		}
	}()

	return ch
}

func (s *ShortenerServer) deleteURL(ch <-chan string, ID string) error {
	var errs error
	for URL := range ch {
		err := s.service.Delete(URL, ID)
		if err != nil {
			errs = err
		}
	}
	return errs
}
