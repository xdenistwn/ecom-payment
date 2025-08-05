package grpc

import (
	"context"
	"log"
	"payment/proto/userpb"
	"time"

	"google.golang.org/grpc"
)

type UserClient struct {
	Client userpb.UserServiceClient
}

func NewUserClient() *UserClient {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: " + err.Error())
	}

	client := userpb.NewUserServiceClient(conn)
	return &UserClient{
		Client: client,
	}
}

func (uc *UserClient) GetUserInfoByUserId(ctx context.Context, userID int64) (*userpb.GetUserInfoResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	userInfo, err := uc.Client.GetUserInfoByUserID(ctx, &userpb.GetUserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}
