package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"sync"
	"time"

	"github.com/bxcodec/faker"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	colorable "github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
)

var (
	broker       string
	topic        string
	redisAddress string
	redisChannel string
	redisCli     *redis.Client
	producer     *kafka.Producer
	pubSub       *redis.PubSub
	inqWaitMap   map[string](chan *Response)
	inqMapMutex  sync.RWMutex
)

type Message struct {
	ID        string `faker:"username"`
	Name      string `faker:"name"`
	Date      string `faker:"date"`
	Timestamp time.Time
}

type Response struct {
	ID        string
	Name      string
	Date      string
	Currency  string
	Amount    float64
	Timestamp *time.Time
}

func main() {
	flag.StringVar(&broker, "broker", "localhost", "Kafka broker address")
	flag.StringVar(&topic, "topic", "test", "Name of the topic")
	flag.StringVar(&redisAddress, "redisAddr", "localhost:6379", "Redis address")
	flag.StringVar(&redisChannel, "redisChan", "inquiry-response", "Redis channel to listen")

	flag.Parse()
	inqWaitMap = make(map[string](chan *Response))
	inqMapMutex = sync.RWMutex{}

	log.SetOutput(colorable.NewColorableStdout())

	initProducer()
	initRedis()

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

func initRedis() {
	log.Infof("Initiating redis...")
	redisCli = redis.NewClient(&redis.Options{
		Addr:         redisAddress,
		Password:     "", // no password set
		DB:           0,  // use default DB
		PoolSize:     1,
		MinIdleConns: 1,
		PoolTimeout:  1 * time.Second,
	})

	err := redisCli.Ping().Err()
	if err != nil {
		panic(err)
	}

	pubSub = redisCli.Subscribe(redisChannel)

	// Wait for confirmation that subscription is created before publishing anything.
	_, err = pubSub.Receive()
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
	message := Message{}
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

func registerWaitChannel(id string) chan *Response {
	ch := make(chan *Response, 1)
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
		resp := Response{}
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
