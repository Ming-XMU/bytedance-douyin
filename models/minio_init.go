package models

import (
	"log"
)
import "github.com/minio/minio-go/v7"
import "github.com/minio/minio-go/v7/pkg/credentials"

var mic *minio.Client

func GetMic() *minio.Client {
	return mic
}

func InitMinio(endpoint string, accessKeyID string, secretAccessKey string) {
	useSSL := false
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL})
	if err != nil {
		log.Fatalln("minio连接错误: ", err)
	}
	mic = client
}
