package confluent_kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/qudj/fly_lib/tools"
	"strings"
	"sync/atomic"
	"time"
)

type KafkaConsumer struct {
	config         *MQConfig
	stop           int32
	running        chan struct{}
	mc             *bool //手动开关
	bc             int
	seekPartitions map[string]kafka.TopicPartition
	logger         Logger
}

func (k *KafkaConsumer) SetLogger(logger Logger) {
	k.logger = logger
}

func (k *KafkaConsumer) BatchCommit(b int) {
	k.bc = b
}

func (k *KafkaConsumer) ManualCommit(r bool) {
	k.mc = &r
}

func (k *KafkaConsumer) manualCommit() bool {
	if k.mc == nil {
		return false
	}
	return *k.mc
}

func (k *KafkaConsumer) Shutdown(ctx context.Context) {
	atomic.StoreInt32(&k.stop, 1)
	<-k.running
}

func (k *KafkaConsumer) Handle(handler MqHandler) {
	consumer := k.init()
	err := consumer.Subscribe(k.config.Topic, nil)
	if err != nil {
		k.logger.Error("cannot init kafka consumer", err)
		return
	}
	k.running = make(chan struct{})
	k.stop = 0
	count := 0
	defer close(k.running)
	for atomic.LoadInt32(&k.stop) == 0 {
		var msg *kafka.Message
		msg, err = consumer.ReadMessage(time.Second / 2)
		if err != nil {
			if kafkaErr, ok := err.(kafka.Error); !ok || kafkaErr.Code() != kafka.ErrTimedOut {
				k.logger.Error("kafka consumer receive message failed", err)
			}
			continue
		}
		if len(msg.Value) == 0 {
			continue
		}

		singleMsg := NewSingleMsgFromKafka(msg)
		err = handler(singleMsg)
		if k.manualCommit() {
			if err != nil {
				k.Seek(msg.TopicPartition, consumer, err)
			} else {
				if count++; count >= k.bc {
					k.Commit(consumer, "batch")
					count = 0
				}
			}
		}
	}
	if k.manualCommit() {
		k.Commit(consumer, "finished")
	}
}

func (k *KafkaConsumer) Seek(partition kafka.TopicPartition, consumer *kafka.Consumer, err error) {
	seekErr := consumer.Seek(partition, 100)
	if seekErr != nil {
		k.logger.Error("kafka seek failed", seekErr)
	}
	k.logger.Info("seek for error,partition:%d,offset:%d,err:%v",
		partition.Partition,
		partition.Offset,
		err)
	k.seekPartition(partition)
}

func (k *KafkaConsumer) Commit(consumer *kafka.Consumer, title string) {
	partitions, err := consumer.Assignment()
	if err != nil {
		k.logger.Error("Kafka Assignment failed", err)
		return
	}
	partitions, err = consumer.Position(partitions)
	if err != nil {
		k.logger.Error("Kafka Position failed", err)
		return
	}
	for i := range partitions {
		k.setSeekPartition(&partitions[i])
	}
	_, err = consumer.CommitOffsets(partitions)
	if err != nil {
		k.logger.Error("Kafka commit failed", err)
		return
	}
	k.resetSeekPartition()
	k.logger.Info("commit,type:%s,partition_count:%d", title, len(partitions))
}

func (k *KafkaConsumer) resetSeekPartition() {
	k.seekPartitions = make(map[string]kafka.TopicPartition)
}

func (k *KafkaConsumer) seekPartition(pts kafka.TopicPartition) {
	key := fmt.Sprintf("%s/%d", pts.Topic, pts.Partition)
	k.seekPartitions[key] = pts
}

func (k *KafkaConsumer) setSeekPartition(pts *kafka.TopicPartition) {
	key := fmt.Sprintf("%s/%d", pts.Topic, pts.Partition)
	if p, ok := k.seekPartitions[key]; ok {
		pts.Offset = p.Offset
	}
}

func (k *KafkaConsumer) init() *kafka.Consumer {
	var (
		consumer  *kafka.Consumer
		waitMaxMs = 100
		minBytes  = 100000000
		cfg       = k.config
	)
	if cfg.FetchConfig.MinBytes > 0 {
		minBytes = cfg.FetchConfig.MinBytes
	}
	if cfg.FetchConfig.MaxWaitMs > 0 {
		waitMaxMs = cfg.FetchConfig.MaxWaitMs
	}
	if k.mc == nil {
		k.ManualCommit(cfg.ManualCommit)
	}
	k.resetSeekPartition()
	if k.bc == 0 {
		if k.bc = cfg.BatchCommit; k.bc == 0 {
			k.bc = 1000
		}
	}

	if k.logger == nil {
		k.logger = DefaultLogger()
	}

	err := tools.Retry(KafkaStartRetryLimit, 3*time.Second, func() error {
		var err error
		consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  cfg.Host,
			"group.id":           cfg.GroupId,
			"auto.offset.reset":  KafkaAutoOffsetResetLatest,
			"fetch.wait.max.ms":  waitMaxMs,
			"fetch.min.bytes":    minBytes,
			"enable.auto.commit": !k.manualCommit(),
		})
		return err
	})
	if err != nil {
		panic(err)
	}
	k.logger.Info("init Consumer,topic:%s,group_id:%s,manual_commit:%v,batch_commit:%d",
		k.config.Topic,
		cfg.GroupId,
		k.manualCommit(),
		k.bc,
	)
	return consumer
}

func InitKafkaConsumer(config *MQConfig) Consumer {
	return &KafkaConsumer{
		stop:   0,
		config: config,
	}
}

func NewSingleMsgFromKafka(msg *kafka.Message) *Msg {
	var (
		msgType []byte
		inArray = map[string]struct{}{
			"TYPE": {}, "EVENT": {},
		}
	)
	for _, v := range msg.Headers {
		if _, ok := inArray[strings.ToUpper(v.Key)]; ok {
			msgType = v.Value
			break
		}
	}

	m := NewMsg(msg.Value, msg.Key, msg.Timestamp.UnixMilli())
	return m.SetPartition(msg.TopicPartition).SetMsgType(msgType)
}
