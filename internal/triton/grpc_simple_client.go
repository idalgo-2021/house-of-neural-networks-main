package triton

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	triton "house-of-neural-networks/pkg/api/triton2"

	"google.golang.org/grpc"
)

const (
	inputSize  = 16
	outputSize = 16
)

type Flags struct {
	ModelName    string
	ModelVersion string
	BatchSize    int
	URL          string
}

type TritonConfig struct {
	Host string `env:"TRITON_HOST" env-default:"localhost"`
	Port string `env:"TRITON_PORT" env-default:"8001"`
}

type TritonClient struct {
	Client triton.GRPCInferenceServiceClient
}

func ServerLiveRequest(client triton.GRPCInferenceServiceClient) *triton.ServerLiveResponse {
	// Create context for our request with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverLiveRequest := triton.ServerLiveRequest{}
	// Submit ServerLive request to server
	serverLiveResponse, err := client.ServerLive(ctx, &serverLiveRequest)
	if err != nil {
		log.Fatalf("Couldn't get server live: %v", err)
	}
	return serverLiveResponse
}

func ServerReadyRequest(client triton.GRPCInferenceServiceClient) *triton.ServerReadyResponse {
	// Create context for our request with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverReadyRequest := triton.ServerReadyRequest{}
	// Submit ServerReady request to server
	serverReadyResponse, err := client.ServerReady(ctx, &serverReadyRequest)
	if err != nil {
		log.Fatalf("Couldn't get server ready: %v", err)
	}
	return serverReadyResponse
}

func ModelMetadataRequest(client triton.GRPCInferenceServiceClient, modelName string, modelVersion string) *triton.ModelMetadataResponse {
	// Create context for our request with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create status request for a given model
	modelMetadataRequest := triton.ModelMetadataRequest{
		Name:    modelName,
		Version: modelVersion,
	}
	// Submit modelMetadata request to server
	modelMetadataResponse, err := client.ModelMetadata(ctx, &modelMetadataRequest)
	if err != nil {
		log.Fatalf("Couldn't get server model metadata: %v", err)
	}
	return modelMetadataResponse
}

func ModelInferRequest(client triton.GRPCInferenceServiceClient, rawInput [][]byte, modelName string, modelVersion string) *triton.ModelInferResponse {
	// Create context for our request with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create request input tensors
	inferInputs := []*triton.ModelInferRequest_InferInputTensor{
		&triton.ModelInferRequest_InferInputTensor{
			Name:     "INPUT0",
			Datatype: "INT32",
			Shape:    []int64{1, 16},
		},
		&triton.ModelInferRequest_InferInputTensor{
			Name:     "INPUT1",
			Datatype: "INT32",
			Shape:    []int64{1, 16},
		},
	}

	// Create request input output tensors
	inferOutputs := []*triton.ModelInferRequest_InferRequestedOutputTensor{
		&triton.ModelInferRequest_InferRequestedOutputTensor{
			Name: "OUTPUT0",
		},
		&triton.ModelInferRequest_InferRequestedOutputTensor{
			Name: "OUTPUT1",
		},
	}

	// Create inference request for specific model/version
	modelInferRequest := triton.ModelInferRequest{
		ModelName:    modelName,
		ModelVersion: modelVersion,
		Inputs:       inferInputs,
		Outputs:      inferOutputs,
	}

	modelInferRequest.RawInputContents = append(modelInferRequest.RawInputContents, rawInput[0])
	modelInferRequest.RawInputContents = append(modelInferRequest.RawInputContents, rawInput[1])

	// Submit inference request to server
	modelInferResponse, err := client.ModelInfer(ctx, &modelInferRequest)
	if err != nil {
		log.Fatalf("Error processing InferRequest: %v", err)
	}
	return modelInferResponse
}

func ModelReadyRequest(client triton.GRPCInferenceServiceClient, modelName string, modelVersion string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	modelReadyRequest := triton.ModelReadyRequest{
		Name:    modelName,
		Version: modelVersion,
	}
	modelReadyResponse, err := client.ModelReady(ctx, &modelReadyRequest)
	if err != nil {
		return false, err
	}
	return modelReadyResponse.GetReady(), nil
}

func LoadModelRequest(client triton.GRPCInferenceServiceClient, modelName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.RepositoryModelLoad(ctx, &triton.RepositoryModelLoadRequest{
		ModelName: modelName,
	})
	if err != nil {
		return err
	}
	return nil
}

func UnloadModelRequest(client triton.GRPCInferenceServiceClient, modelName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.RepositoryModelUnload(ctx, &triton.RepositoryModelUnloadRequest{
		ModelName: modelName,
	})
	if err != nil {
		return err
	}
	return nil
}

// Convert int32 input data into raw bytes (assumes Little Endian)
func Preprocess(inputs [][]int32) [][]byte {
	inputData0 := inputs[0]
	inputData1 := inputs[1]

	var inputBytes0 []byte
	var inputBytes1 []byte
	// Temp variable to hold our converted int32 -> []byte
	bs := make([]byte, 4)
	for i := 0; i < inputSize; i++ {
		binary.LittleEndian.PutUint32(bs, uint32(inputData0[i]))
		inputBytes0 = append(inputBytes0, bs...)
		binary.LittleEndian.PutUint32(bs, uint32(inputData1[i]))
		inputBytes1 = append(inputBytes1, bs...)
	}

	return [][]byte{inputBytes0, inputBytes1}
}

func BytesToInt32(rawInputs [][]byte) [][]int32 {
	inputBytes0 := rawInputs[0]
	inputBytes1 := rawInputs[1]

	outputData0 := make([]int32, outputSize)
	outputData1 := make([]int32, outputSize)
	for i := 0; i < outputSize; i++ {
		outputData0[i] = readInt32(inputBytes0[i*4 : i*4+4])
		outputData1[i] = readInt32(inputBytes1[i*4 : i*4+4])
	}
	return [][]int32{outputData0, outputData1}
}

// Convert slice of 4 bytes to int32 (assumes Little Endian)
func readInt32(fourBytes []byte) int32 {
	buf := bytes.NewBuffer(fourBytes)
	var retval int32
	binary.Read(buf, binary.LittleEndian, &retval)
	return retval
}

// Convert output's raw bytes into int32 data (assumes Little Endian)
func Postprocess(inferResponse *triton.ModelInferResponse) [][]int32 {
	outputBytes0 := inferResponse.RawOutputContents[0]
	outputBytes1 := inferResponse.RawOutputContents[1]

	outputData0 := make([]int32, outputSize)
	outputData1 := make([]int32, outputSize)
	for i := 0; i < outputSize; i++ {
		outputData0[i] = readInt32(outputBytes0[i*4 : i*4+4])
		outputData1[i] = readInt32(outputBytes1[i*4 : i*4+4])
	}
	return [][]int32{outputData0, outputData1}
}

func NewTritonClient(host, port string) *TritonClient {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", host, port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldn't connect to endpoint %s:%s: %v", host, port, err)
	}
	client := triton.NewGRPCInferenceServiceClient(conn)
	return &TritonClient{Client: client}
}
