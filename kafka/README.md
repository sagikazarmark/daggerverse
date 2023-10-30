# Kafka

**Kafka service module**

## Usage

```go
kafka := dag.Kafka()

service := kafka.Service()

kafka.Container().
    WithServiceBinding("kafka", service).
    WithExec([]string{"kafka-topics.sh", "--list", "--bootstrap-server", "kafka:9092"})
```

> [!IMPORTANT]
> Kafka advertises itself as `kafka:9092`. Make sure to attach the service as `kafka` to containers.

## To Do

- [ ] Custom advertise address
