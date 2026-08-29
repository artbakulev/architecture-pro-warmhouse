package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"smarthome/proto/remotecontrol"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

type CommandService struct {
	client remotecontrol.RemoteControlCommandServiceClient
}

func NewCommandService(addr string) (*CommandService, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &CommandService{client: remotecontrol.NewRemoteControlCommandServiceClient(conn)}, nil
}

func (s *CommandService) Send(ctx context.Context, sensorID uint64, code string, payload map[string]interface{}) (*remotecontrol.CommandResult, error) {
	inputPayload, err := structpb.NewStruct(payload)
	if err != nil {
		return nil, err
	}
	return s.client.Send(ctx, &remotecontrol.SendCommandRequest{
		Code:           code,
		IdempotencyKey: newIdempotencyKey(),
		SensorId:       sensorID,
		InputPayload:   inputPayload,
	})
}

func (s *CommandService) GetResult(ctx context.Context, id string) (*remotecontrol.CommandResult, error) {
	return s.client.GetResult(ctx, &remotecontrol.CommandId{Id: id})
}

func newIdempotencyKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
