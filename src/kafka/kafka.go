package kafka

import (
	"context"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

//Message := exporting kafka.Message
type Message = kafka.Message

//Reader := exporting kafka.Reader
type Reader = *kafka.Reader

//Config :- kafka config info
type Config struct {
	Topic     string
	Partition int
	URL       string
	GroupID   string
	MinBytes  int
	MaxBytes  int
	conn      *kafka.Conn
}

//Init :- init for kafka package
func (c *Config) Init() bool {
	var err error
	c.conn, err = kafka.DialLeader(context.Background(), "tcp", c.URL, c.Topic, c.Partition)
	if err != nil {
		return false
	}
	return true
}

//Writer :- func to write data to kafka
func (c *Config) Write(message string, callback func(bool)) {
	c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, err := c.conn.WriteMessages(kafka.Message{Value: []byte(message)})
	if err != nil {
		callback(false)
		return
	}
	callback(true)
	// conn.Close()
}

//Reader :- read msg from kafka
func (c *Config) Reader(readMessageCallback func(reader *kafka.Reader, m kafka.Message)) {

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{c.URL},
		Topic:     c.Topic,
		Partition: c.Partition,
		MinBytes:  c.MinBytes, // 10KB
		MaxBytes:  c.MaxBytes, // 10MB
		GroupID:   c.GroupID,
	})
	ctx := context.Background()
	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			break
		}
		readMessageCallback(r, m)
	}

}

//Commit :- commit msg to kafka
func Commit(r *kafka.Reader, m kafka.Message) {
	ctx := context.Background()
	r.CommitMessages(ctx, m)
}

// kafkaConfig.Reader(kafkaReaderCallback)
// func kafkaReaderCallback(reader kafka.Reader, message kafka.Message) {
// 	kafka.Commit(reader, message)
// 	util.LogInfo(string(message.Value))
// }
