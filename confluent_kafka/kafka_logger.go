package confluent_kafka

import "fmt"

type (
	Logger interface {
		Info(msg string, args ...interface{})
		Error(msg string, err error)
	}
	logger struct {
	}
)

func (*logger) Info(msg string, args ...interface{}) {
	fmt.Println(fmt.Sprintf(msg, args...))
}

func (*logger) Error(msg string, err error) {
	fmt.Println(fmt.Sprintf("%s,%v", msg, err))
}

func DefaultLogger() Logger {
	return &logger{}
}
