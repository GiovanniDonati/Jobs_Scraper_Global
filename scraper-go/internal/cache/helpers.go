package cache

func GetAs[T any](c Cache, key string) (T, bool, error) {
	var target T

	found, err := c.Get(key, &target)
	if err != nil {
		var zero T
		return zero, false, err
	}

	return target, found, nil
}

func MustGetAs[T any](c Cache, key string) (T, error) {
	var target T

	_, err := c.Get(key, &target)
	return target, err
}
