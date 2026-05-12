package inflight

import (
	"fmt"

	"golang.org/x/sync/singleflight"
)

var group singleflight.Group

func Do[T any](key string, fn func() (T, error)) (T, error) {
	v, err, _ := group.Do(key, func() (interface{}, error) {
		return fn()
	})

	if err != nil {
		var zero T
		return zero, err
	}

	result, ok := v.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("invalid singleflight type")
	}

	return result, nil
}
