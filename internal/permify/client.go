package permify

import (
	"fmt"

	permify_grpc "github.com/Permify/permify-go/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Client *permify_grpc.Client

func InitClient() error {
	var err error
	Client, err = permify_grpc.NewClient(
		permify_grpc.Config{
			Endpoint: "permify:3478",
		},
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("permify client init error: %v", err)
	}

	return nil
}
