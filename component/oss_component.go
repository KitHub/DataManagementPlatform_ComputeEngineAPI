package component

import "context"

type OSSComponent interface {
	PutDataFromFile(ctx context.Context, bucket string, key string, contentFilePath string) error
	PutDataFromMemory(ctx context.Context, bucket string, key string, content []byte) error
	GetDataToFile(ctx context.Context, bucket string, key string, contentFilePath string) error
	GetDataToMemory(ctx context.Context, bucket string, key string) ([]byte, error)
}
