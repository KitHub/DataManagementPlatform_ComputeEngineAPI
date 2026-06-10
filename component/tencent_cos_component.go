package component

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"

	"github.com/tencentyun/cos-go-sdk-v5"
)

var tencentCOSComponentInstance *TencentCOSComponent
var onceForTencentCOSComponentInstance sync.Once = sync.Once{}

type TencentCOSComponent struct {
	cosClient *cos.Client
}

// GetDataToFile implements [OSSComponent].
func (t *TencentCOSComponent) GetDataToFile(ctx context.Context, bucket string, key string, contentFilePath string) error {
	slog.InfoContext(ctx, "get data from tencent cos to file", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFilePath", contentFilePath))

	response, err := t.cosClient.Object.Download(ctx, key, contentFilePath, nil)
	slog.InfoContext(ctx, "get data from tencent cos to file", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFilePath", contentFilePath), slog.Any("response", response), slog.Any("error", err))
	if err != nil {
		slog.ErrorContext(ctx, "get data from tencent cos to file failed", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFilePath", contentFilePath), slog.Any("error", err))
		return err
	}

	if response.StatusCode != 200 {
		slog.ErrorContext(ctx, "get data from tencent cos to file failed", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFilePath", contentFilePath), slog.Any("response", response))
		return errors.New(response.Status)
	}

	slog.InfoContext(ctx, "get data from tencent cos to file done", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFilePath", contentFilePath))
	return nil
}

// GetDataToMemory implements [OSSComponent].
func (t *TencentCOSComponent) GetDataToMemory(ctx context.Context, bucket string, key string) ([]byte, error) {
	slog.InfoContext(ctx, "get data from tencent cos to memory", slog.String("bucket", bucket), slog.String("key", key))

	response, err := t.cosClient.Object.Get(ctx, key, nil)
	slog.InfoContext(ctx, "get data from tencent cos to memory", slog.String("bucket", bucket), slog.String("key", key), slog.Any("response", response), slog.Any("error", err))
	if err != nil {
		slog.ErrorContext(ctx, "get data from tencent cos to memory failed", slog.String("bucket", bucket), slog.String("key", key), slog.Any("error", err))
		return nil, err
	}

	if response.StatusCode != 200 {
		slog.ErrorContext(ctx, "get data from tencent cos to memory failed", slog.String("bucket", bucket), slog.String("key", key), slog.Any("response", response))
		return nil, errors.New(response.Status)
	}

	content, err := io.ReadAll(response.Body)
	defer func() {
		innerErr := response.Body.Close()
		if innerErr != nil {
			slog.ErrorContext(ctx, "get data from tencent cos to memory, close reading stream failed", slog.String("bucket", bucket), slog.String("key", key), slog.Any("error", innerErr))
		}
	}()
	if err != nil {
		slog.ErrorContext(ctx, "get data from tencent cos to memory failed", slog.String("bucket", bucket), slog.String("key", key), slog.Any("error", err))
		return nil, errors.New(response.Status)
	}

	slog.InfoContext(ctx, "get data from tencent cos to memory done", slog.String("bucket", bucket), slog.String("key", key))
	return content, nil
}

// PutDataFromFile implements [OSSComponent].
func (t *TencentCOSComponent) PutDataFromFile(ctx context.Context, bucket string, key string, contentFilePath string) error {
	slog.InfoContext(ctx, "put data to tencent cos", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFilePath", contentFilePath))

	result, response, err := t.cosClient.Object.Upload(ctx, key, contentFilePath, nil)
	slog.InfoContext(ctx, "put data to tencent cos", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFilePath", contentFilePath), slog.Any("result", result), slog.Any("response", response), slog.Any("error", err))
	if err != nil {
		slog.ErrorContext(ctx, "put data to tencent cos failed", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFilePath", contentFilePath), slog.Any("error", err))
		return err
	}

	if response.StatusCode != 200 {
		slog.ErrorContext(ctx, "put data to tencent cos failed", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFilePath", contentFilePath), slog.Any("response", response))
		return errors.New(response.Status)
	}

	slog.InfoContext(ctx, "put data to tencent cos done", slog.String("bucket", bucket), slog.String("key", key), slog.String("contentFilePath", contentFilePath))
	return nil
}

// PutDataFromMemory implements [OSSComponent].
func (t *TencentCOSComponent) PutDataFromMemory(ctx context.Context, bucket string, key string, content []byte) error {
	slog.InfoContext(ctx, "put data to tencent cos", slog.String("bucket", bucket), slog.String("key", key))

	response, err := t.cosClient.Object.Put(ctx, key, bytes.NewReader(content), nil)
	slog.InfoContext(ctx, "put data to tencent cos", slog.String("bucket", bucket), slog.String("key", key), slog.Any("response", response), slog.Any("error", err))
	if err != nil {
		slog.ErrorContext(ctx, "put data to tencent cos failed", slog.String("bucket", bucket), slog.String("key", key), slog.Any("error", err))
		return err
	}

	if response.StatusCode != 200 {
		slog.ErrorContext(ctx, "put data to tencent cos failed", slog.String("bucket", bucket), slog.String("key", key), slog.Any("response", response))
		return errors.New(response.Status)
	}

	slog.InfoContext(ctx, "put data to tencent cos done", slog.String("bucket", bucket), slog.String("key", key))
	return nil
}

func NewTencentCOSComponent(ctx context.Context, secretID string, secretKey string, bucketUrl string) *TencentCOSComponent {
	onceForTencentCOSComponentInstance.Do(func() {
		u, _ := url.Parse(bucketUrl)
		b := &cos.BaseURL{BucketURL: u}
		client := cos.NewClient(b, &http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  secretID,
				SecretKey: secretKey,
			},
		})
		tencentCOSComponentInstance = &TencentCOSComponent{
			cosClient: client,
		}
	})
	return tencentCOSComponentInstance
}
