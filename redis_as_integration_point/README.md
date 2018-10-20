> Please make sure to prepare the [pre-requisites](../README.md) ready before following this tasks.

# Summary

In this case, we're going to demonstrate a blocking HTTP call that's going to wait for data from redis while asynchronously publish message to kafka so that the consumer will provide the data into the pre-determined redis key. At the end, we're going to try load test the HTTP server to see whether the solution is good enough.

## Step 1

Create a topic with 4 partitions using this command

> If you have done this before then SKIP this step

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
$ go run redis_as_integration_point/*.go consumer
```

#### Optional flags:

- `-broker` kafka broker host, default: localhost
- `-topic` kafka topic name, default to poc-test
- `-cg` consumer group name, default to testCG
- `-redisAddr` redis address, default to localhost:6379
- `-minD` minimum synthetic delay duration, default to 0s
- `-maxD` maximum synthetic delay duration, default to 0s
- `-async` whether to process each message from kafka asynchronously or not, default to true

### Step 4

Now start another terminal and change directory to the project's root.

Run the http server

```bash
$ go run redis_as_integration_point/*.go http
```

#### Optional flags:

- `-broker` kafka broker host, default: localhost
- `-topic` kafka topic name, default to poc-test
- `-redisAddr` redis address, default to localhost:6379

You can stop the http server using `Ctrl+C`

### Step 5

Now you can start the load test using custom vegeta load test, by using this command:

#### flags

- `-rps` request per second, default: 100
- `-dur` how long the test going to be performed, default to 5s
- `-vars` total num of variance on the request, default to 1000000

```bash
$ go run load_test/http.go -dur=1m -rps=100 | tee result.bin | vegeta report
```

You can NOT use standard [vegetta flags](https://github.com/tsenart/vegeta#usage-manual), but the result (i.e. result.bin) can be inspected by normal vegetta command (e.g. `vegeta report`)

Above we're trying to load test in 1 minutes, rate on 100/s

Using my machine MacBook Pro (15-inch, 2017), CPU 2,9 GHz Intel Core i7, RAM 16 GB 2133 MHz LPDDR3; the result is good enough.

![](https://media.giphy.com/media/11sBLVxNs7v6WA/giphy.gif)

BUT

Change slightly the number of RPS to 500, there are a lot of 500 errors since redis connections exhaustion. Success rate is ~40%

```bash
$ go run load_test/http.go -dur=1m -rps=500 | tee result.bin | vegeta report
```

![](https://media.giphy.com/media/DzIIiyZvSdzxu/giphy.gif)

And...

It's even worse if consumer is set using `-async=false`, since the consumer only utilise 1 goroutine to process instead of exclusive for every message

Restart the consumer using this:

```base
$ go run redis_as_integration_point/*.go consumer -async=false
```

Re-run the load test

```bash
$ go run load_test/http.go -dur=1m -rps=100 | tee result.bin | vegeta report
```

The result was very very dissapointing

Success rate is **8.22%**, with average latency **~11 secs**, 99th percentile **~13 secs**

![](https://media.giphy.com/media/1BXa2alBjrCXC/giphy.gif)
