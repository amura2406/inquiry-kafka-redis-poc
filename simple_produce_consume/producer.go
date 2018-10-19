package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bxcodec/faker"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
)

type Message struct {
	ID   string `faker:"username"`
	Name string `faker:"name"`
	Date string `faker:"date"`
}

type Producer struct {
	Broker   string
	Topic    string
	Producer *kafka.Producer
	Running  bool
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <broker> <topic>\n", os.Args[0])
		os.Exit(1)
	}

	log.SetOutput(colorable.NewColorableStdout())
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		log.Println()
		log.Println(sig)
		done <- true
	}()

	broker := os.Args[1]
	topic := os.Args[2]

	producer := Producer{}
	producer.Init(broker)
	defer producer.Close()
	producer.StartProducing(topic)

	<-done
	log.Infoln("Exiting...")
	producer.Shutdown()
}

func (p *Producer) Init(broker string) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		panic(err)
	}

	p.Producer = producer
}

func (p *Producer) StartProducing(topic string) {
	if p.Running {
		return
	}
	log.Infoln("Start producing indefinitely...")
	p.Running = true

	go func() {
		for p.Running {
			m := Message{}
			err := faker.FakeData(&m)
			if err != nil {
				panic(err)
			}

			mBytes, err := json.Marshal(m)
			if err != nil {
				panic(err)
			}
			err = p.Producer.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
				Value:          mBytes,
			}, nil)

			log.Infof("Produce ID:%s", m.ID)
			time.Sleep(1 * time.Second)
		}
	}()
}

func (p *Producer) Close() {
	p.Producer.Close()
}

func (p *Producer) Shutdown() {
	p.Running = false
}
