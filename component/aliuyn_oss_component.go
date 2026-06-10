package component

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"sync"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

var aliyunOSSComponentInstance *AliyunOSSComponent
var onceForAliyunOSSComponentInstance sync.Once = sync.Once{}

type AliyunOSSComponent struct {
	ossClient *oss.Client
}

func NewAliyunOSSComponent(ctx context.Context, accessKeyID string, accessKeySecret string, ossRegion string) *AliyunOSSComponent {
	onceForAliyunOSSComponentInstance.Do(func() {
		cfg := oss.LoadDefaultConfig().
			WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret)).
			WithRegion(ossRegion)
		ossClient := oss.NewClient(cfg)
		aliyunOSSComponentInstance = &AliyunOSSComponent{
			ossClient: ossClient,
		}
	})
	return aliyunOSSComponentInstance
}

func (a *AliyunOSSComponent) PutDataFromMemory(ctx context.Context, bucket string, key string, content []byte) error {
	slog.InfoContext(ctx, "put data to aliyun oss", slog.String("bucket", bucket), slog.String("key", key))
	request := &oss.PutObjectRequest{
		Bucket: &bucket,
		Key:    &key,
		Body:   bytes.NewReader(content),
	}
	result, err := a.ossClient.PutObject(ctx, request)
	slog.DebugContext(ctx, "put data to aliyun oss", slog.Any("result", result), slog.Any("error", err))
	if err != nil {
		slog.ErrorContext(ctx, "put data to aliyun oss failed", slog.String("bucket", bucket), slog.String("key", key), slog.Any("error", err))
		return err
	}

	if result.StatusCode != 200 {
		slog.ErrorContext(ctx, "put data to aliyun oss failed", slog.String("bucket", bucket), slog.String("key", key), slog.Any("result", result))
		return err
	}

	slog.InfoContext(ctx, "put data to aliyun oss done", slog.String("bucket", bucket), slog.String("key", key))
	return nil
}

func (a *AliyunOSSComponent) PutDataFromFile(ctx context.Context, bucket string, key string, contentFilePath string) error {
	slog.InfoContext(ctx, "put data to aliyun oss", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFile", contentFilePath))

	request := &oss.PutObjectRequest{
		Bucket: &bucket,
		Key:    &key,
	}
	result, err := a.ossClient.PutObjectFromFile(ctx, request, contentFilePath)
	slog.DebugContext(ctx, "put data to aliyun oss", slog.Any("result", result), slog.Any("error", err))
	if err != nil {
		slog.ErrorContext(ctx, "put data to aliyun oss failed", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFile", contentFilePath), slog.Any("error", err))
		return err
	}

	if result.StatusCode != 200 {
		slog.ErrorContext(ctx, "put data to aliyun oss failed", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFile", contentFilePath), slog.Any("result", result))
		return err
	}

	slog.InfoContext(ctx, "put data to aliyun oss done", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFile", contentFilePath))
	return nil
}

func (a *AliyunOSSComponent) GetDataToMemory(ctx context.Context, bucket string, key string) ([]byte, error) {
	slog.InfoContext(ctx, "get data from aliyun oss", slog.String("bucket", bucket), slog.String("key", key))

	request := &oss.GetObjectRequest{
		Bucket: &bucket,
		Key:    &key,
	}

	result, err := a.ossClient.GetObject(ctx, request)
	slog.DebugContext(ctx, "put data to aliyun oss", slog.Any("result", result), slog.Any("error", err))
	if err != nil {
		slog.InfoContext(ctx, "get data from aliyun oss failed", slog.String("bucket", bucket), slog.String("key", key), slog.Any("error", err))
		return nil, err
	}
	defer func() {
		err := result.Body.Close()
		if err != nil {
			slog.ErrorContext(ctx, "get data from aliyun oss, close data stream failed", slog.Any("error", err))
		}
	}()

	if result.StatusCode != 200 {
		slog.InfoContext(ctx, "get data from aliyun oss failed", slog.String("bucket", bucket), slog.String("key", key), slog.Any("result", result))
		return nil, err
	}

	data, err := io.ReadAll(result.Body)
	if err != nil {
		slog.InfoContext(ctx, "get data from aliyun oss, read data failed", slog.String("bucket", bucket), slog.String("key", key), slog.Any("err", err))
		return nil, err
	}

	slog.InfoContext(ctx, "get data from aliyun oss done", slog.String("bucket", bucket), slog.String("key", key))
	return data, nil
}

func (a *AliyunOSSComponent) GetDataToFile(ctx context.Context, bucket string, key string, contentFilePath string) error {
	slog.InfoContext(ctx, "get data from aliyun oss", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFile", contentFilePath))

	request := &oss.GetObjectRequest{
		Bucket: &bucket,
		Key:    &key,
	}

	result, err := a.ossClient.GetObjectToFile(ctx, request, contentFilePath)
	slog.DebugContext(ctx, "put data to aliyun oss", slog.Any("result", result), slog.String("contentFile", contentFilePath), slog.Any("error", err))
	if err != nil {
		slog.InfoContext(ctx, "get data from aliyun oss failed", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFile", contentFilePath), slog.Any("error", err))
		return err
	}
	defer func() {
		err := result.Body.Close()
		if err != nil {
			slog.ErrorContext(ctx, "get data from aliyun oss, close data stream failed", slog.Any("error", err))
		}
	}()

	if result.StatusCode != 200 {
		slog.InfoContext(ctx, "get data from aliyun oss failed", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFile", contentFilePath), slog.Any("result", result))
		return err
	}

	slog.InfoContext(ctx, "get data from aliyun oss done", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFile", contentFilePath))
	return nil
}
