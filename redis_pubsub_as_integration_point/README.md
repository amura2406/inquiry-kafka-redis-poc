> Please make sure to prepare the [pre-requisites](../README.md) ready before following this tasks.

# Summary

In this case, we're going to demonstrate a blocking HTTP call that's going to wait for data from redis by subscribing redis channel to wait for response while asynchronously publish message to kafka so that the consumer will provide the data. At the end, we're going to try load test the HTTP server to see whether the solution is good enough.

## Step 1

Create a topic with 4 partitions using this command

> If you have done this before then SKIP this step

```shell
$ docker exec -it kafka kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 4 --topic poc-test
```

## Step 2

Make sure the topic is there with the expected partitions, you can use this command

```shell
$ docker exec -it kafka kafka-topics --zookeeper localhost:2181 --describe --topic poc-test
```

Or you can navigate to http://localhost:3030/ on your browser and choose **Kafka Topics UI**

## Step 3

Increase the number of file descriptors to accomodate load test

```shell
$ ulimit -n 20000
```

## Step 4

Start a new terminal window, change directory to root of the project.

Run the consumer first

```shell
$ go run redis_pubsub_as_integration_point/*.go consumer
```

#### Optional flags:

- `-broker` kafka broker host, default: localhost
- `-topic` kafka topic name, default to poc-test
- `-cg` consumer group name, default to testCG
- `-redisAddr` redis address, default to localhost:6379
- `-redisChan` redis channel to listen, default to inquiry-response
- `-minD` minimum synthetic delay duration, default to 0s
- `-maxD` maximum synthetic delay duration, default to 0s
- `-async` whether to process each message from kafka asynchronously or not, default to true

## Step 5

Now start another terminal and change directory to the project's root.

Run the http server

```shell
$ go run redis_pubsub_as_integration_point/*.go http
```

#### Optional flags:

- `-broker` kafka broker host, default: localhost
- `-topic` kafka topic name, default to poc-test
- `-redisAddr` redis address, default to localhost:6379
- `-redisChan` redis channel to listen, default to inquiry-response

You can stop the http server using `Ctrl+C`

## Step 6

Now you can start the load test using custom vegeta load test, by using this command:

#### flags

- `-rps` request per second, default: 100
- `-dur` how long the test going to be performed, default to 5s
- `-vars` total num of variance on the request, default to 1000000

```shell
$ go run load_test/http.go -dur=1m -rps=100 | tee result.bin | vegeta report
```

You can NOT use standard [vegetta flags](https://github.com/tsenart/vegeta#usage-manual), but the result (i.e. result.bin) can be inspected by normal vegetta command (e.g. `vegeta report`)

Above we're trying to load test in 1 minutes, rate on 100/s

Using my machine MacBook Pro (15-inch, 2017), CPU 2,9 GHz Intel Core i7, RAM 16 GB 2133 MHz LPDDR3; the result is *very promising*.

![](https://media.giphy.com/media/3o8doOlGO3pjQa5h28/giphy.gif)

The mean latency was **~14 ms** !!!, with max latency at **~800 ms**, success rate is as expected at **100%**

Since I'm very confident with this approach, let's try to introduce synthetic delay and process every message in each goroutine

Restart the consumer

```base
$ go run redis_pubsub_as_integration_point/*.go consumer -minD=3s -maxD=5s
```

Re-run the load test

```shell
$ go run load_test/http.go -dur=1m -rps=100 | tee result.bin | vegeta report
```

![](https://media.giphy.com/media/BkcRn3t709cdi/giphy.gif)

As expected, this turns out to be satisfying, **100%** success rate, mean and max latency are **~4 secs** & **~5 secs** respectively.

Let's try to use all of my machine computing power, try `-rps=500` and `-rps=1000`.

```shell
$ go run load_test/http.go -dur=1m -rps=500 | tee result.bin | vegeta report
```

Result **~88%** success rate:
```shell
Requests      [total, rate]            30000, 500.01
Duration      [total, attack, wait]    1m11.290176547s, 59.9986s, 11.291576547s
Latencies     [mean, 50, 95, 99, max]  12.002787492s, 12.224953731s, 20.625309665s, 20.751974404s, 20.799554682s
Bytes In      [total, mean]            3833947, 127.80
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  88.25%
Status Codes  [code:count]             200:26476  500:3524
```

```shell
$ go run load_test/http.go -dur=1m -rps=1000 | tee result.bin | vegeta report
```

Result **~20%** success rate:
```shell
Requests      [total, rate]            60000, 999.72
Duration      [total, attack, wait]    1m0.016706s, 1m0.016706s, 0s
Latencies     [mean, 50, 95, 99, max]  1.602237798s, 0s, 11.150638556s, 12.78156674s, 28.499800629s
Bytes In      [total, mean]            1714340, 28.57
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  20.12%
Status Codes  [code:count]             0:47918  500:12  200:12070
```

![](https://media.giphy.com/media/lSoncLXrbUaRO/giphy.gif)

With only 1 redis connection being used by the HTTP server, we can maximize throughput & resources, I'm pretty happy with this