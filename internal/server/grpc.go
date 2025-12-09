package server

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/davidbadelllab/go-microservice-grpc-2023/internal/service"
	pb "github.com/davidbadelllab/go-microservice-grpc-2023/proto"
)

// UserServer implements the gRPC UserService
type UserServer struct {
	pb.UnimplementedUserServiceServer
	userService *service.UserService
}

// NewUserServer creates a new UserServer instance
func NewUserServer(userService *service.UserService) *UserServer {
	return &UserServer{
		userService: userService,
	}
}

// CreateUser creates a new user
func (s *UserServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	slog.Info("creating user",
		slog.String("email", req.Email),
		slog.String("name", req.Name))

	user, err := s.userService.CreateUser(ctx, req.Email, req.Name)
	if err != nil {
		slog.Error("failed to create user", slog.String("error", err.Error()))
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &pb.UserResponse{
		User: &pb.User{
			Id:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Unix(),
			UpdatedAt: user.UpdatedAt.Unix(),
		},
	}, nil
}

// GetUser retrieves a user by ID
func (s *UserServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	slog.Info("getting user", slog.Int64("id", req.Id))

	user, err := s.userService.GetUser(ctx, req.Id)
	if err != nil {
		slog.Error("failed to get user", slog.String("error", err.Error()))
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.UserResponse{
		User: &pb.User{
			Id:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Unix(),
			UpdatedAt: user.UpdatedAt.Unix(),
		},
	}, nil
}

// ListUsers lists all users with pagination
func (s *UserServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	slog.Info("listing users",
		slog.Int("page", int(req.Page)),
		slog.Int("page_size", int(req.PageSize)))

	// Go 1.21: min/max built-in functions
	pageSize := min(int(req.PageSize), 100)
	page := max(int(req.Page), 1)

	users, total, err := s.userService.ListUsers(ctx, page, pageSize)
	if err != nil {
		slog.Error("failed to list users", slog.String("error", err.Error()))
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUsers[i] = &pb.User{
			Id:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Unix(),
			UpdatedAt: user.UpdatedAt.Unix(),
		}
	}

	return &pb.ListUsersResponse{
		Users: pbUsers,
		Total: int32(total),
	}, nil
}

// UpdateUser updates an existing user
func (s *UserServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	slog.Info("updating user",
		slog.Int64("id", req.Id),
		slog.String("email", req.Email),
		slog.String("name", req.Name))

	user, err := s.userService.UpdateUser(ctx, req.Id, req.Email, req.Name)
	if err != nil {
		slog.Error("failed to update user", slog.String("error", err.Error()))
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	return &pb.UserResponse{
		User: &pb.User{
			Id:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Unix(),
			UpdatedAt: user.UpdatedAt.Unix(),
		},
	}, nil
}

// DeleteUser deletes a user by ID
func (s *UserServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.Empty, error) {
	slog.Info("deleting user", slog.Int64("id", req.Id))

	err := s.userService.DeleteUser(ctx, req.Id)
	if err != nil {
		slog.Error("failed to delete user", slog.String("error", err.Error()))
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &pb.Empty{}, nil
}

// LoggingInterceptor logs all gRPC requests
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	slog.Info("grpc request",
		slog.String("method", info.FullMethod),
		slog.Duration("duration", time.Since(start)),
		slog.Bool("error", err != nil))

	return resp, err
}

// MetricsInterceptor records metrics for gRPC requests
func MetricsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	// Record metrics (Prometheus)
	duration := time.Since(start).Seconds()
	_ = duration // TODO: Record to Prometheus histogram

	return resp, err
}

// RecoveryInterceptor recovers from panics in gRPC handlers
func RecoveryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic recovered",
				slog.String("method", info.FullMethod),
				slog.Any("panic", r))
			err = status.Errorf(codes.Internal, "internal server error")
		}
	}()

	return handler(ctx, req)
}
