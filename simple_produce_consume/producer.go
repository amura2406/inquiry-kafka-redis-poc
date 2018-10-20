package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bxcodec/faker"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
)

var (
	producer *kafka.Producer
	running  bool
)

func StartProducer() {
	log.Infoln("Producer starting...")

	initProducer()
	startProducing()

	log.Infoln("Produces finished !")
}

func initProducer() {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		panic(err)
	}

	producer = p
}

func startProducing() {
	if running {
		return
	}
	log.Infoln("Start producing indefinitely...")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	defer producer.Close()
	run := true

	for run {
		select {
		case sig := <-sigs:
			log.Infof("Caught signal %v: terminating\n", sig)
			run = false
		default:
			m := Message{}
			err := faker.FakeData(&m)
			if err != nil {
				panic(err)
			}

			mBytes, err := json.Marshal(m)
			if err != nil {
				panic(err)
			}
			err = producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
				Value:          mBytes,
			}, nil)

			log.Infof("Produce ID:%s", m.ID)
			time.Sleep(1 * time.Second)
		}
	}
}
