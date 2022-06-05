package tools

import (
	"context"
	"github.com/tencentyun/cos-go-sdk-v5"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

const (
	SECRETID  = "AKIDfCVBdqpb9SXNyMfmiV5SgjZBaoskRhlc"
	SECRETKET = "g7LL7qNjspa6NPCk4AE3dy4uVrzsWu6Y"
)

// UploadFileToCos
// @author zia
// @Description: 上传文件到cos
// @param filepath 文件路径 | 如 /go/src/public/1.jpg
// @param filename	1.jpg
// @param bucket	桶名称 video | img
// @return error 	错误
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

// UploadFileObjectToCos
// @author zia
// @Description: 上传文件对象
// @param file 文件对象
// @param filename 文件名
// @param bucket 桶名称 video | img
// @param contentType 文件类型 image|jpeg
// @return error
func UploadFileObjectToCos(file multipart.File, filename string, bucket string, contentType string) error {
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
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}
	_, err := client.Object.Put(context.Background(), key, file, opt)
	if err != nil {
		return err
	}
	return nil
}
