package service

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"house-of-neural-networks/internal/models"
	"house-of-neural-networks/internal/triton"
	client "house-of-neural-networks/pkg/api/message"
	"time"
)

type MessageRepo interface {
	SaveMessage(ctx context.Context, msg models.Message) error
	GetMessages(ctx context.Context, userID, modelID int64) ([]models.Message, error)
	GetModelName(ctx context.Context, model models.Model) (string, error)
	GetVersionNumber(ctx context.Context, version models.Version) (int, error)
}

type TritonClient interface {
	RequestAnswer(ctx context.Context, question string) (string, error)
}

type MessageService struct {
	Repo   MessageRepo
	triton *triton.TritonClient
}

func NewMessageService(repo MessageRepo, triton *triton.TritonClient) *MessageService {
	return &MessageService{Repo: repo, triton: triton}
}

func (s *MessageService) ProcessMessage(ctx context.Context, userID, modelID, versionID int64, inputs []*client.Input) ([]string, error) {
	modelName, err := s.Repo.GetModelName(ctx, models.Model{ID: modelID})
	if err != nil {
		return []string{}, status.Errorf(codes.Internal, "SendMessage: %s", err)
	}
	versionNumber, err := s.Repo.GetVersionNumber(ctx, models.Version{ID: versionID})
	if err != nil {
		return []string{}, status.Errorf(codes.Internal, "SendMessage: %s", err)
	}
	ready, err := triton.ModelReadyRequest(s.triton.Client, modelName, fmt.Sprint(versionNumber))
	if err != nil {
		return []string{}, status.Errorf(codes.Internal, "SendMessage: %s", err)
	}
	if !ready {
		err = triton.LoadModelRequest(s.triton.Client, modelName)
		if err != nil {
			return []string{}, status.Errorf(codes.Internal, "SendMessage: %s", err)
		}
	}
	inputsInt := make([][]int32, 0, len(inputs))
	for i, input := range inputs {
		inputsInt = append(inputsInt, make([]int32, 0, len(input.GetValues())))
		for _, val := range input.GetValues() {
			inputsInt[i] = append(inputsInt[i], val)
		}
	}
	rawInput := triton.Preprocess(inputsInt)
	inferResponse := triton.ModelInferRequest(s.triton.Client, rawInput, modelName, fmt.Sprint(versionNumber))
	outputs := triton.Postprocess(inferResponse)
	outputData0 := outputs[0]
	outputData1 := outputs[1]
	resultsStr := make([]string, 0, len(outputData0)*2)
	for i := 0; i < len(outputData0); i++ {
		resultsStr = append(resultsStr, fmt.Sprintf("%d + %d = %d", inputsInt[0][i], inputsInt[1][i], outputData0[i]))
		resultsStr = append(resultsStr, fmt.Sprintf("%d - %d = %d", inputsInt[0][i], inputsInt[1][i], outputData1[i]))
	}
	err = s.Repo.SaveMessage(ctx, models.Message{
		UserID:    userID,
		ModelID:   modelID,
		VersionID: versionID,
		Input1:    rawInput[0],
		Input2:    rawInput[1],
		Results:   resultsStr,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return []string{}, status.Errorf(codes.Internal, "SendMessage: %s", err)
	}

	return resultsStr, nil
}

func (s *MessageService) GetMessages(ctx context.Context, userID, modelID int64) ([]*client.Message, error) {
	dialog, err := s.Repo.GetMessages(ctx, userID, modelID)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "GetMessages: %s", err)
	}

	var messages = make([]*client.Message, 0)
	for _, msg := range dialog {
		rawInput := [][]byte{msg.Input1, msg.Input2}
		inputsInt := triton.BytesToInt32(rawInput)
		messages = append(messages, &client.Message{
			Id:        msg.ID,
			UserId:    msg.UserID,
			ModelId:   msg.ModelID,
			VersionId: msg.VersionID,
			Inputs: []*client.Input{
				{
					Values: inputsInt[0],
				},
				{
					Values: inputsInt[1],
				},
			},
			Results:   msg.Results,
			CreatedAt: timestamppb.New(msg.CreatedAt),
		})
	}
	return messages, nil
}
