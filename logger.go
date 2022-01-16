package socketify

import "fmt"

type Logger interface {
	Error(args ...interface{})
}

type logger struct {
}

func (logger) Error(args ...interface{}) {
	fmt.Println(args...)
}
