package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/bxcodec/faker"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

var (
	consumer *kafka.Consumer
	random   *rand.Rand
)

func StartConsumer() {
	redisOpts := &redis.Options{
		Addr:         redisAddress,
		Password:     "", // no password set
		DB:           0,  // use default DB
		PoolSize:     50,
		MinIdleConns: 10,
		PoolTimeout:  1 * time.Second,
	}

	initRandom()
	initConsumer()
	initRedis(redisOpts)

	log.Infoln("Listening now...")
	for {
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			var metas map[string]interface{}
			metaBytes, err := json.Marshal(msg.TopicPartition)
			if err != nil {
				panic(err)
			}
			json.Unmarshal(metaBytes, &metas)
			if asyncConsume {
				go processMessage(msg)
			} else {
				processMessage(msg)
			}
		} else {
			// The client will automatically try to recover from all errors.
			log.Errorf("Consumer error: %v (%v)\n", err, msg)
		}
	}

	consumer.Close()
}

func initRandom() {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func initConsumer() {
	log.Infoln("Consumer starting...")
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          consumerGroup,
		"auto.offset.reset": "latest",
	})

	if err != nil {
		panic(err)
	}

	c.SubscribeTopics([]string{topic}, nil)

	consumer = c
}

func processMessage(msg *kafka.Message) {
	reqMsg := RequestMessage{}
	err := json.Unmarshal(msg.Value, &reqMsg)
	if err != nil {
		panic(err)
	}

	resMsg := ResponseMessage{}
	err = faker.FakeData(&resMsg)
	if err != nil {
		panic(err)
	}
	resMsg.ID = reqMsg.ID
	resMsg.Name = reqMsg.Name
	resMsg.Date = reqMsg.Date

	resBytes, err := json.Marshal(resMsg)
	if err != nil {
		panic(err)
	}

	Δ := int64(delayMax) - int64(delayMin)
	if Δ > 0 {
		randΔ := rand.Int63n(Δ)
		delta := time.Duration(int64(delayMin) + randΔ)
		log.WithField("Δ", delta).Infof("Delay...")
		time.Sleep(delta)
	}

	err = redisCli.Set(fmt.Sprintf("id:%s", reqMsg.ID), resBytes, 10*time.Second).Err()
	if err != nil {
		log.Errorf("%v\n", err)
		return
	}
	log.WithField("ID", reqMsg.ID).Infof("Successfully put to redis")
}
