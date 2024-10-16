package utils

import "fmt"

func GetMapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	return keys
}

func GetMapValues[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, value := range m {
		values = append(values, value)
	}

	return values
}

func MapSlice[S any, T any](s []S, mapper func(S) T) []T {
	mapped := make([]T, 0, len(s))
	for _, val := range s {
		mapped = append(mapped, mapper(val))
	}

	return mapped
}

func ToStringSlice[S any](s []S) []string {
	return MapSlice(s, func(value S) string { return fmt.Sprint(value) })
}
