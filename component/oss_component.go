package component

import "context"

type OSSComponent interface {
	PutData(ctx context.Context, bucket string, key string, content []byte) error
	GetData(ctx context.Context, bucket string, key string) ([]byte, error)
}
