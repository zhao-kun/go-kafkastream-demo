### Build

```
go build zhao-kun/go-kafkastream-demo
```

### Create topic

```
kafka-topics.sh --topic labels --create  \
--replication-factor 1 --partitions 2  \
--config cleanup.policy=compact \
 --bootstrap-server :9092
 ```

### Producer data

```
 bin/kafka-console-producer.sh --bootstrap-server :9092 --topic labels --property "parse.key=true" --property "key.separator=|"
```

The data is bellow, each line is a message with key.

```
1|{"id":1, "name": "usage", "creator": "foo", "last_modifier": "bar"}
2|{"id":2, "name": "prod", "creator": "foo", "last_modifier": "bar"}
3|{"id":3, "name": "vpc", "creator": "foo", "last_modifier": "bar"}
```


### Start service

```
./go-kafkastream-demo
```

It will output:

```
20XX/XX/XX XX:XX:XX [Web] HTTP server is listening on 127.0.0.1:9527
```

### Send requests

There are two API for search, one is query all, another is query by key;

- Query all
  
```bash
curl -s http://127.0.0.1:9527/api/v1/labels\?jq
```

- Query by ID (key)

```bash
curl -s http://127.0.0.1:9527/api/v1/labels/2
```
