package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bxcodec/faker"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	MaxTryCount = 20
)

var (
	producer *kafka.Producer
)

func StartHttpServer() {
	redisOpts := &redis.Options{
		Addr:         redisAddress,
		Password:     "", // no password set
		DB:           0,  // use default DB
		PoolSize:     10000,
		MinIdleConns: 100,
		PoolTimeout:  1 * time.Second,
	}

	initProducer()
	initRedis(redisOpts)

	r := mux.NewRouter()
	r.HandleFunc("/inquiry/{id}", inquiry).Methods("GET")

	log.Infof("HTTP server is listening...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}

	log.Infof("Shutting down.")
}

func initProducer() {
	log.WithField("topic", topic).Infof("Creating kafka producer")

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		panic(err)
	}

	producer = p
}

func inquiry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	message := RequestMessage{}
	err := faker.FakeData(&message)
	if err != nil {
		panic(err)
	}
	message.ID = vars["id"]
	mBytes, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}

	delivery := make(chan kafka.Event)

	log.Infof("Publishing message [%s] to kafka", message.ID)
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          mBytes,
	}, delivery)

	ev := <-delivery
	km := ev.(*kafka.Message)

	if km.TopicPartition.Error != nil {
		log.Errorf("Delivery failed of [%s]: %v", message.ID, km.TopicPartition.Error)
		http.Error(w, "Can't publish to kafka !", 500)
	} else {
		log.Infof("Successfully delivered to kafka [%s]", message.ID)

		var resBytes []byte
		tryCount := 0

		for tryCount < MaxTryCount {
			tryCount++
			time.Sleep(500 * time.Millisecond)

			resBytes, err := redisCli.Get(fmt.Sprintf("id:%s", message.ID)).Bytes()
			if err != nil {
				log.Warnf("Redis Err: %v\n", err)
				continue
			}
			res := ResponseMessage{}
			err = json.Unmarshal(resBytes, &res)
			if err != nil {
				panic(err)
			}
			log.WithField("ID", res.ID).WithField("Amount", res.Amount).Infoln("Response received from redis")

			w.Header().Set("Content-Type", "application/json")
			w.Write(resBytes)
		}

		if resBytes == nil {
			http.Error(w, "Server Error", 500)
		}
	}

	close(delivery)
}
