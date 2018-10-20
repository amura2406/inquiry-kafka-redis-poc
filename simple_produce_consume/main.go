package main

import (
	"flag"
	"fmt"
	"os"

	colorable "github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
)

var (
	broker        string
	topic         string
	consumerGroup string
)

type Message struct {
	ID   string `faker:"username"`
	Name string `faker:"name"`
	Date string `faker:"date"`
}

func init() {
	log.SetOutput(colorable.NewColorableStdout())
}

func main() {
	consumerSubCmd := flag.NewFlagSet("consumer", flag.ExitOnError)
	producerSubCmd := flag.NewFlagSet("producer", flag.ExitOnError)

	consumerSubCmd.StringVar(&broker, "broker", "localhost", "Kafka broker address")
	consumerSubCmd.StringVar(&topic, "topic", "test", "Name of the topic")
	consumerSubCmd.StringVar(&consumerGroup, "cg", "testCG", "Name of the Kafka consumer group")

	producerSubCmd.StringVar(&broker, "broker", "localhost", "Kafka broker address")
	producerSubCmd.StringVar(&topic, "topic", "test", "Name of the topic")

	if len(os.Args) < 2 {
		fmt.Println("consumer or producer sub command is required !")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "consumer":
		consumerSubCmd.Parse(os.Args[2:])
	case "producer":
		producerSubCmd.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if consumerSubCmd.Parsed() {
		StartConsumer()
	} else {
		StartProducer()
	}
}
