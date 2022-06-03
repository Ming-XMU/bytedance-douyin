package tools

import (
	"context"
	"douyin/models"
	"github.com/minio/minio-go/v7"
	"log"
	"mime/multipart"
)

const (
	location = "cn-south-1"
)

// UploadFileObjectToMinio
// @author zia
// @Description: 上传文件file对象
// @param bucketName 桶名称
// @param objectName 文件名称
// @param file	文件对象
// @param contentType 文件数据类型 如：image/jpeg | video/mp4
// @return error
func UploadFileObjectToMinio(bucketName string, objectName string, file *multipart.FileHeader, contentType string) error {
	// 初使化 minio client对象
	mic := models.GetMic()
	// 创建一个桶
	ctx := context.Background()
	err := createBucket(bucketName, mic, ctx)
	if err != nil {
		return err
	}
	//传入文件
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	//上传文件对象
	_, err = mic.PutObject(ctx, bucketName, objectName, src, -1, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// UploadFileToMinio
// @author zia
// @Description: 上传磁盘文件
// @param bucketName	桶名称
// @param objectName	存储到桶中的文件名	| 名称相同文件会覆盖
// @param filepath		文件所在路径(包括文件名) 如 D:/upload/1.jpg
// @param contentType	文件数据类型 如：image/jpeg | video/mp4
// @return error
func UploadFileToMinio(bucketName string, objectName string, filepath string, contentType string) error {
	// 初使化 minio client对象
	mic := models.GetMic()
	// 创建一个桶
	ctx := context.Background()
	err := createBucket(bucketName, mic, ctx)
	if err != nil {
		return err
	}
	// 使用FPutObject上传磁盘文件
	n, err := mic.FPutObject(ctx, bucketName, objectName, filepath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return err
	}
	log.Printf("Successfully uploaded %s of size %d\n", objectName, n.Size)
	return nil
}

// createBucket
// @author zia
// @Description: 创建桶
// @param bucketName 桶名称
// @param mic minio连接
// @param ctx 上下文默认值
// @return error
func createBucket(bucketName string, mic *minio.Client, ctx context.Context) error {
	err := mic.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location, ObjectLocking: false})
	//判断错误原因是否是桶已创建
	if err != nil {
		exists, err1 := mic.BucketExists(ctx, bucketName)
		if err1 == nil && exists {
			//之前已创建桶,正常返回
			log.Printf("We already own %s\n", bucketName)
			return nil
		} else {
			//其他错误
			return err1
		}
	}
	//桶创建成功，正常返回
	log.Printf("Successfully created %s\n", bucketName)
	return nil
}
