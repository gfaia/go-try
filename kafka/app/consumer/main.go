package main

import (
	"flag"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	brokers := flag.String("bootstrap.servers", "localhost:9092", "broker addresses")
	consumerGroup := flag.String("group.id", "test", "consumer group")
	topic := flag.String("topic", "topic", "topic")
	flag.Parse()

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": *brokers,
		"group.id":          *consumerGroup,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	c.SubscribeTopics([]string{*topic}, nil)
	defer c.Close()

	for {
		msg, err := c.ReadMessage(-1)
		if err != nil {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}

		fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
	}
}
