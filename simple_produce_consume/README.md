> Please make sure to prepare the [pre-requisites](../README.md) ready before following this tasks.

# Summary

In order to demonstrate that a Kafka consumer will only be consumed by single `goroutine` regardless of how many partitions a topic has, please follow these steps:

## Step 1

Create a topic with 4 partitions using this command

```shell
$ docker exec -it kafka kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 4 --topic poc-test
```

## Step 2

Make sure the topic is there with the expected partitions, you can use this command

```bash
$ docker exec -it kafka kafka-topics --zookeeper localhost:2181 --describe --topic poc-test
```

Or you can navigate to http://localhost:3030/ on your browser and choose **Kafka Topics UI**

## Step 3

Start a new terminal window, change directory to root of the project.

Run the consumer first

```bash
$ go run simple_produce_consume/*.go consumer
```

#### Optional flags:

- `-broker` kafka broker host, default: localhost
- `-topic` kafka topic name, default to poc-test
- `-cg` consumer group name, default to testCG

### Step 4

Now start another terminal and change directory to the project's root.

Run the producer

```bash
$ go run simple_produce_consume/*.go producer
```

#### Optional flags:

- `-broker` kafka broker host, default: localhost
- `-topic` kafka topic name, default to poc-test

You can stop the producer using `Ctrl+C`

![](https://media.giphy.com/media/ehKrjRCEacBBS/giphy.gif)

