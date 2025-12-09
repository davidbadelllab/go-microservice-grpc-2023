package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/davidbadelllab/go-microservice-grpc-2023/proto"
)

func main() {
	// Go 1.21: Built-in structured logging with slog
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Connect to gRPC server
	conn, err := grpc.Dial("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		slog.Error("failed to connect", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer conn.Close()

	slog.Info("connected to gRPC server", slog.String("address", "localhost:50051"))

	client := pb.NewUserServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a user
	createResp, err := client.CreateUser(ctx, &pb.CreateUserRequest{
		Email: "user@example.com",
		Name:  "John Doe",
	})
	if err != nil {
		slog.Error("failed to create user", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("user created",
		slog.Int64("id", createResp.User.Id),
		slog.String("email", createResp.User.Email),
		slog.String("name", createResp.User.Name))

	// Get the user
	getResp, err := client.GetUser(ctx, &pb.GetUserRequest{
		Id: createResp.User.Id,
	})
	if err != nil {
		slog.Error("failed to get user", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("user retrieved",
		slog.Int64("id", getResp.User.Id),
		slog.String("email", getResp.User.Email))

	// List users
	listResp, err := client.ListUsers(ctx, &pb.ListUsersRequest{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		slog.Error("failed to list users", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("users listed", slog.Int("count", len(listResp.Users)))
	for _, user := range listResp.Users {
		slog.Info("user",
			slog.Int64("id", user.Id),
			slog.String("email", user.Email),
			slog.String("name", user.Name))
	}

	// Update user
	updateResp, err := client.UpdateUser(ctx, &pb.UpdateUserRequest{
		Id:    createResp.User.Id,
		Email: "updated@example.com",
		Name:  "Jane Doe",
	})
	if err != nil {
		slog.Error("failed to update user", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("user updated",
		slog.Int64("id", updateResp.User.Id),
		slog.String("email", updateResp.User.Email),
		slog.String("name", updateResp.User.Name))

	// Delete user
	_, err = client.DeleteUser(ctx, &pb.DeleteUserRequest{
		Id: createResp.User.Id,
	})
	if err != nil {
		slog.Error("failed to delete user", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("user deleted", slog.Int64("id", createResp.User.Id))
}
