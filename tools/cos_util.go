package tools

import (
	"context"
	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
	"os"
)

const (
	SECRETID  = "AKIDfCVBdqpb9SXNyMfmiV5SgjZBaoskRhlc"
	SECRETKET = "g7LL7qNjspa6NPCk4AE3dy4uVrzsWu6Y"
)

func UploadFileToCos(filepath string, filename string, bucket string) error {
	u, _ := url.Parse("https://static-1304359512.cos.ap-guangzhou.myqcloud.com")
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			//SecretId
			SecretID: os.Getenv(SECRETID),
			//SecretKey
			SecretKey: os.Getenv(SECRETKET),
		},
	})
	key := bucket + "/" + filename
	_, _, err := client.Object.Upload(
		context.Background(), key, filepath, nil,
	)
	if err != nil {
		return err
	}
	return nil
}
