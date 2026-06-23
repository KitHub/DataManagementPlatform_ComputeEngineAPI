package component

import (
	"context"
	"sync"
)

var initComponentInstance *InitComponent
var onceForInitComponentInstance sync.Once = sync.Once{}

type InitCallback func(ctx context.Context) error

type InitComponent struct {
	initCallbacks      []InitCallback
	lock               *sync.Mutex
	initCallbacksDoing bool
}

func NewInitComponent(ctx context.Context, secretID string, secretKey string, bucketUrl string) *InitComponent {
	onceForInitComponentInstance.Do(func() {
		initComponentInstance = &InitComponent{
			initCallbacks: make([]InitCallback, 0),
			lock:          &sync.Mutex{},
		}
	})
	return initComponentInstance
}
