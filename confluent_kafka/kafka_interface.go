package confluent_kafka

import (
	"context"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	KafkaStartRetryLimit       = 10
	KafkaAutoOffsetResetLatest = "latest"
)

type (
	FetchConfig struct {
		MinBytes  int `mapstructure:"min_bytes"`
		MaxWaitMs int `mapstructure:"max_wait_ms"`
	}

	MQConfig struct {
		Host         string      `mapstructure:"host"`
		Topic        string      `mapstructure:"topic"`
		Mod          string      `mapstructure:"mod"`
		Token        string      `mapstructure:"token"`
		AssumeRole   string      `mapstructure:"assume_role"`
		GroupId      string      `mapstructure:"group_id"`
		ManualCommit bool        `mapstructure:"manual_commit"`
		BatchCommit  int         `mapstructure:"batch_commit"`
		FetchConfig  FetchConfig `mapstructure:"fetch_config"`
	}

	Consumer interface {
		Handle(MqHandler)
		Shutdown(ctx context.Context)
		ManualCommit(bool)
		BatchCommit(b int)
		SetLogger(logger Logger)
	}

	MqHandler func(*Msg) error
	Msg       struct {
		value     []byte
		key       []byte
		msgType   []byte
		timestamp int64
		partition kafka.TopicPartition
		header    []kafka.Header
	}
)

func (s *Msg) Header() []kafka.Header {
	return s.header
}

func (s *Msg) Topic() string {
	return *s.partition.Topic
}
func (s *Msg) Offset() int64 {
	return int64(s.partition.Offset)
}

func (s *Msg) Partition() int {
	return int(s.partition.Partition)
}

func (s *Msg) Ts() int64 {
	return s.timestamp
}
func (s *Msg) MsgType() string {
	return string(s.msgType)
}
func (s *Msg) ValueString() string {
	return string(s.value)
}

func (s *Msg) KeyString() string {
	return string(s.key)
}

func (s *Msg) Unmarshal(i interface{}) error {
	return json.Unmarshal(s.value, i)
}

func (s *Msg) Value() []byte {
	return s.value
}

func (s *Msg) Key() []byte {
	return s.key
}

func (s *Msg) SetKey(b []byte) *Msg {
	s.key = b
	return s
}

func (s *Msg) SetValue(b []byte) *Msg {
	s.value = b
	return s
}

func (s *Msg) SetMsgType(b []byte) *Msg {
	s.msgType = b
	return s
}

func (s *Msg) SetPartition(p kafka.TopicPartition) *Msg {
	s.partition = p
	return s
}

func (s *Msg) AddHeader(h kafka.Header) *Msg {
	s.header = append(s.header, h)
	return s
}

func NewMsg(value, key []byte, ts int64) *Msg {
	return &Msg{
		value:     value,
		key:       key,
		timestamp: ts,
	}
}
