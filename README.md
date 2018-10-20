This POC project is divided into 3 parts and each demonstrate different approach from the simplest one into more advanced one.

# Prerequisite

## Step 1
Have `kafka` cluster ready, you can easily get it from docker [fast-kafka-dev](https://hub.docker.com/r/landoop/fast-data-dev/)

```bash
$ docker pull landoop/fast-data-dev
```

Run and expose it, if you're on Mac you can

```bash
$ docker run -d --rm --name kafka \
-p 2181:2181 -p 3030:3030 -p 8081-8083:8081-8083 \
-p 9581-9585:9581-9585 -p 9092:9092 \
-e ADV_HOST=127.0.0.1 \
landoop/fast-data-dev
```

Make sure kafka cluster is running properly

```bash
$ docker logs kafka
```

## Step 2

Have `redis` ready, you can also use docker, get it from official [redis](https://hub.docker.com/r/library/redis/)

```bash
$ docker pull redis
```

Run and expose it, if you're on Mac you can

```bash
$ docker run --name redis --rm -p 6379:6379 --ulimit nofile=90000:90000 -d redis
```

Above you may notice that I set ulimit to something high in order to accomodate load test later on.

Now make sure redis is running properly

```bash
$ docker logs redis
```

## Step 3

Install [vegeta](https://github.com/tsenart/vegeta)

If you're on Mac, you can install Vegeta using the Homebrew package manager on Mac OS X:

```bash
$ brew update && brew install vegeta
```

![](https://media.giphy.com/media/26FxCOdhlvEQXbeH6/giphy.gif)

**Finally done !**

# Examples

Now you can follow this link to continue on each approach

* [Simple kafka consumer & producer](src/master/simple_produce_consume/README.md)
* [Blocking HTTP call waiting on redis value by polling](src/master/redis_as_integration_point/README.md)
* [Blocking HTTP call waiting by subscribing to redis](src/master/redis_pubsub_as_integration_point/README.md)