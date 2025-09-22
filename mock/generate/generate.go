//go:generate go run -mod=mod github.com/golang/mock/mockgen -package redis -destination ./redis/redis.go  github.com/go-redis/redis/v8 Cmdable
//go:generate go run -mod=mod github.com/golang/mock/mockgen -package kafka -destination ./kafka/kafka_consumer_group.go  github.com/Shopify/sarama ConsumerGroup
//go:generate go run -mod=mod github.com/golang/mock/mockgen -package kafka -destination ./kafka/kafka_consumer.go  github.com/Shopify/sarama Consumer
//go:generate go run -mod=mod github.com/golang/mock/mockgen -package kafka -destination ./kafka/kafka_partition_consumer.go  github.com/Shopify/sarama PartitionConsumer
//go:generate go run -mod=mod github.com/golang/mock/mockgen -package kafka -destination ./kafka/kafka_offset_manager.go  github.com/Shopify/sarama OffsetManager
//go:generate go run -mod=mod github.com/golang/mock/mockgen -package kafka -destination ./kafka/kafka_partition_offset_manager.go  github.com/Shopify/sarama PartitionOffsetManager
package generate
