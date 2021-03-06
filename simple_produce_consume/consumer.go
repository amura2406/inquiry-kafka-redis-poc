package main

import (
	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
)

func StartConsumer() {
	log.Infoln("Consumer starting...")

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          consumerGroup,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	c.SubscribeTopics([]string{topic}, nil)

	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			var metas map[string]interface{}
			metaBytes, err := json.Marshal(msg.TopicPartition)
			if err != nil {
				panic(err)
			}
			json.Unmarshal(metaBytes, &metas)
			log.WithFields(metas).Infof("Message: %s\n", string(msg.Value))
		} else {
			// The client will automatically try to recover from all errors.
			log.Errorf("Consumer error: %v (%v)\n", err, msg)
		}
	}

	c.Close()
}
