package socketify

import (
	"encoding/json"
)

type EmptyInput struct{}

type mapper interface {
	Handle(message json.RawMessage) error
}

type dataMapper[T any] struct {
	handler func(T) error
}

func (u dataMapper[T]) Handle(data json.RawMessage) error {
	var t T

	if _, ok := any(t).(EmptyInput); !ok {
		err := json.Unmarshal(data, &t)
		if err != nil {
			return err
		}
	}

	return u.handler(t)
}

func DataMapper[T any](handler func(T) error) dataMapper[T] {
	return dataMapper[T]{handler: handler}
}
