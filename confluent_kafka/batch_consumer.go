package confluent_kafka

import (
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/qudj/fly_lib/tools"
	"time"
)

const (
	QuiteReceive  = 1
	QuiteConsumer = 2
	QuiteDealFunc = 3
)

type (
	BatchHandler func([]*kafka.Message) error

	BatchConsumer struct {
		Consumer    *kafka.Consumer
		Topic       string
		Quite       int
		Pause       bool
		MsgChan     chan *kafka.Message
		BatchCommit int
		AutoCommit  bool
	}
)

func InitConsumer(cfg MQConfig) *BatchConsumer {
	kc := &BatchConsumer{}
	var (
		consumer *kafka.Consumer
	)
	err := tools.Retry(3, 3*time.Second, func() error {
		var err error
		consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  cfg.Host,
			"group.id":           cfg.GroupId,
			"auto.offset.reset":  "latest",
			"enable.auto.commit": !cfg.ManualCommit,
		})
		return err
	})
	if err != nil {
		panic(err)
	}
	err = consumer.SubscribeTopics([]string{cfg.Topic}, nil)
	if err != nil {
		panic(err)
	}
	kc.Consumer = consumer
	kc.Topic = cfg.Topic
	kc.BatchCommit = cfg.BatchCommit
	kc.AutoCommit = !cfg.ManualCommit
	kc.MsgChan = make(chan *kafka.Message)
	return kc
}

func (c *BatchConsumer) Start(handler BatchHandler) {
	go c.ReceiveMsg(handler)
	for {
		if c.Quite > 0 {
			break
		}
		if c.Pause {
			time.Sleep(time.Second)
			continue
		}
		msg, err := c.Consumer.ReadMessage(time.Second)
		if err != nil {
			if IsKafkaTimeOut(err) {
				continue
			}
			tools.Logger().Error(fmt.Sprintf("ReadMessage error:%v", err))
			continue
		}
		c.MsgChan <- msg
	}
	close(c.MsgChan)
	c.Quite = QuiteConsumer
}

func (c *BatchConsumer) ReceiveMsg(handler BatchHandler) {
	sleepTime := time.Second
	ticker := time.Tick(sleepTime)
	needSend := false
	msgArray := make([]*kafka.Message, 0)
	partitionMap := map[int32]kafka.TopicPartition{}
	for {
		if c.Quite == QuiteConsumer {
			break
		}
		select {
		case <-ticker:
			needSend = true
		case msg, ok := <-c.MsgChan:
			if !ok {
				continue
			}
			msgArray = append(msgArray, msg)
			partitionMap[msg.TopicPartition.Partition] = msg.TopicPartition
		}
		msgNum := len(msgArray)
		if msgNum >= c.BatchCommit {
			needSend = true
		}
		if msgNum > 0 && needSend {
			err := handler(msgArray)
			if err != nil {
				tools.Logger().Errorf("batch deal msg error: %v", err)
				c.Pause = true
				continue
			}

			if !c.AutoCommit {
				//批量提交代码
				commitPartitions := make([]kafka.TopicPartition, 0)
				for _, partition := range partitionMap {
					commitPartitions = append(commitPartitions, partition)
				}
				if _, err := c.Consumer.CommitOffsets(commitPartitions); err != nil {
					tools.Logger().Errorf("Failed to commit offsets: %v", err)
					c.Pause = true
					continue
				}
			}

			msgArray = make([]*kafka.Message, 0)
			partitionMap = map[int32]kafka.TopicPartition{}
			c.Pause = false
			needSend = false
			ticker = time.Tick(sleepTime)
		}
	}
	c.Quite = QuiteDealFunc
	tools.Logger().Infof("DealMsg has quit")
	return
}

func (c *BatchConsumer) Shutdown() {
	c.Quite = QuiteReceive
	for {
		if c.Quite == QuiteDealFunc {
			break
		}
		time.Sleep(time.Second)
	}
}

func IsKafkaTimeOut(err error) bool {
	var kafkaError kafka.Error
	ok := errors.As(err, &kafkaError)
	if ok {
		if kafkaError.Code() == kafka.ErrTimedOut {
			return true
		}
	}
	return false
}
