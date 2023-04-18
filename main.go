package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"slices"

	"google.golang.org/grpc"
	pb "github.com/davidbadelllab/go-microservice-grpc-2023/proto"
)

type server struct {
	pb.UnimplementedUserServiceServer
}

func main() {
	// Go 1.21: Built-in structured logging with slog
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("starting gRPC server",
		slog.String("service", "user-service"),
		slog.String("version", "1.0.0"),
		slog.Int("port", 50051))

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("failed to listen", slog.String("error", err.Error()))
		os.Exit(1)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &server{})

	slog.Info("server listening", slog.String("address", lis.Addr().String()))

	if err := s.Serve(lis); err != nil {
		slog.Error("failed to serve", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	slog.Info("creating user",
		slog.String("email", req.Email),
		slog.String("name", req.Name))

	user := &pb.User{
		Id:    generateID(),
		Email: req.Email,
		Name:  req.Name,
	}

	return &pb.UserResponse{User: user}, nil
}

func (s *server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// Go 1.21: slices package
	users := []string{"user1", "user2", "user3"}

	// Sort users
	slices.Sort(users)

	// Check if contains
	hasUser := slices.Contains(users, "user1")

	slog.Info("listing users",
		slog.Int("count", len(users)),
		slog.Bool("hasUser1", hasUser))

	// Go 1.21: min/max built-in functions
	maxID := max(1, 2, 3)
	minID := min(1, 2, 3)

	slog.Debug("id range", slog.Int("min", minID), slog.Int("max", maxID))

	return &pb.ListUsersResponse{
		Users: []*pb.User{
			{Id: 1, Email: "user1@example.com", Name: "User 1"},
			{Id: 2, Email: "user2@example.com", Name: "User 2"},
		},
	}, nil
}

func generateID() int64 {
	return 1
}
