package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/bxcodec/faker"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	producer    *kafka.Producer
	pubSub      *redis.PubSub
	inqWaitMap  map[string](chan *ResponseMessage)
	inqMapMutex sync.RWMutex
)

func StartHttpServer() {
	inqWaitMap = make(map[string](chan *ResponseMessage))
	inqMapMutex = sync.RWMutex{}

	redisOpts := &redis.Options{
		Addr:         redisAddress,
		Password:     "", // no password set
		DB:           0,  // use default DB
		PoolSize:     1,
		MinIdleConns: 1,
		PoolTimeout:  1 * time.Second,
	}

	initProducer()
	initRedis(redisOpts)
	initPubsubRedis()

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

func initPubsubRedis() {
	pubSub = redisCli.Subscribe(redisChannel)

	// Wait for confirmation that subscription is created before publishing anything.
	_, err := pubSub.Receive()
	if err != nil {
		panic(err)
	}

	// Spawn goroutine to listen
	go listenFromRedisChannel()
}

func inquiry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	msgId := vars["id"]

	mBytes, err := buildMessage(msgId)
	if err != nil {
		panic(err)
	}

	respCh := registerWaitChannel(msgId)
	log.Infof("Publishing message [%s] to kafka", msgId)
	err = pushToKafka(mBytes)

	if err != nil {
		log.Errorf("Delivery failed of [%s]: %v", msgId, err)
		http.Error(w, "Can't publish to kafka !", 500)
		return
	}
	log.Infof("Successfully delivered to kafka [%s]", msgId)

	// Careful to use time.After since underlying Timer is not garbage collected
	timeout := time.After(10 * time.Second)

	select {
	case resp := <-respCh:
		log.WithField("ID", resp.ID).WithField("Amount", resp.Amount).Infoln("Response received from redis")

		resBytes, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(resBytes)
	case <-timeout:
		http.Error(w, "Too long waiting", 500)
	}
}

func buildMessage(id string) ([]byte, error) {
	message := RequestMessage{}
	err := faker.FakeData(&message)
	if err != nil {
		return nil, err
	}
	message.ID = id
	message.Timestamp = time.Now()
	mBytes, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}

	return mBytes, nil
}

func registerWaitChannel(id string) chan *ResponseMessage {
	ch := make(chan *ResponseMessage, 1)
	inqMapMutex.Lock()
	inqWaitMap[id] = ch
	inqMapMutex.Unlock()
	return ch
}

func pushToKafka(msg []byte) error {
	delivery := make(chan kafka.Event)
	defer close(delivery)

	_ = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          msg,
	}, delivery)

	ev := <-delivery
	km := ev.(*kafka.Message)

	if km.TopicPartition.Error != nil {
		return km.TopicPartition.Error
	}
	return nil
}

func listenFromRedisChannel() {
	log.WithField("Channel", redisChannel).Info("Start subscribing to redis")
	// Go channel which receives messages.
	ch := pubSub.Channel()

	for msg := range ch {
		resp := ResponseMessage{}
		err := json.Unmarshal([]byte(msg.Payload), &resp)
		if err != nil {
			log.Errorf("JSON Parse Error: %v\n", err)
			continue
		}

		tenSecsAgo := time.Now().Add(-10 * time.Second)
		if resp.Timestamp.Before(tenSecsAgo) {
			log.WithField("ID", resp.ID).WithField("Timestamp", resp.Timestamp).Debugf("SKIP message: too long ago")
			continue
		}

		inqMapMutex.RLock()
		ch := inqWaitMap[resp.ID]
		inqMapMutex.RUnlock()
		if ch == nil {
			log.WithField("ID", resp.ID).Debugf("SKIP message: not relevant")
			continue
		}

		ch <- &resp
		inqMapMutex.Lock()
		delete(inqWaitMap, resp.ID)
		inqMapMutex.Unlock()
	}
}
