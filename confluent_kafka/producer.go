package confluent_kafka

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"time"
)

type (
	KafkaConfOption struct {
		Key   string
		Value string
	}

	KafkaMQSender struct {
		topic    string
		producer *kafka.Producer
	}
)

func InitProducer(host, topic string, options ...KafkaConfOption) *KafkaMQSender {
	var err error
	var producer *kafka.Producer

	conf := kafka.ConfigMap{
		"bootstrap.servers": host,
	}
	if len(options) != 0 {
		for _, v := range options {
			conf[v.Key] = v.Value
		}
	}
	producer, err = kafka.NewProducer(&conf)
	if err != nil {
		panic(err)
	}
	return &KafkaMQSender{
		topic:    topic,
		producer: producer,
	}
}

func (k *KafkaMQSender) SendMsgWithKey(msgByte []byte, key string) error {
	msg := NewMsg(msgByte, []byte(key), time.Now().UnixMilli())
	return k.Send(msg)
}

// SendMsg 实现方式
// producer =  InitPushTaskProducer(host, topic)
// producer.SendMsg(msg)
func (k *KafkaMQSender) SendMsg(msgByte []byte) error {
	return k.Send(NewMsg(msgByte, nil, time.Now().UnixMilli()))
}

func (k *KafkaMQSender) Send(msg *Msg) error {
	deliveryChan := make(chan kafka.Event)
	err := k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: kafka.PartitionAny},
		Value:          msg.Value(),
		Key:            msg.Key(),
		Headers:        msg.Header(),
	}, deliveryChan)
	if err != nil {
		return err
	}
	e := <-deliveryChan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		return m.TopicPartition.Error
	}
	close(deliveryChan)
	return nil
}

// AsyncSend 异步推送
func (k *KafkaMQSender) AsyncSend(msg *Msg) error {
	return k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: kafka.PartitionAny},
		Value:          msg.Value(),
		Key:            msg.Key(),
		Headers:        msg.Header(),
	}, nil)
}
