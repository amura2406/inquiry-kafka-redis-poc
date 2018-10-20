package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis"
	colorable "github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
)

var (
	broker        string
	topic         string
	consumerGroup string
	redisAddress  string
	asyncConsume  bool
	delayMin      time.Duration
	delayMax      time.Duration
	redisCli      *redis.Client
)

type RequestMessage struct {
	ID   string `faker:"username"`
	Name string `faker:"name"`
	Date string `faker:"date"`
}

type ResponseMessage struct {
	ID       string  `faker:"username"`
	Name     string  `faker:"name"`
	Date     string  `faker:"date"`
	Currency string  `faker:"currency"`
	Amount   float64 `faker:"amount"`
}

func init() {
	log.SetOutput(colorable.NewColorableStdout())
}

func main() {
	consumerSubCmd := flag.NewFlagSet("consumer", flag.ExitOnError)
	httpSubCmd := flag.NewFlagSet("http", flag.ExitOnError)

	consumerSubCmd.StringVar(&broker, "broker", "localhost", "Kafka broker address")
	consumerSubCmd.StringVar(&topic, "topic", "poc-test", "Name of the topic")
	consumerSubCmd.StringVar(&consumerGroup, "cg", "testCG", "Name of the Kafka consumer group")
	consumerSubCmd.StringVar(&redisAddress, "redisAddr", "localhost:6379", "Redis address")
	consumerSubCmd.DurationVar(&delayMin, "minD", 0*time.Second, "Minimum synthetic delay duration")
	consumerSubCmd.DurationVar(&delayMax, "maxD", 0*time.Second, "Maximum synthetic delay duration")
	consumerSubCmd.BoolVar(&asyncConsume, "async", true, "Whether to process each message from kafka asynchronously or not")

	httpSubCmd.StringVar(&broker, "broker", "localhost", "Kafka broker address")
	httpSubCmd.StringVar(&topic, "topic", "poc-test", "Name of the topic")
	httpSubCmd.StringVar(&redisAddress, "redisAddr", "localhost:6379", "Redis address")

	if len(os.Args) < 2 {
		fmt.Println("consumer or http sub command is required !")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "consumer":
		consumerSubCmd.Parse(os.Args[2:])
	case "http":
		httpSubCmd.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if consumerSubCmd.Parsed() {
		StartConsumer()
	}
	if httpSubCmd.Parsed() {
		StartHttpServer()
	}
}

func initRedis(options *redis.Options) {
	log.Infof("Initiating redis...")
	redisCli = redis.NewClient(options)

	err := redisCli.Ping().Err()
	if err != nil {
		panic(err)
	}
}
