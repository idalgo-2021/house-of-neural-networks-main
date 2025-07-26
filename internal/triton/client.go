package triton

//
//import (
//	"context"
//	"fmt"
//	"google.golang.org/grpc"
//	triton "house-of-neural-networks/pkg/api/triton"
//	"log"
//	"time"
//)
//
//type TritonConfig struct {
//	Host string `env:"TRITON_HOST" env-default:"localhost"`
//	Port string `env:"TRITON_PORT" env-default:"8001"`
//}
//
//type Client struct {
//	Client triton.GRPCInferenceServiceClient
//}
//
//func NewClient(host string, port string) *Client {
//	fmt.Println(host, port)
//	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", host, port), grpc.WithInsecure())
//	if err != nil {
//		log.Fatalf("Couldn't connect to endpoint %s%d: %v", host, port, err)
//	}
//
//	grpcServiceClient := triton.NewGRPCInferenceServiceClient(conn)
//
//	return &Client{
//		Client: grpcServiceClient,
//	}
//}
//
//func (c *Client) ModelLoadRequest(modelName string) error {
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	modelLoadRequest := triton.RepositoryModelLoadRequest{
//		ModelName: modelName,
//	}
//
//	_, err := c.Client.RepositoryModelLoad(ctx, &modelLoadRequest)
//	return err
//}
//
//func (c *Client) ModelInferRequest(inputs [][]byte, modelName string, modelVersion string) (*triton.ModelInferResponse, error) {
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	modelInferRequest := triton.ModelInferRequest{
//		ModelName:    modelName,
//		ModelVersion: modelVersion,
//		Inputs: []*triton.ModelInferRequest_InferInputTensor{
//			&triton.ModelInferRequest_InferInputTensor{
//				Name:     "INPUT0",
//				Datatype: "TYPE_INT32",
//				Shape:    []int64{6},
//			},
//			&triton.ModelInferRequest_InferInputTensor{
//				Name:     "INPUT1",
//				Datatype: "TYPE_INT32",
//				Shape:    []int64{6},
//			},
//		},
//		RawInputContents: inputs,
//	}
//
//	modelInferResponse, err := c.Client.ModelInfer(ctx, &modelInferRequest)
//	if err != nil {
//		return nil, err
//	}
//	return modelInferResponse, nil
//}
